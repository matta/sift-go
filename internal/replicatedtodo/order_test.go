package replicatedtodo_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/matta/sift/internal/replicatedtodo"
)

func TestMidString(t *testing.T) {
	type testCase struct {
		left, right, expected string
	}
	cases := []testCase{
		{"", "", "n"},
		{"b", "", "o"},
		{"", "b", "an"},
		{"n", "", "u"},
		{"", "n", "g"},
		{"z", "", "zn"},
		{"", "z", "m"},
		{"kilroy", "", "s"},
		{"b", "d", "c"},
		{"b", "z", "n"},
		{"bc", "c", "bo"},
		{"ab", "b", "ao"},
		{"", "ab", "aan"},
		{"az", "b", "azn"},
	}

	for _, c := range cases {
		result, err := replicatedtodo.MidString([]byte(c.left), []byte(c.right))
		if err != nil {
			t.Errorf("Unexpected error for MidString(%q, %q); %s", c.left, c.right, err.Error())
		} else if !bytes.Equal(result, []byte(c.expected)) {
			t.Errorf("Unexpected result for MidString(%q, %q) -> %q expected %q",
				c.left, c.right, result, c.expected)
		}
	}
}

func TestMidStringErrors(t *testing.T) {
	type testCase struct {
		left, right string
	}
	cases := []testCase{
		{"", "\x00"},   // invalid character in right
		{"a\xff", "b"}, // invalid character in left
		{"a", "b"},     // invalid character in left
		{"b", "b"},     // equal strings
		{"b", "baa"},
		{"c", "ca0"},
		{"n", "m"},
		{"hb", "h"}, // left is greater than right
		{"", "a"},   // invalid ending characdter in right
		{"a", ""},   // invalid ending characdter in right
	}
	for _, c := range cases {
		res, err := replicatedtodo.MidString([]byte(c.left), []byte(c.right))
		if err == nil {
			t.Fatalf("Failed to return an error for input %q %q -> %q", c.left, c.right, res)
		}
	}
}

func FuzzMidString(f *testing.F) {
	type FuzzTestCase struct {
		Left, Right []byte
	}
	f.Fuzz(func(t *testing.T, left []byte, right []byte) {
		result, err := replicatedtodo.MidString(left, right)
		if err == nil {
			problem := ""
			switch {
			case !replicatedtodo.OrderString(result).Valid():
				problem = "invalid result string"
			case bytes.Compare(left, result) >= 0:
				problem = "result is not greater than the left string"
			case len(right) > 0 && (bytes.Compare(result, right) >= 0):
				problem = fmt.Sprintf(
					"result is greater than or equal to the right string: bytes.Compare(%q, %q) -> %d",
					result, right, bytes.Compare(result, right))
			}
			if problem != "" {
				t.Fatalf("Unexpected success for MidString(%q, %q) -> %q; %s",
					left, right, result, problem)
			}
		} else if bytes.Compare(left, right) < 0 &&
			replicatedtodo.OrderString(left).Valid() &&
			replicatedtodo.OrderString(right).Valid() {
			t.Fatalf("Unexpected error for MidString(%q, %q); %s", left, right, err.Error())
		}
	})
}
