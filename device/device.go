package device

type Type string

const (
	RM1     Type = "rm1"
	RM2     Type = "rm2"
	RMPP    Type = "rmpp"
	RMPPM   Type = "rmppm"
	Unknown Type = "unknown"
)

type Architecture string

const (
	Arm32   Architecture = "arm32"
	Aarch64 Architecture = "aarch64"
)

func (t Type) IsPaperPro() bool {
	return t == RMPP || t == RMPPM
}

func (t Type) DisplayName() string {
	switch t {
	case RM1:
		return "reMarkable 1"
	case RM2:
		return "reMarkable 2"
	case RMPP:
		return "reMarkable Paper Pro"
	case RMPPM:
		return "reMarkable Paper Pro Move"
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
}

func IsPaperProModel(model string) bool {
	for _, m := range paperProModels {
		if model == m {
			return true
		}
	}
	return false
}

func TypeFromModel(model string) Type {
	switch model {
	case "Ferrari":
		return RMPP
	case "Chiappa":
		return RMPPM
	default:
		return Unknown
	}
}
