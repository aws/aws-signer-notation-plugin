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
	"encoding/base64"
	"strings"
	"time"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	notationregistry "github.com/notaryproject/notation-go/registry"
)

type repoAndExpiry struct {
	repo   notationregistry.Repository
	expiry time.Time
}

// credentialCache is unbounded cache used to store notationregistry.Repository.
// The cache is required since Amazon ECR credentials expire after 12 hours.
var credentialCache = map[string]repoAndExpiry{}

// GetECRClient returns ecr client for give region.
func GetECRClient(ctx context.Context, region string) (*ecr.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return ecr.NewFromConfig(awsConfig), nil
}

// GetNotationRepository creates notationregistry.Repository required to access artifacts in Amazon ECR for sign and verify operations.
func GetNotationRepository(ctx context.Context, client *ecr.Client, ref registry.Reference) (notationregistry.Repository, error) {
	repoPass, ok := credentialCache[ref.Host()]
	// Check if Repository object exists in cache and is not expired
	if ok && time.Now().Before(repoPass.expiry) {
		return repoPass.repo, nil
	}

	// else fetch credential from ECR
	cred, expiry, err := getECRCredentials(ctx, client)
	if err != nil {
		return nil, err
	}

	authClient := &auth.Client{
		Credential: auth.StaticCredential(ref.Host(), cred),
		Cache:      auth.NewCache(),
		ClientID:   "example-notation-go",
	}
	authClient.SetUserAgent("aws-signer-notation-go-example/1.0")

	remoteRepo := &remote.Repository{
		Client:    authClient,
		Reference: ref,
	}
	remoteRepo.SetReferrersCapability(false)

	notationRepo := notationregistry.NewRepository(remoteRepo)
	credentialCache[ref.Host()] = repoAndExpiry{
		repo:   notationRepo,
		expiry: *expiry,
	}
	return notationRepo, nil
}

func getECRCredentials(ctx context.Context, client *ecr.Client) (auth.Credential, *time.Time, error) {
	token, err := client.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return auth.Credential{}, nil, err
	}
	authData := token.AuthorizationData[0]
	creds, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return auth.Credential{}, nil, err
	}

	// Get password from credential pair
	credsSplit := strings.Split(string(creds), ":")
	credential := auth.Credential{
		Username: "AWS",
		Password: credsSplit[1],
	}

	return credential, authData.ExpiresAt, nil
}
