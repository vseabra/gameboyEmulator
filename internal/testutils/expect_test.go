package testutils

import (
	"testing"
)

func TestExpect(t *testing.T) {
	e := Expect(t, 42)
	if e.value != 42 {
		t.Errorf("Expect did not initialize correctly, expected 42, got %v", e.value)
	}
}

func TestToEqual_Success(t *testing.T) {
	e := Expect(t, 42)
	e.ToEqual(42)
}

func TestToEqual_Failure(t *testing.T) {
	failed := false
	tt := &testing.T{}
	e := Expect(tt, 42)
	e.ToEqual(100)

	if tt.Failed() {
		failed = true
	}

	if !failed {
		t.Errorf("ToEqual did not fail when it should")
	}
}
