// Package partition reads and switches the A/B boot partitions on reMarkable devices.
package partition

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/rmitchellscott/remarkable-go/device"
	"github.com/rmitchellscott/remarkable-go/executor"
	"github.com/rmitchellscott/remarkable-go/filesystem"
	"github.com/rmitchellscott/remarkable-go/version"
)

// Errors returned by Manager operations.
var (
	ErrEncryptionBlocked  = errors.New("cannot switch to pre-3.18 firmware with encryption enabled")
	ErrInvalidPartition   = errors.New("invalid partition number (must be 2 or 3)")
	ErrDeviceNotSupported = errors.New("device type not supported")
	ErrVersionDetection   = errors.New("failed to detect OS version")
	// ErrSlotUnhealthy means the target slot is positively bad (must not be booted).
	ErrSlotUnhealthy = errors.New("target partition failed integrity check")
	// ErrSlotCheckUnavailable means the health check could not be run.
	ErrSlotCheckUnavailable = errors.New("could not verify target partition health")
)

// SwitchOption configures a SwitchBoot call.
type SwitchOption func(*switchConfig)

type switchConfig struct {
	force bool
}

// WithForce skips the target-slot health check (use only for recovery to a
// slot known to be good).
func WithForce() SwitchOption {
	return func(c *switchConfig) { c.force = true }
}

var mountMu sync.Mutex

// Info describes one partition slot.
type Info struct {
	Number     int    `json:"number"`
	Version    string `json:"version"`
	Label      string `json:"label"`
	IsActive   bool   `json:"isActive"`
	IsNextBoot bool   `json:"isNextBoot"`
}

// SystemInfo describes both partition slots and the device's overall boot state.
type SystemInfo struct {
	Active        Info        `json:"active"`
	Fallback      Info        `json:"fallback"`
	DeviceType    device.Type `json:"deviceType"`
	Encrypted     bool        `json:"encrypted"`
	UpdatePending bool        `json:"updatePending"`
}

// SwitchResult describes the outcome of a boot-partition switch.
type SwitchResult struct {
	Success      bool   `json:"success"`
	Method       string `json:"method"`
	PreviousBoot int    `json:"previousBoot"`
	NewBoot      int    `json:"newBoot"`
	Message      string `json:"message"`
}

// Manager inspects and controls a device's A/B boot partitions.
type Manager interface {
	// GetSystemInfo returns the active and fallback slots and boot state. A slot
	// whose version can't be read is reported with version "unknown" rather than
	// failing the call.
	GetSystemInfo(ctx context.Context) (*SystemInfo, error)
	// SwitchBoot sets partition (2 or 3) as the next boot target. Unless
	// WithForce is passed, switching to the non-running slot first verifies it
	// with VerifySlotHealth and refuses if the slot is unhealthy.
	SwitchBoot(ctx context.Context, partition int, opts ...SwitchOption) (*SwitchResult, error)
	// VerifySlotHealth checks that partition (2 or 3) holds a healthy, complete
	// OS. Returns nil if healthy, ErrSlotUnhealthy if positively bad, or
	// ErrSlotCheckUnavailable if the check could not be run.
	VerifySlotHealth(ctx context.Context, partition int) error
	// IsUpdateInProgress reports whether an OS install is actively writing the
	// inactive slot, regardless of which updater is doing it. It reads only
	// device state, so it stays correct across a controller restart.
	IsUpdateInProgress(ctx context.Context) (bool, error)
	// GetPartitionVersion returns the OS version of the running system.
	GetPartitionVersion(ctx context.Context, partition int) (string, error)
	// IsEncryptionEnabled reports whether the device uses encrypted data partitions.
	IsEncryptionEnabled(ctx context.Context) (bool, error)
	// CanSwitchTo returns nil if switching to targetPartition is allowed, or an
	// error describing why it is not.
	CanSwitchTo(info *SystemInfo, targetPartition int) error
	// Reboot reboots the device.
	Reboot(ctx context.Context) error
}

type manager struct {
	exec        executor.Executor
	fs          filesystem.FS
	mountableFS filesystem.MountableFS
	deviceType  device.Type
}

