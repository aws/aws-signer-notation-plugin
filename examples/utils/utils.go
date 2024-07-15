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

package utils

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2/registry"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	awsplugin "github.com/aws/aws-signer-notation-plugin/plugin"
)

func ParseReference(reference string) (registry.Reference, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return registry.Reference{}, fmt.Errorf("%q: %w. Expecting <registry>/<repository>@<digest>", reference, err)
	}
	if ref.Reference == "" {
		return registry.Reference{}, fmt.Errorf("%q: invalid reference. Expecting <registry>/<repository>@<digest>", reference)
	}
	if err := ref.ValidateReferenceAsDigest(); err != nil {
		return registry.Reference{}, fmt.Errorf("%q: tag resolution not supported. Expecting <registry>/<repository>@<digest>", reference)

	}
	return ref, nil
}

// GetAWSSignerPlugin returns the AWS Signer's Notation plugin
func GetAWSSignerPlugin(ctx context.Context, region string) (*awsplugin.AWSSignerPlugin, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return awsplugin.NewAWSSigner(signer.NewFromConfig(awsConfig)), nil
}
