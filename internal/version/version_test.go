package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	Version = "1.1.1"
	assert.Equal(t, "1.1.1", GetVersion())
}
