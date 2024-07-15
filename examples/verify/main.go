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
	"encoding/pem"
	"fmt"
	"io"
	"net/http"

	"github.com/notaryproject/notation-go/verifier/trustpolicy"
)

// awsSignerRootURL is the url of AWS Signer's root certificate. The URL is copied from AWS Signer's documentation
// https://docs.aws.amazon.com/signer/latest/developerguide/image-signing-prerequisites.html
const awsSignerRootURL = "https://d2hvyiie56hcat.cloudfront.net/aws-signer-notation-root.cert"

// Downloads and caches AWS Signer's Root Certificate required for signature verification.
var awsSignerRoot = getAWSSignerRootCert()

func main() {
	// variable required for verification
	ctx := context.Background()
	awsRegion := "us-west-2"                                                                   // AWS region where you created signing profile and ECR image.
	ecrImageURI := "111122223333.dkr.ecr.region.amazonaws.com/curl@sha256:EXAMPLEHASH"         // ECR image URI
	awsSignerProfileArn := "arn:aws:signer:region:111122223333:/signing-profiles/profile_name" // AWS Signer's signing profile ARN
	userMetadata := map[string]string{"buildId": "101"}                                        // Optional, add if you want to verify metadata in the signature, else use nil
	tPolicy := getTrustPolicy(awsSignerProfileArn)

	// signature verification
	verifier, err := NewNotationVerifier(ctx, awsRegion)
	if err != nil {
		panic(err)
	}
	outcome, err := verifier.Verify(context.Background(), ecrImageURI, []*x509.Certificate{awsSignerRoot}, tPolicy, userMetadata)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Sucessfully verified signature associated with %s.\n", ecrImageURI)
	for _, out := range outcome.VerificationResults {
		fmt.Printf("  individual validation result: %+v \n", out)
	}
}

// getTrustPolicy returns global trust policy
func getTrustPolicy(awsSignerProfileArn string) *trustpolicy.Document {
	return &trustpolicy.Document{
		Version: "1.0",
		TrustPolicies: []trustpolicy.TrustPolicy{
			{
				Name:           "global_trust_policy",
				RegistryScopes: []string{"*"},
				SignatureVerification: trustpolicy.SignatureVerification{
					VerificationLevel: "strict",
				},
				TrustStores:       []string{"signingAuthority:aws-signer-ts"},
				TrustedIdentities: []string{awsSignerProfileArn},
			},
		},
	}
}

// getAWSSignerRootCert returns the AWS Signer's root certificate
func getAWSSignerRootCert() *x509.Certificate {
	resp, err := http.Get(awsSignerRootURL)
	if err != nil {
		panic(fmt.Sprintf("failed to get AWS Signer's root certificate: %s", err.Error())) // handle error
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to get AWS Signer's root certificate: %s", err.Error())) // handle error
	}

	block, _ := pem.Decode(data)
	switch block.Type {
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			panic(fmt.Sprintf("failed to parse AWS Signer's root certificate: %s", err.Error())) // handle error
		}
		return cert
	default:
		panic(fmt.Sprintf("failed to parse AWS Signer's root certificate: unsupported certificate type :%s", block.Type)) // handle error
	}
}
