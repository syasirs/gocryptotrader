package gctwrapper

import (
	"reflect"
	"testing"
)

func TestSetup(t *testing.T) {
	x := Setup()
	xType := reflect.TypeOf(x).String()
	if xType != "*gctwrapper.Wrapper" {
		t.Fatalf("Setup() should return pointer to Wrapper instead received: %v", x)
	}
}
