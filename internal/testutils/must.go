package testutils

import (
	"testing"
)

func Must(t *testing.T, err error, msg string) {
	if err != nil {
		t.Errorf(msg, err)
	}
}
