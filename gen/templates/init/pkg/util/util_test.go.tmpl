package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDerefString(t *testing.T) {
	expected := "hello"
	actual := DerefString(&expected)
	if expected != actual {
		t.Errorf("Not equal %s %s", expected, actual)
	}

	expected = ""
	var str *string
	actual = DerefString(str)
	if expected != actual {
		t.Errorf("Nil pointer did not produce empty string")
	}
}

type testType struct{}

func TestTypeOf(t *testing.T) {
	typ := TypeOf(testType{})

	assert.Equal(t, "testType", typ)
}

func TestTypeOfPointer(t *testing.T) {
	typ := TypeOf(&testType{})

	assert.Equal(t, "testType", typ)
}

func TestEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			"string",
			"",
			true,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			if Empty(c.input) != c.expected {
				t.Errorf("For test %s, for input %s, got %v, expected %v", c.name, c.input, Empty(c.input), c.expected)
			}
		})
	}
}
