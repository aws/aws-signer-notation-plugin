// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
// http://aws.amazon.com/apache2.0
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

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
