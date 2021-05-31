package fscache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDirEntry(t *testing.T) {
	tmp, err := os.MkdirTemp("", "*")
	require.NoError(t, err)

	entry, err := getDirEntry(tmp)
	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, filepath.Base(tmp), entry.Name())
	assert.True(t, entry.IsDir())

	os.Remove(tmp)
}
