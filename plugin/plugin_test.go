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

package plugin

import (
	"context"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-signer-notation-plugin/internal/client"
	"go.uber.org/mock/gomock"
	"github.com/notaryproject/notation-plugin-framework-go/plugin"
	"github.com/stretchr/testify/assert"
)

const (
	testProfileArn        = "arn:aws:signer:us-west-2:000000000000:/signing-profiles/NotaryPluginIntegProfile"
	testProfileVersionArn = testProfileArn + "/OF8IVUsPJq"
	testJobArn            = "arn:aws:signer:us-west-2:000000000000:/signing-jobs/97af3947-e7b2-4533-8d9d-6741156f0b79"
	testCertificate       = `-----BEGIN CERTIFICATE-----
MIIDQDCCAiigAwIBAgIRAMH0R+Owv6zXRzRJgjkWUPEwDQYJKoZIhvcNAQELBQAw
ETEPMA0GA1UECgwGY2hpZW5iMB4XDTIyMDcxNTE3MjQ0MVoXDTIzMDgxNTE4MjQ0
MFowEjEQMA4GA1UEAwwHZm9vLmJhcjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBANn0mo5gw6VYKfLGHre6zy6eo6f1Fe2p2o5nbClmkA43OgWF0ngnwJPd
Hhfy17pqDOrfs3Uj8gGwhlZYbWVYORWHGwbHRV9FsBP3wq8HrQ2I+7UAZNsRBxWQ
Lbo0ha0NzYLIG1DYuPrNCBSzdlkjNhNZJR8QRn0+5LW8AfcOD3x6UBhDgk8kE/Y/
9outGzynHVDXObpylh6xie+PXJ6y8aPM0PZszwWv+mJznXchyvrVDUxpETI/EnL9
QMq2STEgAS0f8PCYQkKxz1s1ODb2AWwuIdqJmDhmwkYs4kqV/kyNN42H6gfgSQXf
IJMLX2fn/ZOz431jV8fUDSKUFSdJw2sCAwEAAaOBkTCBjjASBgNVHREECzAJggdm
b28uYmFyMAkGA1UdEwQCMAAwHwYDVR0jBBgwFoAU3gzqhkSDrYSfGn5E8e/3qUAw
xowwHQYDVR0OBBYEFETbSw2Lt2WIQlolvzg1lKadc0oQMA4GA1UdDwEB/wQEAwIF
oDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDQYJKoZIhvcNAQELBQAD
ggEBALc3rxSZLVc4ammm7XKQDAh7B+MX4LOj8TleVr/aYZ1iN9y2VVsKmUtLCJBa
gU2HWaoAQN2qO0FvbANuBpgf2dF8fUFygKURo9fnFSIng2rgd38cIrJP0mYYPg4x
EizD3ZznlFE7zu4MVBcZTOTAgqyzsjg/K1YfdBTCmEoNv09P7u4r1KiATBsaiKaH
h770TLUfa+PzpbIinp2cF/XYVchepCiCJDAdTR1tWKHaqeuW/WQHKso7Z6wyPO24
d3m5GyGuIRMddbp6zclSRP/I4TCS/0cOru9ATc94PaKWjDOTClYH8ykRZom8OICq
KCzg3o7lofVNdVFxDM8rrMJ06cY=
-----END CERTIFICATE-----`
)

func TestNew(t *testing.T) {
	c, _ := client.NewAWSSigner(context.TODO(), map[string]string{})
	awsSignerPlugin := NewAWSSigner(c)
	assert.NotNil(t, awsSignerPlugin, "NewAWSSigner should return a non-nil instance of AWSSignerPlugin")
}

func TestNewForCLI(t *testing.T) {
	awsSignerPlugin := NewAWSSignerForCLI()
	assert.NotNil(t, awsSignerPlugin, "NewAWSSignerForCLI should return a non-nil instance of AWSSignerPlugin")
}

func TestVerifySignature(t *testing.T) {
	request, expectedResp := getVerifySignatureRequestResponse()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSignerClient := client.NewMockInterface(mockCtrl)
	mockSignerClient.EXPECT().GetRevocationStatus(gomock.Any(), gomock.Any()).Return(&signer.GetRevocationStatusOutput{RevokedEntities: []string{}}, nil)

	resp, err := NewAWSSigner(mockSignerClient).VerifySignature(context.TODO(), request)
	assert.NoError(t, err, "VerifySignature() returned error")
	assert.Equal(t, expectedResp, resp, "VerifySignatureResponse mismatch")
}

func TestVerifySignature_ValidationError(t *testing.T) {
	tests := map[string]*plugin.VerifySignatureRequest{
		"nilRequest":     nil,
		"invalidRequest": {ContractVersion: ""},
	}
	for name, req := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewAWSSignerForCLI().VerifySignature(context.TODO(), req)
			assert.Error(t, err, "VerifySignatureRequest() expected error but not found")
		})
	}
}