// NewManager returns a Manager for the given device type, using exec for
// commands and fs for file access. If fs also implements filesystem.MountableFS,
// the fallback partition's version is read by mounting it.
func NewManager(exec executor.Executor, fs filesystem.FS, deviceType device.Type) Manager {
	var mountableFS filesystem.MountableFS
	if mfs, ok := fs.(filesystem.MountableFS); ok {
		mountableFS = mfs
	}

	return &manager{
		exec:        exec,
		fs:          fs,
		mountableFS: mountableFS,
		deviceType:  deviceType,
	}
}

func (m *manager) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	var runningP, otherP, bootP int
	var err error

	if m.deviceType.IsPaperPro() {
		runningP, otherP, bootP, err = m.getPaperProPartitionInfo(ctx)
	} else {
		runningP, otherP, bootP, err = m.getRM12PartitionInfo(ctx)
	}
	if err != nil {
		return nil, err
	}

	activeVersion, err := m.GetPartitionVersion(ctx, runningP)
	if err != nil {
		activeVersion = "unknown"
	}

	fallbackVersion, err := m.getVersionFromMountedPartition(ctx, otherP)
	if err != nil {
		fallbackVersion = "unknown"
	}

	encrypted, _ := m.IsEncryptionEnabled(ctx)

	var updatePending bool
	if m.deviceType.IsPaperPro() {
		if data, err := m.fs.ReadFile("/sys/devices/platform/lpgpr/swu_status"); err == nil {
			updatePending = strings.TrimSpace(string(data)) == "1"
		}
	}

	return &SystemInfo{
		Active: Info{
			Number:     runningP,
			Version:    activeVersion,
			Label:      partitionLabel(runningP),
			IsActive:   true,
			IsNextBoot: bootP == runningP,
		},
		Fallback: Info{
			Number:     otherP,
			Version:    fallbackVersion,
			Label:      partitionLabel(otherP),
			IsActive:   false,
			IsNextBoot: bootP == otherP,
		},
		DeviceType:    m.deviceType,
		Encrypted:     encrypted,
		UpdatePending: updatePending,
	}, nil
}

func (m *manager) SwitchBoot(ctx context.Context, partition int, opts ...SwitchOption) (*SwitchResult, error) {
	if partition != 2 && partition != 3 {
		return nil, ErrInvalidPartition
	}

	var cfg switchConfig
	for _, o := range opts {
		o(&cfg)
	}

	info, err := m.GetSystemInfo(ctx)
	if err != nil {
		return nil, err
	}

	if err := m.CanSwitchTo(info, partition); err != nil {
		return nil, err
	}

	// Verify the target before switching to it. The running slot is always safe,
	// so only the non-running slot is checked; WithForce bypasses for recovery.
	if !cfg.force && partition != info.Active.Number {
		if err := m.VerifySlotHealth(ctx, partition); err != nil {
			return nil, err
		}
	}

	currentBoot := info.Active.Number
	if info.Fallback.IsNextBoot {
		currentBoot = info.Fallback.Number
	}

	var result *SwitchResult
	if m.deviceType.IsPaperPro() {
		targetVersion := info.Fallback.Version
		if partition == info.Active.Number {
			targetVersion = info.Active.Version
		}
		result, err = m.switchPaperProBoot(ctx, partition, targetVersion)
	} else {
		result, err = m.switchRM12Boot(ctx, partition, currentBoot)
	}

	if err != nil {
		return nil, err
	}

	result.PreviousBoot = currentBoot
	result.NewBoot = partition
	return result, nil
}

func (m *manager) GetPartitionVersion(ctx context.Context, partition int) (string, error) {
	v, err := version.ReadFromFS(m.fs, "")
	if err != nil {
		return "", ErrVersionDetection
	}
	return v, nil
}

func (m *manager) IsEncryptionEnabled(ctx context.Context) (bool, error) {
	data, err := m.fs.ReadFile("/proc/mounts")
	if err != nil {
		return false, err
	}
	return strings.Contains(string(data), "/dev/mapper/"), nil
}

