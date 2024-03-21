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

type notationSigner struct {
	ecrClient    *ecr.Client
	signerPlugin *awsplugin.AWSSignerPlugin
}

func newNotationSigner(ctx context.Context, region string) (*notationSigner, error) {
	pl, err := utils.GetAWSSignerPlugin(ctx, region)
	if err != nil {
		return nil, err
	}
	ecrClient, err := utils.GetECRClient(ctx, region)
	if err != nil {
		return nil, err
	}
	return &notationSigner{
		ecrClient:    ecrClient,
		signerPlugin: pl,
	}, nil
}

func (n *notationSigner) sign(ctx context.Context, keyId, reference string, userMetadata map[string]string) error {
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
			SigningAgent:       "zop-zap",
		},
		ArtifactReference: reference,
		UserMetadata:      userMetadata,
	}

	sigSigner, err := signer.NewPluginSigner(n.signerPlugin, keyId, map[string]string{})
	if err != nil {
		return err
	}

	_, err = notation.Sign(ctx, sigSigner, regClient, opts)
	return err
}
