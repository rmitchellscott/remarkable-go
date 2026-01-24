package version

import "testing"

func TestCompare(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   int
	}{
		{"3.20", "3.20", 0},
		{"3.20", "3.21", -1},
		{"3.21", "3.20", 1},
		{"3.20.0", "3.20", 0},
		{"3.20.1", "3.20", 1},
		{"3.20", "3.20.1", -1},
		{"3.22", "3.18", 1},
		{"3.18", "3.22", -1},
		{"3.20.0.92", "3.15.2", 1},
		{"3.15.2", "3.20.0.92", -1},
		{"10.0", "9.0", 1},
		{"1.2.3.4", "1.2.3.4", 0},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			got := Compare(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	v := Version("3.22")

	if !v.GreaterOrEqual(V3_18) {
		t.Error("3.22 should be >= 3.18")
	}
	if !v.GreaterOrEqual(V3_22) {
		t.Error("3.22 should be >= 3.22")
	}
	if v.LessThan(V3_18) {
		t.Error("3.22 should not be < 3.18")
	}
}

func TestVersion_LessThan(t *testing.T) {
	v := Version("3.17")

	if !v.LessThan(V3_18) {
		t.Error("3.17 should be < 3.18")
	}
	if !v.LessThan(V3_22) {
		t.Error("3.17 should be < 3.22")
	}
}