func (m *manager) CanSwitchTo(info *SystemInfo, targetPartition int) error {
	var targetVersion string
	if targetPartition == info.Active.Number {
		targetVersion = info.Active.Version
	} else {
		targetVersion = info.Fallback.Version
	}

	if info.Encrypted && !info.DeviceType.IsPaperPro() {
		if version.Compare(targetVersion, "3.18") < 0 {
			return ErrEncryptionBlocked
		}
	}

	return nil
}

func (m *manager) Reboot(ctx context.Context) error {
	_, err := m.exec.Run(ctx, "reboot")
	return err
}

func (m *manager) VerifySlotHealth(ctx context.Context, partition int) error {
	if partition != 2 && partition != 3 {
		return ErrInvalidPartition
	}

	base, err := m.rootBaseDevice(ctx)
	if err != nil {
		return fmt.Errorf("%w: could not determine slot device: %v", ErrSlotCheckUnavailable, err)
	}
	dev := fmt.Sprintf("%sp%d", base, partition)

	if err := m.fsckSlot(ctx, dev); err != nil {
		return err
	}
	return m.verifySlotContents(partition, dev)
}

func (m *manager) IsUpdateInProgress(ctx context.Context) (bool, error) {
	res, err := m.exec.Run(ctx, "rootdev")
	if err != nil {
		return false, err
	}
	if !res.Success() {
		return false, fmt.Errorf("rootdev failed (exit %d)", res.ExitCode)
	}
	running, err := parsePartitionNumber(strings.TrimSpace(res.Stdout))
	if err != nil {
		return false, err
	}
	other := 3
	if running == 3 {
		other = 2
	}
	base := regexp.MustCompile(`p\d+$`).ReplaceAllString(strings.TrimSpace(res.Stdout), "")
	// fuser exits 0 if a process holds the device open, 1 if none.
	fres, err := m.exec.Run(ctx, "fuser", fmt.Sprintf("%sp%d", base, other))
	if err != nil {
		return false, err
	}
	return fres.Success(), nil
}

func (m *manager) rootBaseDevice(ctx context.Context) (string, error) {
	result, err := m.exec.Run(ctx, "rootdev")
	if err != nil {
		return "", err
	}
	if !result.Success() {
		return "", fmt.Errorf("rootdev failed (exit %d): %s", result.ExitCode, strings.TrimSpace(result.Stderr))
	}
	return regexp.MustCompile(`p\d+$`).ReplaceAllString(strings.TrimSpace(result.Stdout), ""), nil
}

// fsckSlot runs a read-only forced check. e2fsck exit is a bitmask: 0 = clean;
// bits 1/4 = errors; 127 = missing; other = operational failure.
func (m *manager) fsckSlot(ctx context.Context, dev string) error {
	res, err := m.exec.Run(ctx, "e2fsck", "-fn", dev)
	if err != nil {
		return fmt.Errorf("%w: e2fsck could not run on %s: %v", ErrSlotCheckUnavailable, dev, err)
	}
	switch {
	case res.ExitCode == 0:
		return nil
	case res.ExitCode == 127:
		return fmt.Errorf("%w: e2fsck not available on device", ErrSlotCheckUnavailable)
	case res.ExitCode&4 != 0 || res.ExitCode&1 != 0:
		return fmt.Errorf("%w: %s has filesystem errors (e2fsck exit %d)", ErrSlotUnhealthy, dev, res.ExitCode)
	default:
		return fmt.Errorf("%w: e2fsck on %s exited %d: %s", ErrSlotCheckUnavailable, dev, res.ExitCode, strings.TrimSpace(res.Stderr))
	}
}

