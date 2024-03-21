package main

import (
	"context"
	"crypto/x509"
	"fmt"

	"example/utils"

	"github.com/notaryproject/notation-go/verifier/trustpolicy"
)

func main() {
	// variable required for verification
	ctx := context.Background()
	awsRegion := "us-west-2"
	ecrImageURI := "111122223333.dkr.ecr.region.amazonaws.com/curl@sha256:ca78e5f730f9a789ef8c63bb55275ac12dfb9e8099e6EXAMPLE"
	awsSignerProfileArn := "arn:aws:signer:region:111122223333:/signing-profiles/ecr_signing_profile"
	userMetadata := map[string]string{"buildId": "101"}
	tPolicy := getTrustPolicy(awsSignerProfileArn)
	awsSignerRootCert, err := utils.GetAWSSignerRootCert()
	if err != nil {
		panic(err)
	}

	// signature verification
	verifier, err := NewNotationVerifier(ctx, awsRegion)
	if err != nil {
		panic(err)
	}
	outcome, err := verifier.Verify(context.Background(), ecrImageURI, []*x509.Certificate{awsSignerRootCert}, tPolicy, userMetadata)
	if err != nil {
		panic(err)
	}
	fmt.Printf("sucessfully verified signature for %s \n", ecrImageURI)
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
