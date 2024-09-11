package wirebson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeScalarField(t *testing.T) {
	buf := make([]byte, 13)
	encodeScalarField(0, buf, "foo", "bar")

	expected := []byte{0x02, 0x66, 0x6f, 0x6f, 0x0, 0x4, 0x0, 0x0, 0x0, 0x62, 0x61, 0x72, 0x0}
	actual := buf

	assert.Equal(t, expected, actual)
}
