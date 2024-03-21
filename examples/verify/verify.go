package main

import (
	"context"
	"crypto/x509"
	"fmt"

	"example/utils"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awsplugin "github.com/aws/aws-signer-notation-plugin/plugin"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/notaryproject/notation-plugin-framework-go/plugin"
)

type notationVerifier struct {
	ecrClient    *ecr.Client
	signerPlugin *awsplugin.AWSSignerPlugin
}

func newNotationVerifier(ctx context.Context, region string) (*notationVerifier, error) {
	pl, err := utils.GetAWSSignerPlugin(ctx, region)
	if err != nil {
		return nil, err
	}
	ecrClient, err := utils.GetECRClient(ctx, region)
	if err != nil {
		return nil, err
	}
	return &notationVerifier{
		ecrClient:    ecrClient,
		signerPlugin: pl,
	}, nil
}

func (n *notationVerifier) verify(ctx context.Context, reference string, trustedRoots []*x509.Certificate, tPolicy *trustpolicy.Document, userMetadata map[string]string) (*notation.VerificationOutcome, error) {
	ref, err := utils.ParseReference(reference)
	if err != nil {
		return nil, err
	}

	regClient, err := utils.GetNotationRepository(ctx, n.ecrClient, ref)
	if err != nil {
		return nil, err
	}

	tStore := &customTrustStore{certs: trustedRoots}
	pluginManager := &customPluginManager{awsPlugin: n.signerPlugin}
	sigVerifier, err := verifier.New(tPolicy, tStore, pluginManager)
	if err != nil {
		return nil, err
	}
	verifyOpts := notation.VerifyOptions{
		ArtifactReference:    reference,
		MaxSignatureAttempts: 100,
		UserMetadata:         userMetadata,
	}
	_, outcome, err := notation.Verify(ctx, sigVerifier, regClient, verifyOpts)
	if err != nil {
		return nil, err
	}

	return outcome[len(outcome)-1], nil
}

// customTrustStore returns AWS Signer root cert.
type customTrustStore struct {
	certs []*x509.Certificate
}

func (ts *customTrustStore) GetCertificates(_ context.Context, _ truststore.Type, _ string) ([]*x509.Certificate, error) {
	return ts.certs, nil
}

// customPluginManager manages plugins installed on the system.
type customPluginManager struct {
	awsPlugin *awsplugin.AWSSignerPlugin
}

func (p *customPluginManager) Get(_ context.Context, name string) (plugin.Plugin, error) {
	if name == awsplugin.Name {
		return p.awsPlugin, nil
	}
	return nil, fmt.Errorf("%s plugin not supported", name)
}

func (p *customPluginManager) List(_ context.Context) ([]string, error) {
	return []string{awsplugin.Name}, nil
}