func TestGenerateEnvelope(t *testing.T) {
	request, expectedResp := getGenerateEnvRequestResponse()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSignerClient := client.NewMockInterface(mockCtrl)
	mockSignerClient.EXPECT().SignPayload(gomock.Any(), gomock.Any()).Return(&signer.SignPayloadOutput{Signature: expectedResp.SignatureEnvelope}, nil)

	resp, err := NewAWSSigner(mockSignerClient).GenerateEnvelope(context.TODO(), request)
	assert.NoError(t, err, "GenerateEnvelope() returned error")
	assert.Equal(t, expectedResp, resp, "GenerateEnvelopeResponse mismatch")
}

func TestGenerateEnvelope_ValidationError(t *testing.T) {
	tests := map[string]*plugin.GenerateEnvelopeRequest{
		"nilRequest":     nil,
		"invalidRequest": {ContractVersion: ""},
	}
	for name, req := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewAWSSignerForCLI().GenerateEnvelope(context.TODO(), req)
			assert.Error(t, err, "GenerateEnvelope() expected error but not found")
		})
	}
}

func TestGetMetadata(t *testing.T) {
	resp, err := NewAWSSignerForCLI().GetMetadata(context.TODO(), nil)
	assert.NoError(t, err, "GetMetadata() returned error")
	assert.Equal(t, "com.amazonaws.signer.notation.plugin", resp.Name)
	assert.Equal(t, "AWS Signer plugin for Notation", resp.Description)
	assert.Equal(t, "https://docs.aws.amazon.com/signer", resp.URL)
	assert.Equal(t, []string{plugin.ContractVersion}, resp.SupportedContractVersions)
	assert.NotEmpty(t, resp.Version, "version is empty")
}

func TestGenerateSignature_Error(t *testing.T) {
	_, err := NewAWSSignerForCLI().GenerateSignature(context.TODO(), nil)
	assert.Error(t, err, "expected UnsupportedError but not found")
}

func TestDescribeKey_Error(t *testing.T) {
	_, err := NewAWSSignerForCLI().DescribeKey(context.TODO(), nil)
	assert.Error(t, err, "expected UnsupportedError but not found")
}

func TestGetSignerClientIfNotPresent(t *testing.T) {
	signerPlugin := AWSSignerPlugin{}
	err := signerPlugin.setSignerClientIfNotPresent(context.TODO(), nil)
	assert.NoError(t, err)
}

func convertCert(certs ...string) [][]byte {
	var o [][]byte
	for _, cert := range certs {
		block, _ := pem.Decode([]byte(cert))
		o = append(o, block.Bytes)
	}
	return o
}

func getVerifySignatureRequestResponse() (*plugin.VerifySignatureRequest, *plugin.VerifySignatureResponse) {
	attrSigningJob := "com.amazonaws.signer.signingJob"
	attrSigningProfileVersion := "com.amazonaws.signer.signingProfileVersion"
	now := time.Now()
	req := &plugin.VerifySignatureRequest{
		ContractVersion: "1.0",
		Signature: plugin.Signature{
			CriticalAttributes: plugin.CriticalAttributes{
				ContentType:          "application/vnd.cncf.notary.payload.v1+json",
				SigningScheme:        "notary.x509.signingAuthority",
				AuthenticSigningTime: &now,
				Expiry:               &now,
				ExtendedAttributes: map[string]interface{}{
					attrSigningJob:            testJobArn,
					attrSigningProfileVersion: testProfileVersionArn,
				},
			},
			UnprocessedAttributes: []string{
				attrSigningJob,
				attrSigningProfileVersion,
			},
			CertificateChain: convertCert(testCertificate),
		},
		TrustPolicy: plugin.TrustPolicy{
			TrustedIdentities: []string{testProfileArn},
			SignatureVerification: []plugin.Capability{
				plugin.CapabilityTrustedIdentityVerifier,
				plugin.CapabilityRevocationCheckVerifier,
			},
		},
	}

	expectedResp := &plugin.VerifySignatureResponse{
		VerificationResults: map[plugin.Capability]*plugin.VerificationResult{
			plugin.CapabilityTrustedIdentityVerifier: {
				Success: true,
				Reason:  fmt.Sprintf("Signature publisher matched \"%s\" trusted identity.", testProfileArn),
			},
			plugin.CapabilityRevocationCheckVerifier: {
				Success: true,
				Reason:  "Signature is not revoked.",
			},
		},
		ProcessedAttributes: []interface{}{attrSigningProfileVersion, attrSigningJob},
	}

	return req, expectedResp
}

func getGenerateEnvRequestResponse() (*plugin.GenerateEnvelopeRequest, *plugin.GenerateEnvelopeResponse) {
	testPayloadType := "application/vnd.oci.descriptor.v1+json"
	testSigEnvType := "application/jose+json"
	req := &plugin.GenerateEnvelopeRequest{
		ContractVersion:       plugin.ContractVersion,
		SignatureEnvelopeType: testSigEnvType,
		Payload:               []byte("sigME!"),
		PayloadType:           testPayloadType,
		KeyID:                 "arn:aws:signer:us-west-2:780792624090:/signing-profiles/NotationProfile",
	}
	expectedResp := &plugin.GenerateEnvelopeResponse{
		SignatureEnvelope:     []byte("sigEnv"),
		SignatureEnvelopeType: testSigEnvType,
		Annotations:           nil,
	}
	return req, expectedResp
}
