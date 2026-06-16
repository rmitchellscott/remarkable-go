// Package device identifies reMarkable hardware models and CPU architectures.
package device

// Type is a reMarkable hardware model.
type Type string

const (
	RM1      Type = "rm1"
	RM2      Type = "rm2"
	RMPP     Type = "rmpp"
	RMPPMove Type = "rmppmove"
	RMPPure  Type = "rmppure"
	Unknown  Type = "unknown"
)

// Architecture is a device CPU architecture.
type Architecture string

const (
	Arm32   Architecture = "arm32"
	Aarch64 Architecture = "aarch64"
)

// IsPaperPro reports whether the model is a Paper Pro variant.
func (t Type) IsPaperPro() bool {
	return t == RMPP || t == RMPPMove || t == RMPPure
}

// DisplayName returns the human-facing product name.
func (t Type) DisplayName() string {
	switch t {
	case RM1:
		return "reMarkable 1"
	case RM2:
		return "reMarkable 2"
	case RMPP:
		return "Paper Pro"
	case RMPPMove:
		return "Paper Pro Move"
	case RMPPure:
		return "Paper Pure"
	default:
		return string(t)
	}
}

// CodeName returns the internal SoC codename (e.g. "ferrari" for RMPP).
func (t Type) CodeName() string {
	switch t {
	case RM1:
		return "rm1"
	case RM2:
		return "rm2"
	case RMPP:
		return "ferrari"
	case RMPPMove:
		return "chiappa"
	case RMPPure:
		return "tatsu"
	default:
		return string(t)
	}
}

func (t Type) String() string {
	return string(t)
}

var paperProModels = []string{
	"Ferrari",
	"Chiappa",
	"Tatsu",
}

// IsPaperProModel reports whether a raw /sys model string names a Paper Pro device.
func IsPaperProModel(model string) bool {
	for _, m := range paperProModels {
		if model == m {
			return true
		}
	}
	return false
}

// TypeFromModel maps a raw model string to a Type, returning Unknown if unrecognized.
func TypeFromModel(model string) Type {
	switch model {
	case "Ferrari":
		return RMPP
	case "Chiappa":
		return RMPPMove
	case "Tatsu":
		return RMPPure
	default:
		return Unknown
	}
}
