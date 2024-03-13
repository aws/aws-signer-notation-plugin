package client

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-signer-notation-plugin/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewAWSSigner(t *testing.T) {
	tests := map[string]map[string]string{
		"emptyConfig":           {},
		configKeySignerEndpoint: {configKeySignerEndpoint: "https://127.0.0.1:80/some-endpoint"},
		configKeyAwsRegion:      {configKeyAwsRegion: "us-east-1"},
	}
	for name, config := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewAWSSigner(context.TODO(), config)
			assert.Nil(t, err, "NewAWSSigner returned error")
		})
	}
}

func TestNewAWSSigner_Debug(t *testing.T) {
	// we need this because build fleet might not have XDG_CONFIG_HOME set
	tempDir, _ := os.MkdirTemp("", "tempDir")
	defer os.RemoveAll(tempDir)
	t.Setenv("XDG_CONFIG_HOME", os.TempDir())

	ctx := context.TODO()
	dl, _ := logger.New()
	ctx = dl.UpdateContext(ctx)
	_, err := NewAWSSigner(ctx, map[string]string{})
	assert.Nil(t, err, "NewAWSSigner returned error")
}

func TestNewAWSSigner_InvalidProfile(t *testing.T) {
	// we need this because build fleet might not have XDG_CONFIG_HOME set
	tempDir, _ := os.MkdirTemp("", "tempDir")
	defer os.RemoveAll(tempDir)
	t.Setenv("XDG_CONFIG_HOME", os.TempDir())

	ctx := context.TODO()
	dl, _ := logger.New()
	ctx = dl.UpdateContext(ctx)
	_, err := NewAWSSigner(ctx, map[string]string{configKeyAwsProfile: "someProfile"})
	assert.Error(t, err, "NewAWSSigner returned error")
}
