package workflows

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposer_Compose_CorrectOutputFromGeneratedFunc(t *testing.T) {
	// Arrange
	w := NewComposer[int, string]()
	itoa := Fn(w, Itoa).Add()
	Fn(w, AddBar).Param(itoa).Add()
	fn, err := w.Compose()
	require.NoError(t, err)

	// Act
	out, err := fn(1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "1bar", out)
}

func Itoa(i int) (string, error) {
	return strconv.Itoa(i), nil
}

func AddBar(in string) (string, error) {
	return in + "bar", nil
}

func SomeError(in string) (string, error) {
	return "", assert.AnError
}
