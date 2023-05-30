package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrettyStructMap(t *testing.T) {
	t.Parallel()

	data := map[string]string{
		"foo": "bar",
		"raz": "baz",
	}

	expected := string("{\n    \"foo\": \"bar\",\n    \"raz\": \"baz\"\n}")
	s, err := PrettyStruct(data)

	assert.Nil(t, err, "unable to PrettyStruct map to string")
	assert.EqualValues(t, expected, s)
}

func TestPrettyStructArray(t *testing.T) {
	t.Parallel()

	data := []int{420, 69, 8, 0, 0, 8, 1, 3, 5}

	expected := string("[\n    420,\n    69,\n    8,\n    0,\n    0,\n    8,\n    1,\n    3,\n    5\n]")
	s, err := PrettyStruct(data)

	assert.Nil(t, err, "unable to PrettyStruct array to string")
	assert.EqualValues(t, expected, s)
}
