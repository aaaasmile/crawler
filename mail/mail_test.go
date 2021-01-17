package mail

import (
	"testing"
)

func TestRandomIdAscii(t *testing.T) {
	res := randomIdAscii(8)
	if len(res) != 8 {
		t.Error("Result is not 8 size", res)
	}
}
