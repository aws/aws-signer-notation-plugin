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
	fmt.Printf("Sucessfully signed artifact: %s.", ecrImageURI)
}
