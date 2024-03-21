package main

import (
	"context"
	"fmt"
)

func main() {
	// variable required for signing
	ctx := context.Background()
	awsRegion := "us-west-2"
	ecrImageURI := "111122223333.dkr.ecr.region.amazonaws.com/curl@sha256:ca78e5f730f9a789ef8c63bb55275ac12dfb9e8099e6EXAMPLE"
	awsSignerProfileArn := "arn:aws:signer:region:111122223333:/signing-profiles/ecr_signing_profile"
	userMetadata := map[string]string{"buildId": "101"}

	// signing
	signer, err := NewNotationSigner(ctx, awsRegion)
	if err != nil {
		panic(err)
	}
	err = signer.Sign(ctx, awsSignerProfileArn, ecrImageURI, userMetadata)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Sucessfully signed %s", ecrImageURI)
}
