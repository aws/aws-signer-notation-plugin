package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/signer"
)

// Interface facilitates unit testing
type Interface interface {
	SignPayload(ctx context.Context, params *signer.SignPayloadInput, optFns ...func(*signer.Options)) (*signer.SignPayloadOutput, error)
	GetRevocationStatus(ctx context.Context, params *signer.GetRevocationStatusInput, optFns ...func(*signer.Options)) (*signer.GetRevocationStatusOutput, error)
}
