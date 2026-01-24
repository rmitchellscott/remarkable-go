package partition

import (
	"context"
	"testing"

	"github.com/rmitchellscott/remarkable-go/device"
	"github.com/rmitchellscott/remarkable-go/executor"
	"github.com/rmitchellscott/remarkable-go/filesystem"
)

func TestManager_GetSystemInfo_RM2(t *testing.T) {
	exec := executor.NewDryRun()
	exec.SetResponse("rootdev", &executor.Result{
		ExitCode: 0,
		Stdout:   "/dev/mmcblk0p3",
	})
	exec.SetResponse("fw_printenv", &executor.Result{
		ExitCode: 0,
		Stdout:   "active_partition=3",
	})

	fs := filesystem.NewMock()
	fs.SetFileString("/usr/share/remarkable/update.conf", "RELEASE_VERSION=3.20.0.92\n")
	fs.SetFileString("/proc/mounts", "")

	mgr := NewManager(exec, fs, device.RM2)
	info, err := mgr.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	if info.Active.Number != 3 {
		t.Errorf("Active.Number = %d, want 3", info.Active.Number)
	}
	if info.Active.Version != "3.20.0.92" {
		t.Errorf("Active.Version = %q, want %q", info.Active.Version, "3.20.0.92")
	}
	if info.Active.Label != "B" {
		t.Errorf("Active.Label = %q, want %q", info.Active.Label, "B")
	}
	if !info.Active.IsActive {
		t.Error("Active.IsActive should be true")
	}
	if !info.Active.IsNextBoot {
		t.Error("Active.IsNextBoot should be true")
	}

	if info.Fallback.Number != 2 {
		t.Errorf("Fallback.Number = %d, want 2", info.Fallback.Number)
	}
	if info.Fallback.Label != "A" {
		t.Errorf("Fallback.Label = %q, want %q", info.Fallback.Label, "A")
	}
	if info.Fallback.IsNextBoot {
		t.Error("Fallback.IsNextBoot should be false")
	}

	if info.DeviceType != device.RM2 {
		t.Errorf("DeviceType = %v, want %v", info.DeviceType, device.RM2)
	}
}

func TestManager_GetSystemInfo_PaperPro(t *testing.T) {
	exec := executor.NewDryRun()
	exec.SetResponse("swupdate", &executor.Result{
		ExitCode: 0,
		Stdout:   "/dev/mmcblk0p2",
	})

	fs := filesystem.NewMock()
	fs.SetFileString("/usr/share/remarkable/update.conf", "RELEASE_VERSION=3.22.1.5\n")
	fs.SetFileString("/proc/mounts", "")
	fs.SetFileString("/sys/bus/mmc/devices/mmc0:0001/boot_part", "1")

	mgr := NewManager(exec, fs, device.RMPP)
	info, err := mgr.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	if info.Active.Number != 2 {
		t.Errorf("Active.Number = %d, want 2", info.Active.Number)
	}
	if info.Active.Label != "A" {
		t.Errorf("Active.Label = %q, want %q", info.Active.Label, "A")
	}
	if info.DeviceType != device.RMPP {
		t.Errorf("DeviceType = %v, want %v", info.DeviceType, device.RMPP)
	}
}

func TestManager_GetSystemInfo_PaperPro_PendingSwu(t *testing.T) {
	exec := executor.NewDryRun()
	exec.SetResponse("swupdate", &executor.Result{
		ExitCode: 0,
		Stdout:   "/dev/mmcblk0p2",
	})

	fs := filesystem.NewMock()
	fs.SetFileString("/usr/share/remarkable/update.conf", "RELEASE_VERSION=3.22.1.5\n")
	fs.SetFileString("/proc/mounts", "")
	fs.SetFileString("/sys/bus/mmc/devices/mmc0:0001/boot_part", "1")
	fs.SetFileString("/sys/devices/platform/lpgpr/swu_status", "1")

	mgr := NewManager(exec, fs, device.RMPP)
	info, err := mgr.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	if info.Active.IsNextBoot {
		t.Error("Active.IsNextBoot should be false when swu_status=1")
	}
	if !info.Fallback.IsNextBoot {
		t.Error("Fallback.IsNextBoot should be true when swu_status=1")
	}
}

