package main

import (
	"context"
	"fmt"
)

func main() {
	ecrImageURI := "111122223333.dkr.ecr.region.amazonaws.com/curl@sha256:ca78e5f730f9a789ef8c63bb55275ac12dfb9e8099e6EXAMPLE"
	awsRegion := "us-west-2"
	userMetadata := map[string]string{"buildId": "101"}
	awsSignerProfileArn := "arn:aws:signer:region:111122223333:/signing-profiles/ecr_signing_profile"

	ctx := context.Background()
	signer, err := newNotationSigner(ctx, awsRegion)
	if err != nil {
		panic(err)
	}

	err = signer.sign(ctx, awsSignerProfileArn, ecrImageURI, userMetadata)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Sucessfully signed %s", ecrImageURI)
}
