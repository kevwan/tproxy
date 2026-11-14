package protocol

import (
	"bytes"
	"testing"
)

func TestReadBson2Json(t *testing.T) {
	// A wire-format BSON document containing a string, int32, and boolean.
	doc := []byte{
		42, 0, 0, 0,
		2, 'n', 'a', 'm', 'e', 0, 7, 0, 0, 0, 't', 'p', 'r', 'o', 'x', 'y', 0,
		16, 'c', 'o', 'u', 'n', 't', 0, 42, 0, 0, 0,
		8, 'a', 'c', 't', 'i', 'v', 'e', 0, 1,
		0,
	}
	want := `{"active":true,"count":42,"name":"tproxy"}`
	if got := readBson2Json(bytes.NewReader(doc)); got != want {
		t.Fatalf("readBson2Json() = %s, want %s", got, want)
	}
}
