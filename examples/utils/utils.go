package utils

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"

	"oras.land/oras-go/v2/registry"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	awsplugin "github.com/aws/aws-signer-notation-plugin/plugin"
)

const awsSignerRootURL = "https://d2hvyiie56hcat.cloudfront.net/aws-signer-notation-root.cert"

var awsSignerRootCache *x509.Certificate

func ParseReference(reference string) (registry.Reference, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return registry.Reference{}, fmt.Errorf("%q: %w. Expecting <registry>/<repository>@<digest>", reference, err)
	}
	if ref.Reference == "" {
		return registry.Reference{}, fmt.Errorf("%q: invalid reference. Expecting <registry>/<repository>@<digest>", reference)
	}
	if err := ref.ValidateReferenceAsDigest(); err != nil {
		return registry.Reference{}, fmt.Errorf("%q: tag resolution not supported. Expecting <registry>/<repository>@<digest>", reference)

	}
	return ref, nil
}

func GetAWSSignerPlugin(ctx context.Context, region string) (*awsplugin.AWSSignerPlugin, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return awsplugin.NewAWSSigner(signer.NewFromConfig(awsConfig)), nil
}

func GetAWSSignerRootCert() (*x509.Certificate, error) {
	if awsSignerRootCache != nil {
		return awsSignerRootCache, nil
	}

	resp, err := http.Get(awsSignerRootURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	switch block.Type {
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		awsSignerRootCache = cert
	default:
		return nil, fmt.Errorf("unsupported certificate type :%s", block.Type)
	}

	return awsSignerRootCache, nil
}