// verifySlotContents RO-mounts the slot and confirms it holds a readable OS
// version, distinguishing an installed rootfs from a clean-but-empty filesystem.
func (m *manager) verifySlotContents(partition int, dev string) error {
	if m.mountableFS == nil {
		return nil // cannot mount; fsck already validated structure
	}

	mountPoint := fmt.Sprintf("/tmp/mount_health_p%d", partition)

	mountMu.Lock()
	defer mountMu.Unlock()

	if err := m.fs.MkdirAll(mountPoint, 0755); err != nil {
		return fmt.Errorf("%w: %v", ErrSlotCheckUnavailable, err)
	}
	defer func() {
		m.mountableFS.Unmount(mountPoint)
		m.fs.RemoveAll(mountPoint)
	}()

	if err := m.mountableFS.Mount(dev, mountPoint, true); err != nil {
		// fsck passed but the slot won't mount — treat as bad rather than just unverifiable.
		return fmt.Errorf("%w: cannot mount %s: %v", ErrSlotUnhealthy, dev, err)
	}

	if v, err := version.ReadFromFS(m.fs, mountPoint); err != nil || strings.TrimSpace(v) == "" {
		return fmt.Errorf("%w: OS version unreadable on %s", ErrSlotUnhealthy, dev)
	}
	return nil
}

func (m *manager) getRM12PartitionInfo(ctx context.Context) (running, other, boot int, err error) {
	result, err := m.exec.Run(ctx, "rootdev")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get root device: %w", err)
	}
	if !result.Success() {
		return 0, 0, 0, fmt.Errorf("rootdev failed (exit %d): %s", result.ExitCode, strings.TrimSpace(result.Stderr))
	}

	running, err = parsePartitionNumber(strings.TrimSpace(result.Stdout))
	if err != nil {
		return 0, 0, 0, err
	}

	other = 3
	if running == 3 {
		other = 2
	}

	boot = running
	result, err = m.exec.Run(ctx, "fw_printenv", "active_partition")
	if err == nil {
		parts := strings.Split(strings.TrimSpace(result.Stdout), "=")
		if len(parts) == 2 {
			if bp, e := strconv.Atoi(parts[1]); e == nil {
				boot = bp
			}
		}
	}

	return running, other, boot, nil
}

func (m *manager) getPaperProPartitionInfo(ctx context.Context) (running, other, boot int, err error) {
	result, err := m.exec.Run(ctx, "swupdate", "-g")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get active partition: %w", err)
	}
	if !result.Success() {
		return 0, 0, 0, fmt.Errorf("swupdate -g failed (exit %d): %s", result.ExitCode, strings.TrimSpace(result.Stderr))
	}

	running, err = parsePartitionNumber(strings.TrimSpace(result.Stdout))
	if err != nil {
		return 0, 0, 0, err
	}

	other = 3
	if running == 3 {
		other = 2
	}

	currentVersion, _ := m.GetPartitionVersion(ctx, running)
	boot, err = m.getPaperProNextBoot(currentVersion)
	if err != nil {
		boot = running
	}

	if boot == running {
		if data, err := m.fs.ReadFile("/sys/devices/platform/lpgpr/swu_status"); err == nil {
			if strings.TrimSpace(string(data)) == "1" {
				boot = other
			}
		}
	}

	return running, other, boot, nil
}

func (m *manager) getPaperProNextBoot(currentVersion string) (int, error) {
	if version.Compare(currentVersion, "3.22") >= 0 {
		data, err := m.fs.ReadFile("/sys/bus/mmc/devices/mmc0:0001/boot_part")
		if err == nil {
			bootPart := strings.TrimSpace(string(data))
			if bootPart == "1" {
				return 2, nil
			} else if bootPart == "2" {
				return 3, nil
			}
		}
	}

	data, err := m.fs.ReadFile("/sys/devices/platform/lpgpr/root_part")
	if err != nil {
		return 0, err
	}

	rootPart := strings.TrimSpace(string(data))
	if rootPart == "a" {
		return 2, nil
	} else if rootPart == "b" {
		return 3, nil
	}

	return 0, fmt.Errorf("unexpected root_part value: %s", rootPart)
}

