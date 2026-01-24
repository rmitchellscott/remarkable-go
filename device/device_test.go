package device

import "testing"

func TestType_IsPaperPro(t *testing.T) {
	if !RMPP.IsPaperPro() {
		t.Error("RMPP should be Paper Pro")
	}
	if !RMPPM.IsPaperPro() {
		t.Error("RMPPM should be Paper Pro")
	}
	if RM1.IsPaperPro() {
		t.Error("RM1 should not be Paper Pro")
	}
	if RM2.IsPaperPro() {
		t.Error("RM2 should not be Paper Pro")
	}
}

func TestType_DisplayName(t *testing.T) {
	tests := []struct {
		device Type
		want   string
	}{
		{RM1, "reMarkable 1"},
		{RM2, "reMarkable 2"},
		{RMPP, "reMarkable Paper Pro"},
		{RMPPM, "reMarkable Paper Pro Move"},
		{Unknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.device), func(t *testing.T) {
			if got := tt.device.DisplayName(); got != tt.want {
				t.Errorf("DisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeFromModel(t *testing.T) {
	tests := []struct {
		model string
		want  Type
	}{
		{"Ferrari", RMPP},
		{"Chiappa", RMPPM},
		{"SomethingElse", Unknown},
		{"", Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			if got := TypeFromModel(tt.model); got != tt.want {
				t.Errorf("TypeFromModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestIsPaperProModel(t *testing.T) {
	if !IsPaperProModel("Ferrari") {
		t.Error("Ferrari should be a Paper Pro model")
	}
	if !IsPaperProModel("Chiappa") {
		t.Error("Chiappa should be a Paper Pro model")
	}
	if IsPaperProModel("Unknown") {
		t.Error("Unknown should not be a Paper Pro model")
	}
}
