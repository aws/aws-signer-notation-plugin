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

type ecrRepoPassword struct {
	repo   notationregistry.Repository
	expiry time.Time
}

var credentialCache = map[string]ecrRepoPassword{}

func GetECRClient(ctx context.Context, region string) (*ecr.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return ecr.NewFromConfig(awsConfig), nil
}

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
	credentialCache[ref.Host()] = ecrRepoPassword{
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
