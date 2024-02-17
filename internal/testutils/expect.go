package testutils

import (
	"reflect"
	"testing"
)

type expectation struct {
	t     *testing.T
	value interface{}
	title string
}

func (e *expectation) ToEqual(value interface{}, message ...string) {
	if !reflect.DeepEqual(e.value, value) {
		// handle formatting for different types
		formatEValue, formatValue := "%+v", "%+v"

		switch e.value.(type) {
		case int8, int16, uint8, uint16:
			formatEValue = "0x%X"
		}

		switch value.(type) {
		case int8, int16, uint8, uint16:
			formatValue = "0x%X"
		}

		e.t.Errorf("%v: expected "+formatValue+", but got "+formatEValue, e.title, value, e.value)
	}
}

func Expect(t *testing.T, value interface{}, title ...string) *expectation {
	if len(title) == 0 {
		title = append(title, "Expectation")
	}
	return &expectation{t: t, value: value, title: title[0]}

}
