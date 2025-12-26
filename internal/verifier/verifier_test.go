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

package verifier

import (
	"context"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-signer-notation-plugin/internal/client"
	"github.com/aws/smithy-go"
	"github.com/notaryproject/notation-plugin-framework-go/plugin"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go-v2/service/signer"
	"go.uber.org/mock/gomock"
)

const (
	testProfileArn        = "arn:aws:signer:us-west-2:000000000000:/signing-profiles/NotaryPluginIntegProfile"
	testProfileVersionArn = testProfileArn + "/OF8IVUsPJq"
	testJobArn            = "arn:aws:signer:us-west-2:000000000000:/signing-jobs/97af3947-e7b2-4533-8d9d-6741156f0b79"
	testCertificate1      = `-----BEGIN CERTIFICATE-----
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
	testCertificate1Hash = "13a01b7e1de3aee0367615c59f6d001238913e594626d0e3c8784489b15a18fada1c31f39d3ba9318cb673ffd8cd679b"
	testCertificate2     = `-----BEGIN CERTIFICATE-----
MIIC7zCCAdegAwIBAgIRAPxhWP65yw1qFSMD39FxuUwwDQYJKoZIhvcNAQELBQAw
ETEPMA0GA1UECgwGY2hpZW5iMB4XDTE5MTAwNzE4MDIxMVoXDTI5MTAwNzE5MDIx
MVowETEPMA0GA1UECgwGY2hpZW5iMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEA4SSFBInasnQCgPLDZzz0NNlTlRm4yn2lyUP7gEzBQZc0Hp+PKE3dnMGH
bQ6w0FmGD5sMKMTIfUCRJyjiJPi0RvCEOmU+nY2UYZf+ttrVx33pWrHpkxXORxA4
rp7SzxP5GFl78Mo0CEFxOKHPqLC/Nm4SmQKhhMUJkiqc3X/9WFigBIfkFXLFZQ64
yoCq+ekvKW9GGh2Mq9VwSnB+6wem/3mPJ8x4sX1UtGu/DL5gc7gyzVCbfn8SZpb6
L7y++9zGmRwmcKMv8IaLj07fr9Ho34zm9CbwMHwUZeHC5uXcGR54t9sTNq5rgu1k
Q9LskOmPcEEkTkyKtrAs5WKHrSYdWQIDAQABo0IwQDAPBgNVHRMBAf8EBTADAQH/
MB0GA1UdDgQWBBTeDOqGRIOthJ8afkTx7/epQDDGjDAOBgNVHQ8BAf8EBAMCAYYw
DQYJKoZIhvcNAQELBQADggEBAJSVnGSdpX6nSYcsCMHu99dN/xVn+Qtvj0ovdKQo
JC5cQNjFQ7wXCSgYa2DtSMQ0McysZ+TkNWDGwi2c+HCoHAL/XNWDU261Hj/VwVI4
2p46Q4UzpWmhx5dkDV2xRhK8QMPwW2NRQqkd/75FUfRpq5xdL4IzeaNcYKXMBJyX
zSZee7oqEixEVzis7Ex7mvXBiRdjZBp8cFuRJVKPBgK7SmFkJwyLtd2OLtNehUsh
Af8fCVvIhr9YxXK+RqiRUhvJDrS9DlKA6dT4KvR41B/a8NLf6PJGyHdSFuvKZr0z
C+gMfNFGs1L2QLg1+xnoLHIey4tRXYHjpD2b/KALNr4/v+c=
-----END CERTIFICATE-----`
	testCertificate2Hash = "ff41924f0940448d7e46b8c327e129813b1442fb17c9b2a86d49edcb00b707c9662f561c8a3e11a592b25061d488f2a3"
)

var testTISuccessReason = fmt.Sprintf(reasonTrustedIdentitySuccessFmt, testProfileArn)

func TestVerify(t *testing.T) {
	request := mockVerifySigRequest()
	expectedGRSInput := getRevocationStatusInput(request)
	grsOutput := &signer.GetRevocationStatusOutput{
		RevokedEntities: []string{},
	}
	mockSignerClient, mockCtrl := getMockClient(expectedGRSInput, grsOutput, nil, t)
	defer mockCtrl.Finish()

	actualResponse, err := New(mockSignerClient).Verify(context.TODO(), request)
	expectedResponse := getVerifySigResponse(true, testTISuccessReason, true, reasonNotRevoked)
	validateResponse(t, expectedResponse, *actualResponse, err)
}

func TestVerify_ValidTrustedIdentity(t *testing.T) {
	tests := map[string]struct {
		tis        []string
		expectedTi string
	}{
		"signingProfileArn": {
			tis:        []string{testProfileArn},
			expectedTi: testProfileArn,
		},
		"signingProfileVersionArn": {
			tis:        []string{testProfileVersionArn},
			expectedTi: testProfileVersionArn,
		},
		"signingProfileAndVersionArn": {
			tis:        []string{testProfileArn, testProfileVersionArn},
			expectedTi: testProfileArn,
		},
		"signingProfileVersionAndProfileArn": {
			tis:        []string{testProfileVersionArn, testProfileArn},
			expectedTi: testProfileVersionArn,
		},
		"signingProfileArnWithGarbage": {
			tis:        []string{"shop", testProfileArn, "zop"},
			expectedTi: testProfileArn,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockSignerClient, mockCtrl := getMockClient(nil, &signer.GetRevocationStatusOutput{}, nil, t)
			defer mockCtrl.Finish()

			expectedResponse := getVerifySigResponse(true, fmt.Sprintf(reasonTrustedIdentitySuccessFmt, test.expectedTi), true, reasonNotRevoked)

			request := mockVerifySigRequest()
			request.TrustPolicy.TrustedIdentities = test.tis
			actualResponse, err := New(mockSignerClient).Verify(context.TODO(), request)

			validateResponse(t, expectedResponse, *actualResponse, err)
		})
	}
}

func TestVerify_InvalidTrustedIdentity(t *testing.T) {
	tests := map[string][]string{
		"empty":                     {},
		"nonArn":                    {"Cheers!🍺"},
		"nonAWSSignerArn":           {"arn:aws:dynamodb:us-east-2:123456789012:table/myDynamoDBTable"},
		"SigningJobArn":             {testJobArn},
		"invalidSignerArn1":         {"arn:aws:signer:us-west-2:000000000000:/beer/NotaryPluginIntegProfile"},
		"invalidSigningProfileArn":  {"arn:us-west-2:000000000000:/signing-profile/NotaryPluginIntegProfile/1234/asda"},
		"invalidSigningProfileArn2": {testProfileVersionArn + "/asda"},
	}

	for name, tis := range tests {
		t.Run(name, func(t *testing.T) {
			expectedResponse := getVerifySigResponse(false, reasonTrustedIdentityFailure, true, reasonNotRevoked)
			delete(expectedResponse.VerificationResults, plugin.CapabilityRevocationCheckVerifier)

			request := mockVerifySigRequest()
			request.TrustPolicy.TrustedIdentities = tis
			request.TrustPolicy.SignatureVerification = []plugin.Capability{plugin.CapabilityTrustedIdentityVerifier}
			c, _ := client.NewAWSSigner(context.Background(), map[string]string{})
			actualResponse, err := New(c).Verify(context.TODO(), request)

			validateResponse(t, expectedResponse, *actualResponse, err)
		})
	}
}

func TestVerify_RevokedResources(t *testing.T) {
	tests := map[string]struct {
		revokedResources []string
		errorMsg         string
	}{
		"revokedSigningJob": {
			revokedResources: []string{testJobArn},
			errorMsg:         fmt.Sprintf(reasonRevokedResourceFmt, testJobArn),
		},
		"revokedSigningProfile": {
			revokedResources: []string{testProfileVersionArn},
			errorMsg:         fmt.Sprintf(reasonRevokedResourceFmt, testProfileVersionArn),
		},
		"revokedCertificate": {
			revokedResources: []string{testCertificate1Hash},
			errorMsg:         reasonRevokedCertificate,
		},
		"revokedCertificates": {
			revokedResources: []string{testCertificate1Hash, testCertificate2Hash},
			errorMsg:         reasonRevokedCertificate,
		},
		"revokedSigningJobAndProfile": {
			revokedResources: []string{testJobArn, testProfileArn},
			errorMsg:         fmt.Sprintf(reasonRevokedResourceFmt, testJobArn+", "+testProfileArn),
		},
		"revokedSigningJobAndCert": {
			revokedResources: []string{testJobArn, testCertificate1},
			errorMsg:         fmt.Sprintf(reasonRevokedResourceFmt, testJobArn) + reasonRevokedCertificate,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			revokedGRSOutput := signer.GetRevocationStatusOutput{
				RevokedEntities: test.revokedResources,
			}
			mockSignerClient, mockCtrl := getMockClient(nil, &revokedGRSOutput, nil, t)
			defer mockCtrl.Finish()

			expectedResponse := getVerifySigResponse(true, testTISuccessReason, false, test.errorMsg)
			actualResponse, err := New(mockSignerClient).Verify(context.TODO(), mockVerifySigRequest())

			validateResponse(t, expectedResponse, *actualResponse, err)
		})
	}
}

func TestVerify_RevocationCheckError(t *testing.T) {
	apiError := smithy.GenericAPIError{
		Code:    "ERROR",
		Message: "AWSSigner unreachable. 5xx",
	}
	mockSignerClient, mockCtrl := getMockClient(nil, nil, &apiError, t)
	defer mockCtrl.Finish()

	revReason := "GetRevocationStatus call failed with error: api error ERROR: " + apiError.ErrorMessage()
	expectedResponse := getVerifySigResponse(true, testTISuccessReason, false, revReason)
	actualResponse, err := New(mockSignerClient).Verify(context.TODO(), mockVerifySigRequest())

	validateResponse(t, expectedResponse, *actualResponse, err)
}

func TestVerify_MalformedRequest(t *testing.T) {
	badContractVersionReq := mockVerifySigRequest()
	badContractVersionReq.ContractVersion = "2.0"

	wildcardTrustedIdentityReq := mockVerifySigRequest()
	wildcardTrustedIdentityReq.TrustPolicy.TrustedIdentities = []string{"*"}

	invalidArnTrustedIdentityReq := mockVerifySigRequest()
	invalidArnTrustedIdentityReq.TrustPolicy.TrustedIdentities = []string{"*"}

	unsupportedCapability := mockVerifySigRequest()
	unsupportedCapability.TrustPolicy.SignatureVerification = append(unsupportedCapability.TrustPolicy.SignatureVerification, "unsupported")

	invalidCertReq := mockVerifySigRequest()
	invalidCertReq.Signature.CertificateChain = [][]byte{[]byte("BadCertificate")}

	zeroAuthSignTImeReq := mockVerifySigRequest()
	zeroAuthSignTImeReq.Signature.CriticalAttributes.AuthenticSigningTime = &time.Time{}

	invalidSignSchemeReq := mockVerifySigRequest()
	invalidSignSchemeReq.Signature.CriticalAttributes.SigningScheme = "badSignScheme"

	invalidCritAttr := mockVerifySigRequest()
	invalidCritAttr.Signature.CriticalAttributes.ExtendedAttributes[attrSigningProfileVersion] = nil

	invalidCritAttrJob := mockVerifySigRequest()
	//invalidCritAttrJob.TrustPolicy.SignatureVerification = []plugin.Capability{plugin.CapabilityRevocationCheckVerifier}
	//invalidCritAttr.Signature.CriticalAttributes.ExtendedAttributes[attrSigningJob] = "asda"
	delete(invalidCritAttrJob.Signature.CriticalAttributes.ExtendedAttributes, attrSigningJob)
	invalidCritAttrJob.Signature.CriticalAttributes.ExtendedAttributes = nil

	invalidCritAttrProfile := mockVerifySigRequest()
	invalidCritAttrProfile.TrustPolicy.SignatureVerification = []plugin.Capability{plugin.CapabilityRevocationCheckVerifier}
	delete(invalidCritAttrProfile.Signature.CriticalAttributes.ExtendedAttributes, attrSigningProfileVersion)

	tests := map[string]struct {
		req      *plugin.VerifySignatureRequest
		code     plugin.ErrorCode
		errorMsg string
	}{
		"badContractVersionReq": {
			req:      badContractVersionReq,
			code:     plugin.ErrorCodeUnsupportedContractVersion,
			errorMsg: "\"2.0\" is not a supported notary plugin contract version",
		},
		"wildcardTrustedIdentityReq": {
			req:      wildcardTrustedIdentityReq,
			code:     plugin.ErrorCodeValidation,
			errorMsg: errMsgWildcardIdentity,
		},
		"unsupportedCapability": {
			req:      unsupportedCapability,
			code:     plugin.ErrorCodeValidation,
			errorMsg: "'unsupported' is not a supported plugin capability",
		},
		"invalidCertReq": {
			req:      invalidCertReq,
			code:     plugin.ErrorCodeValidation,
			errorMsg: errMsgCertificateParse,
		},
		"zeroAuthSignTImeReq": {
			req:      zeroAuthSignTImeReq,
			code:     plugin.ErrorCodeValidation,
			errorMsg: "missing authenticSigningTime",
		},
		"invalidSignSchemeReq": {
			req:      invalidSignSchemeReq,
			code:     plugin.ErrorCodeValidation,
			errorMsg: "'badSignScheme' signing scheme is not supported",
		},
		"invalidCritAttr": {
			req:      invalidCritAttr,
			code:     plugin.ErrorCodeValidation,
			errorMsg: "unable to parse attribute \"com.amazonaws.signer.signingProfileVersion\".",
		},
		"invalidCritAttrJob": {
			req:      invalidCritAttrJob,
			code:     plugin.ErrorCodeValidation,
			errorMsg: "unable to parse attribute \"com.amazonaws.signer.signingProfileVersion\".",
		},
		"invalidCritAttrProfile": {
			req:      invalidCritAttrProfile,
			code:     plugin.ErrorCodeValidation,
			errorMsg: "unable to parse attribute \"com.amazonaws.signer.signingProfileVersion\".",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			awsSigner, _ := client.NewAWSSigner(context.TODO(), map[string]string{})
			_, err := New(awsSigner).Verify(context.TODO(), test.req)
			plgErr := toPluginError(err, t)
			assert.Equal(t, test.errorMsg, plgErr.Message, "error message mismatch")
			assert.Equal(t, test.code, plgErr.ErrCode, "error code mismatch")
		})
	}
}

func validateResponse(t *testing.T, expected plugin.VerifySignatureResponse, actual plugin.VerifySignatureResponse, err error) {
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
	} else {
		assert.Equal(t, expected, actual, "VerifySignatureResponse mismatch")
	}
}

func getVerifySigResponse(ti bool, tiReason string, notRevoked bool, revokedReason string) plugin.VerifySignatureResponse {
	return plugin.VerifySignatureResponse{
		VerificationResults: map[plugin.Capability]*plugin.VerificationResult{
			plugin.CapabilityTrustedIdentityVerifier: {Success: ti, Reason: tiReason},
			plugin.CapabilityRevocationCheckVerifier: {Success: notRevoked, Reason: revokedReason},
		},
		ProcessedAttributes: []interface{}{
			attrSigningProfileVersion,
			attrSigningJob,
		},
	}
}

func mockVerifySigRequest() *plugin.VerifySignatureRequest {
	signingTime, _ := time.Parse(time.RFC3339, "2022-07-06T19:10:28+00:00")
	expiryTime, _ := time.Parse(time.RFC3339, "2022-10-06T07:01:20Z")
	req := plugin.VerifySignatureRequest{
		ContractVersion: "1.0",
		Signature: plugin.Signature{
			CriticalAttributes: plugin.CriticalAttributes{
				ContentType:          "application/vnd.cncf.notary.payload.v1+json",
				SigningScheme:        signingSchemeAuthority,
				AuthenticSigningTime: &signingTime,
				Expiry:               &expiryTime,
				ExtendedAttributes: map[string]interface{}{
					attrSigningJob:            testJobArn,
					attrSigningProfileVersion: testProfileVersionArn,
				},
			},
			UnprocessedAttributes: []string{
				attrSigningJob,
				attrSigningProfileVersion,
			},
			CertificateChain: convertCert(testCertificate1, testCertificate2),
		},
		TrustPolicy: plugin.TrustPolicy{
			TrustedIdentities: []string{
				testProfileArn,
				"x509.subject: C=US, ST=WA, L=Seattle, O=acme-rockets.io, OU=Finance, CN=SecureBuilder",
			},
			SignatureVerification: []plugin.Capability{
				plugin.CapabilityTrustedIdentityVerifier,
				plugin.CapabilityRevocationCheckVerifier,
			},
		},
	}
	return &req
}

func convertCert(certs ...string) [][]byte {
	var o [][]byte
	for _, cert := range certs {
		block, _ := pem.Decode([]byte(cert))
		o = append(o, block.Bytes)
	}
	return o
}

func getRevocationStatusInput(request *plugin.VerifySignatureRequest) *signer.GetRevocationStatusInput {
	signingJobArn, _ := request.Signature.CriticalAttributes.ExtendedAttributes[attrSigningJob].(string)
	profileVersionArn, _ := request.Signature.CriticalAttributes.ExtendedAttributes[attrSigningProfileVersion].(string)

	return &signer.GetRevocationStatusInput{
		CertificateHashes: []string{
			testCertificate1Hash + testCertificate2Hash,
			testCertificate2Hash + testCertificate2Hash,
		},
		JobArn:             &signingJobArn,
		PlatformId:         aws.String(platformNotation),
		ProfileVersionArn:  &profileVersionArn,
		SignatureTimestamp: request.Signature.CriticalAttributes.AuthenticSigningTime,
	}
}

func toPluginError(err error, t *testing.T) *plugin.Error {
	if err == nil {
		t.Fatal("expected error but not found")
	}
	plgErr, ok := err.(*plugin.Error)
	if !ok {
		t.Errorf("Expected error of type 'plugin.Error' but not found")
	}
	return plgErr
}

func getMockClient(expectedInput *signer.GetRevocationStatusInput, op *signer.GetRevocationStatusOutput, err error, t *testing.T) (client.Interface, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	mockSignerClient := client.NewMockInterface(mockCtrl)
	mockSignerClient.EXPECT().GetRevocationStatus(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, input *signer.GetRevocationStatusInput, optFns ...func(*signer.Options)) (*signer.GetRevocationStatusOutput, error) {
			if expectedInput != nil {
				assert.Equal(t, expectedInput, input, "GetRevocationStatusInput mismatch")
			}
			return op, err
		})

	return mockSignerClient, mockCtrl
}