func (m *manager) switchRM12Boot(ctx context.Context, newPart, oldPart int) (*SwitchResult, error) {
	commands := [][]string{
		{"fw_setenv", "upgrade_available", "1"},
		{"fw_setenv", "bootcount", "0"},
		{"fw_setenv", "fallback_partition", strconv.Itoa(oldPart)},
		{"fw_setenv", "active_partition", strconv.Itoa(newPart)},
	}

	for _, cmd := range commands {
		result, err := m.exec.Run(ctx, cmd[0], cmd[1:]...)
		if err != nil {
			return nil, fmt.Errorf("failed to run %v: %w", cmd, err)
		}
		if !result.Success() {
			return nil, fmt.Errorf("command %v failed with exit code %d", cmd, result.ExitCode)
		}
	}

	return &SwitchResult{
		Success: true,
		Method:  "fw_setenv",
		Message: fmt.Sprintf("Successfully set next boot to partition %d", newPart),
	}, nil
}

func (m *manager) switchPaperProBoot(ctx context.Context, newPart int, targetVersion string) (*SwitchResult, error) {
	currentVersion, _ := m.GetPartitionVersion(ctx, 0)

	currentIsNew := version.Compare(currentVersion, "3.22") >= 0
	targetIsNew := version.Compare(targetVersion, "3.22") >= 0

	var newPartLabel string
	if newPart == 2 {
		newPartLabel = "a"
	} else {
		newPartLabel = "b"
	}

	errCntPath := fmt.Sprintf("/sys/devices/platform/lpgpr/root%s_errcnt", newPartLabel)
	_ = m.fs.WriteFile(errCntPath, []byte("0"), 0644)

	method := ""

	if !currentIsNew {
		if err := m.fs.WriteFile("/sys/devices/platform/lpgpr/root_part", []byte(newPartLabel), 0644); err != nil {
			return nil, fmt.Errorf("failed to set boot partition via sysfs: %w", err)
		}
		method = "sysfs"
	}

	if targetIsNew || currentIsNew {
		var mmcArgs []string
		if newPart == 2 {
			mmcArgs = []string{"bootpart", "enable", "1", "0", "/dev/mmcblk0boot0"}
		} else {
			mmcArgs = []string{"bootpart", "enable", "2", "0", "/dev/mmcblk0boot1"}
		}

		result, err := m.exec.Run(ctx, "mmc", mmcArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to run mmc bootpart enable: %w", err)
		}
		if !result.Success() {
			return nil, fmt.Errorf("mmc bootpart failed with exit code %d", result.ExitCode)
		}

		if method == "sysfs" {
			method = "sysfs+mmc"
		} else {
			method = "mmc"
		}
	}

	return &SwitchResult{
		Success: true,
		Method:  method,
		Message: fmt.Sprintf("Successfully set Paper Pro next boot to partition %d", newPart),
	}, nil
}

func (m *manager) getVersionFromMountedPartition(ctx context.Context, partNum int) (string, error) {
	if m.mountableFS == nil {
		return "unknown", nil
	}

	result, err := m.exec.Run(ctx, "rootdev")
	if err != nil {
		return "", err
	}

	baseDev := regexp.MustCompile(`p\d+$`).ReplaceAllString(strings.TrimSpace(result.Stdout), "")
	mountPoint := fmt.Sprintf("/tmp/mount_p%d", partNum)

	mountMu.Lock()
	defer mountMu.Unlock()

	if err := m.fs.MkdirAll(mountPoint, 0755); err != nil {
		return "", err
	}

	defer func() {
		m.mountableFS.Unmount(mountPoint)
		m.fs.RemoveAll(mountPoint)
	}()

	device := fmt.Sprintf("%sp%d", baseDev, partNum)
	if err := m.mountableFS.Mount(device, mountPoint, true); err != nil {
		return "", err
	}

	v, err := version.ReadFromFS(m.fs, mountPoint)
	if err != nil {
		return "unknown", nil
	}
	return v, nil
}

func parsePartitionNumber(device string) (int, error) {
	re := regexp.MustCompile(`p(\d+)$`)
	matches := re.FindStringSubmatch(device)
	if len(matches) < 2 {
		return 0, fmt.Errorf("could not parse partition number from %s", device)
	}
	return strconv.Atoi(matches[1])
}

func partitionLabel(partNum int) string {
	if partNum == 2 {
		return "A"
	}
	return "B"
}
