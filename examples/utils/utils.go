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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"

	"oras.land/oras-go/v2/registry"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	awsplugin "github.com/aws/aws-signer-notation-plugin/plugin"
)

// awsSignerRootURL is the url of AWS Signer's root certificate. The URL is copied from AWS Signer's documentation
// https://docs.aws.amazon.com/signer/latest/developerguide/image-signing-prerequisites.html
const awsSignerRootURL = "https://d2hvyiie56hcat.cloudfront.net/aws-signer-notation-root.cert"

// Cache to store AWS Signer's Root Certificate so that we dont need to fetch root certificate for every signature verification.
var awsSignerRootCache *x509.Certificate

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

// GetAWSSignerRootCert returns the AWS Signer's root certificate
func GetAWSSignerRootCert() (*x509.Certificate, error) {
	if awsSignerRootCache != nil {
		return awsSignerRootCache, nil
	}

	resp, err := http.Get(awsSignerRootURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	switch block.Type {
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		awsSignerRootCache = cert
	default:
		return nil, fmt.Errorf("unsupported certificate type :%s", block.Type)
	}

	return awsSignerRootCache, nil
}