func TestManager_SwitchBoot_RM2(t *testing.T) {
	exec := executor.NewDryRun()
	exec.SetResponse("rootdev", &executor.Result{
		ExitCode: 0,
		Stdout:   "/dev/mmcblk0p3",
	})
	exec.SetResponse("fw_printenv", &executor.Result{
		ExitCode: 0,
		Stdout:   "active_partition=3",
	})

	fs := filesystem.NewMock()
	fs.SetFileString("/usr/share/remarkable/update.conf", "RELEASE_VERSION=3.20.0.92\n")
	fs.SetFileString("/proc/mounts", "")

	mgr := NewManager(exec, fs, device.RM2)
	result, err := mgr.SwitchBoot(context.Background(), 2)
	if err != nil {
		t.Fatalf("SwitchBoot failed: %v", err)
	}

	if !result.Success {
		t.Error("Switch should succeed")
	}
	if result.Method != "fw_setenv" {
		t.Errorf("Method = %q, want %q", result.Method, "fw_setenv")
	}
	if result.NewBoot != 2 {
		t.Errorf("NewBoot = %d, want 2", result.NewBoot)
	}

	log := exec.Log()
	foundSetenv := false
	for _, entry := range log {
		if entry == "[DRY RUN] fw_setenv active_partition 2" {
			foundSetenv = true
			break
		}
	}
	if !foundSetenv {
		t.Error("Expected fw_setenv active_partition 2 in log")
	}
}

func TestManager_SwitchBoot_InvalidPartition(t *testing.T) {
	exec := executor.NewDryRun()
	fs := filesystem.NewMock()

	mgr := NewManager(exec, fs, device.RM2)
	_, err := mgr.SwitchBoot(context.Background(), 1)
	if err != ErrInvalidPartition {
		t.Errorf("Expected ErrInvalidPartition, got %v", err)
	}

	_, err = mgr.SwitchBoot(context.Background(), 4)
	if err != ErrInvalidPartition {
		t.Errorf("Expected ErrInvalidPartition, got %v", err)
	}
}

func TestManager_CanSwitchTo_EncryptionBlocked(t *testing.T) {
	mgr := &manager{deviceType: device.RM2}

	info := &SystemInfo{
		Active: Info{
			Number:  3,
			Version: "3.20.0.92",
		},
		Fallback: Info{
			Number:  2,
			Version: "3.15.0",
		},
		DeviceType: device.RM2,
		Encrypted:  true,
	}

	err := mgr.CanSwitchTo(info, 2)
	if err != ErrEncryptionBlocked {
		t.Errorf("Expected ErrEncryptionBlocked, got %v", err)
	}

	err = mgr.CanSwitchTo(info, 3)
	if err != nil {
		t.Errorf("Should allow switching to current partition: %v", err)
	}
}

func TestManager_CanSwitchTo_PaperProAllowed(t *testing.T) {
	mgr := &manager{deviceType: device.RMPP}

	info := &SystemInfo{
		Active: Info{
			Number:  3,
			Version: "3.22.0.1",
		},
		Fallback: Info{
			Number:  2,
			Version: "3.15.0",
		},
		DeviceType: device.RMPP,
		Encrypted:  true,
	}

	err := mgr.CanSwitchTo(info, 2)
	if err != nil {
		t.Errorf("Paper Pro should allow switch even with encryption: %v", err)
	}
}

func TestManager_IsEncryptionEnabled(t *testing.T) {
	exec := executor.NewDryRun()

	fs := filesystem.NewMock()
	fs.SetFileString("/proc/mounts", "/dev/mmcblk0p3 / ext4 rw 0 0\n")

	mgr := NewManager(exec, fs, device.RM2)
	encrypted, _ := mgr.IsEncryptionEnabled(context.Background())
	if encrypted {
		t.Error("Should not detect encryption")
	}

	fs.SetFileString("/proc/mounts", "/dev/mapper/root / ext4 rw 0 0\n")

	encrypted, _ = mgr.IsEncryptionEnabled(context.Background())
	if !encrypted {
		t.Error("Should detect encryption")
	}
}

func TestParsePartitionNumber(t *testing.T) {
	tests := []struct {
		device string
		want   int
		err    bool
	}{
		{"/dev/mmcblk0p2", 2, false},
		{"/dev/mmcblk0p3", 3, false},
		{"/dev/sda1", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.device, func(t *testing.T) {
			got, err := parsePartitionNumber(tt.device)
			if tt.err {
				if err == nil {
					t.Error("Expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parsePartitionNumber(%q) = %d, want %d", tt.device, got, tt.want)
			}
		})
	}
}

func TestPartitionLabel(t *testing.T) {
	if partitionLabel(2) != "A" {
		t.Error("Partition 2 should be A")
	}
	if partitionLabel(3) != "B" {
		t.Error("Partition 3 should be B")
	}
}
