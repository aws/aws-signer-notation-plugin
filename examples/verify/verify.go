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

package main

import (
	"context"
	"crypto/x509"
	"fmt"

	"example/utils"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awsplugin "github.com/aws/aws-signer-notation-plugin/plugin"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/notaryproject/notation-plugin-framework-go/plugin"
)

// NotationVerifier facilitates signature verification for OCI artifacts using notation and AWS Signer plugin
type NotationVerifier struct {
	ecrClient    *ecr.Client
	signerPlugin *awsplugin.AWSSignerPlugin
}

// NewNotationVerifier creates various AWS service clients and returns NotationVerifier
func NewNotationVerifier(ctx context.Context, region string) (*NotationVerifier, error) {
	pl, err := utils.GetAWSSignerPlugin(ctx, region)
	if err != nil {
		return nil, err
	}
	ecrClient, err := utils.GetECRClient(ctx, region)
	if err != nil {
		return nil, err
	}
	return &NotationVerifier{
		ecrClient:    ecrClient,
		signerPlugin: pl,
	}, nil
}

// The Verify function verifies the signature stored in the registry against the provided truststore and trust policy using the Notation and AWS Signer plugin
func (n *NotationVerifier) Verify(ctx context.Context, reference string, trustedRoots []*x509.Certificate, tPolicy *trustpolicy.Document, userMetadata map[string]string) (*notation.VerificationOutcome, error) {
	ref, err := utils.ParseReference(reference)
	if err != nil {
		return nil, err
	}

	regClient, err := utils.GetNotationRepository(ctx, n.ecrClient, ref)
	if err != nil {
		return nil, err
	}

	tStore := &customTrustStore{certs: trustedRoots}
	pluginManager := &customPluginManager{awsPlugin: n.signerPlugin}
	sigVerifier, err := verifier.New(tPolicy, tStore, pluginManager)
	if err != nil {
		return nil, err
	}
	verifyOpts := notation.VerifyOptions{
		ArtifactReference:    reference,
		MaxSignatureAttempts: 100,
		UserMetadata:         userMetadata,
	}
	_, outcome, err := notation.Verify(ctx, sigVerifier, regClient, verifyOpts)
	if err != nil {
		return nil, err
	}

	return outcome[len(outcome)-1], nil
}

// customTrustStore implements truststore.X509TrustStore and returns the trusted certificates for a given trust-store.
// This implementation currently returns only AWS Signer trusted root but can be extended to support multiple trust-stores.
type customTrustStore struct {
	certs []*x509.Certificate
}

func (ts *customTrustStore) GetCertificates(_ context.Context, _ truststore.Type, _ string) ([]*x509.Certificate, error) {
	return ts.certs, nil
}

// customPluginManager implements plugin.Manager.
// This implementation currently supports AWS Signer plugin but can be extended to support any plugin.
type customPluginManager struct {
	awsPlugin *awsplugin.AWSSignerPlugin
}

func (p *customPluginManager) Get(_ context.Context, name string) (plugin.Plugin, error) {
	if name == awsplugin.Name {
		return p.awsPlugin, nil
	}
	return nil, fmt.Errorf("%s plugin not supported", name)
}

func (p *customPluginManager) List(_ context.Context) ([]string, error) {
	return []string{awsplugin.Name}, nil
}
