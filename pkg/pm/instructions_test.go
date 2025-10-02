package pm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInstructions(t *testing.T) {
	config := DefaultConfig()
	instructions := GetInstructions(config)
	assert.NotEmpty(t, instructions)
	assert.Contains(t, instructions, "Project Management")
}
