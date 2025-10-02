package pm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateProcessing(t *testing.T) {
	fs := NewMockFileSystem()
	config := DefaultConfig()
	tp := NewTemplateProcessor(fs, config)

	// Test feature template processing
	err := tp.ProcessTemplate("/tmp/test-feature.md", "user-auth", TypeFeature)
	require.NoError(t, err)

	content, err := fs.ReadFile("/tmp/test-feature.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "Feature: user-auth")
	assert.Contains(t, string(content), "## Status: PROPOSED")
}

func TestTemplateProcessingBug(t *testing.T) {
	fs := NewMockFileSystem()
	config := DefaultConfig()
	tp := NewTemplateProcessor(fs, config)

	err := tp.ProcessTemplate("/tmp/test-bug.md", "null-pointer", TypeBug)
	require.NoError(t, err)

	content, err := fs.ReadFile("/tmp/test-bug.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "Bug: null-pointer")
}

func TestTemplateProcessingExperiment(t *testing.T) {
	fs := NewMockFileSystem()
	config := DefaultConfig()
	tp := NewTemplateProcessor(fs, config)

	err := tp.ProcessTemplate("/tmp/test-experiment.md", "ai-assistant", TypeExperiment)
	require.NoError(t, err)

	content, err := fs.ReadFile("/tmp/test-experiment.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "Experiment: ai-assistant")
}

func TestTemplateProcessingInvalidType(t *testing.T) {
	fs := NewMockFileSystem()
	config := DefaultConfig()
	tp := NewTemplateProcessor(fs, config)

	err := tp.ProcessTemplate("/tmp/test-invalid.md", "test", ItemType("invalid"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported item type")
}
