package mail

import (
	"testing"
)

func TestRandomIdAscii(t *testing.T) {
	t.Error(randomIdAscii(8))
}
