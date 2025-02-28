package setlist

import (
	"reflect"
	"testing"
)

func TestParseNicknameMapping(t *testing.T) {
	tt := []struct {
		name     string
		mapping  string
		expected map[string]string
	}{
		{
			"knowngood",
			"01234=foo,9876=bar",
			map[string]string{
				"01234": "foo",
				"9876":  "bar",
			},
		},
		{
			"empty mapping",
			"",
			map[string]string{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, _ := ParseNicknameMapping(tc.mapping)

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("unexpected output: got %v, want %v", actual, tc.expected)
			}
		})
	}
}
