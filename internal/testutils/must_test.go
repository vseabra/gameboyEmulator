package testutils

import (
	"errors"
	"testing"
)

func TestMust_NoError(t *testing.T) {
	Must(t, nil, "This should not fail")
}

func TestMust_WithError(t *testing.T) {
	failed := false
	tt := &testing.T{}
	Must(tt, errors.New("test error"), "Error occurred: %v")

	if tt.Failed() {
		failed = true
	}

	if !failed {
		t.Errorf("Must did not fail when it should")
	}
}
