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

	"example/utils"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awsplugin "github.com/aws/aws-signer-notation-plugin/plugin"
	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/signer"
)

// NotationSigner facilitates signing of OCI artifacts using notation and AWS Signer plugin
type NotationSigner struct {
	ecrClient    *ecr.Client
	signerPlugin *awsplugin.AWSSignerPlugin
}

// NewNotationSigner creates various AWS service clients and returns NotationSigner
func NewNotationSigner(ctx context.Context, region string) (*NotationSigner, error) {
	pl, err := utils.GetAWSSignerPlugin(ctx, region)
	if err != nil {
		return nil, err
	}
	ecrClient, err := utils.GetECRClient(ctx, region)
	if err != nil {
		return nil, err
	}
	return &NotationSigner{
		ecrClient:    ecrClient,
		signerPlugin: pl,
	}, nil
}

// Sign function signs the given reference stored in registry and pushes signature back to registry.
func (n *NotationSigner) Sign(ctx context.Context, keyId, reference string, userMetadata map[string]string) error {
	ref, err := utils.ParseReference(reference)
	if err != nil {
		return err
	}

	regClient, err := utils.GetNotationRepository(ctx, n.ecrClient, ref)
	if err != nil {
		return err
	}

	opts := notation.SignOptions{
		SignerSignOptions: notation.SignerSignOptions{
			SignatureMediaType: jws.MediaTypeEnvelope,
			SigningAgent:       "aws-signer-notation-go-example/1.0.0",
		},
		ArtifactReference: reference,
		UserMetadata:      userMetadata,
	}

	sigSigner, err := signer.NewFromPlugin(n.signerPlugin, keyId, map[string]string{})
	if err != nil {
		return err
	}

	_, err = notation.Sign(ctx, sigSigner, regClient, opts)
	return err
}
