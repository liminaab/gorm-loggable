package loggable

import (
	"testing"
)

func TestToSnakeCase(t *testing.T) {

	data := []struct {
		In string
		Out string
	} {
		{ In: "hello_there", Out: "HelloThere"},
		{ In: "HelloThere", Out:  "HelloThere"},
	}

	for _, d := range data {
		if ToSnakeCase(d.In) != d.Out {
			t.Errorf("Expected: %s, Got: %s", d.Out, ToSnakeCase(d.In))
		}
	}
}
