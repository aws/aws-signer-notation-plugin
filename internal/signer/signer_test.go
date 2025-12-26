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

package signer

import (
	"context"
	"fmt"
	nethttp "net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-signer-notation-plugin/internal/client"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/mock/gomock"
	"github.com/notaryproject/notation-plugin-framework-go/plugin"
	"github.com/stretchr/testify/assert"
)

const testProfile = "NotationProfile"
const testPayloadType = "application/vnd.oci.descriptor.v1+json"
const testSigEnvType = "application/jose+json"

var testPayload = []byte("Sign ME")
var testSig = []byte("dummySignature")
var testSigMetadata = map[string]string{"metadatakey": "metadatavalue"}

func TestGenerateEnvelope(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSignerClient := client.NewMockInterface(mockCtrl)

	jobId := "1"
	mockSignerClient.EXPECT().SignPayload(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, input *signer.SignPayloadInput, optFns ...func(*signer.Options)) (*signer.SignPayloadOutput, error) {
			assert.Equal(t, testProfile, *input.ProfileName, "ProfileName mismatch")
			assert.Equal(t, testPayloadType, *input.PayloadFormat, "PayloadFormat name mismatch")
			assert.Equal(t, testPayload, input.Payload, "Payload mismatch")
			output := &signer.SignPayloadOutput{
				Signature: testSig,
				JobId:     &jobId,
				Metadata:  testSigMetadata,
			}
			return output, nil
		})

	response, _ := New(mockSignerClient).GenerateEnvelope(context.TODO(), mockGenerateEnvReq())
	assert.Equal(t, testSigEnvType, response.SignatureEnvelopeType, "SignatureEnvelopeType mismatch")
	assert.Equal(t, testSig, response.SignatureEnvelope, "Signature mismatch")
	assert.Equal(t, map[string]string{"metadatakey": "metadatavalue"}, response.Annotations, "metadata mismatch")
}

func TestGenerateEnvelope_MalformedRequest(t *testing.T) {
	badEnvTypeReq := mockGenerateEnvReq()
	badEnvTypeReq.SignatureEnvelopeType = "badType"

	badContractVersionReq := mockGenerateEnvReq()
	badContractVersionReq.ContractVersion = "2.0"

	badProfileArnReq := mockGenerateEnvReq()
	badProfileArnReq.KeyID = "NotationProfile"

	invalidSigningProfileArnReq := mockGenerateEnvReq()
	invalidSigningProfileArnReq.KeyID = "arn:aws:signer:us-west-2:123:/signing-profiles/name/version/invalid"

	nonNilExpiryReq := mockGenerateEnvReq()
	nonNilExpiryReq.ExpiryDurationInSeconds = 4

	tests := map[string]struct {
		req      *plugin.GenerateEnvelopeRequest
		errorMsg string
	}{
		"badEnvelopeTypeReq": {
			req:      badEnvTypeReq,
			errorMsg: "envelope type \"badType\" is not supported",
		},
		"badContractVersionReq": {
			req:      badContractVersionReq,
			errorMsg: "\"2.0\" is not a supported notary plugin contract version",
		},
		"badProfileArnReq": {
			req:      badProfileArnReq,
			errorMsg: fmt.Sprintf(errorMsgMalformedSigningProfileFmt, "NotationProfile"),
		},
		"invalidSigningProfileArnReq": {
			req:      invalidSigningProfileArnReq,
			errorMsg: fmt.Sprintf(errorMsgMalformedSigningProfileFmt, "arn:aws:signer:us-west-2:123:/signing-profiles/name/version/invalid"),
		},
		"nonNilExpiryReq": {
			req:      nonNilExpiryReq,
			errorMsg: "AWSSigner plugin doesn't support -e (--expiry) argument. Please use signing profile to set signature expiry.",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			awsSigner, _ := client.NewAWSSigner(context.TODO(), map[string]string{})
			_, err := New(awsSigner).GenerateEnvelope(context.TODO(), test.req)
			plgErr := toPluginError(err, t)
			assert.Equal(t, test.errorMsg, plgErr.Message, "error message mismatch")
		})
	}
}

func TestGenerateEnvelope_AWSSignerError(t *testing.T) {
	awsErrMsg := "aws error message"
	tests := map[string]struct {
		err      error
		errorMsg string
		errCode  plugin.ErrorCode
	}{
		"AccessDeniedException": {
			err: &smithy.GenericAPIError{
				Code:    "AccessDeniedException",
				Message: awsErrMsg,
			},
			errorMsg: "Failed to call AWSSigner. Error: aws error message.",
			errCode:  plugin.ErrorCodeAccessDenied,
		},
		"ThrottlingException": {
			err: &smithy.GenericAPIError{
				Code:    "ThrottlingException",
				Message: awsErrMsg,
			},
			errorMsg: "Failed to call AWSSigner. Error: aws error message.",
			errCode:  plugin.ErrorCodeThrottled,
		},
		"ResourceNotFoundException": {
			err: &smithy.GenericAPIError{
				Code:    "ResourceNotFoundException",
				Message: awsErrMsg,
			},
			errorMsg: "Failed to call AWSSigner. Error: aws error message.",
			errCode:  plugin.ErrorCodeValidation,
		},
		"GenericException": {
			err: &smithy.GenericAPIError{
				Code:    "GenericException",
				Message: awsErrMsg,
			},
			errorMsg: "Failed to call AWSSigner. Error: aws error message.",
			errCode:  plugin.ErrorCodeGeneric,
		},
		"HttpError": {
			err: &http.ResponseError{
				ResponseError: &smithyhttp.ResponseError{
					Response: &smithyhttp.Response{
						Response: &nethttp.Response{
							StatusCode: 200,
						},
					},
				},
				RequestID: "123456789",
			},
			errorMsg: "https response error StatusCode: 200, RequestID: 123456789, <nil>",
			errCode:  plugin.ErrorCodeGeneric,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockSignerClient, mockCtrl := getMockErrorClient(test.err, t)
			defer mockCtrl.Finish()

			_, err := New(mockSignerClient).GenerateEnvelope(context.TODO(), mockGenerateEnvReq())
			plgErr := toPluginError(err, t)
			assert.Equal(t, test.errorMsg, plgErr.Message, "Error message mismatch")
			assert.Equal(t, test.errCode, plgErr.ErrCode, "Wrong error code.")
		})
	}
}

func toPluginError(err error, t *testing.T) *plugin.Error {
	if err == nil {
		t.Error("expected error but not found")
	}
	plgErr, ok := err.(*plugin.Error)
	if !ok {
		t.Errorf("Expected error of type 'plugin.Error' but not found")
	}
	return plgErr
}

func mockGenerateEnvReq() *plugin.GenerateEnvelopeRequest {
	return &plugin.GenerateEnvelopeRequest{
		ContractVersion:       plugin.ContractVersion,
		SignatureEnvelopeType: testSigEnvType,
		Payload:               testPayload,
		PayloadType:           testPayloadType,
		KeyID:                 "arn:aws:signer:us-west-2:780792624090:/signing-profiles/NotationProfile",
	}
}

func getMockErrorClient(err error, t *testing.T) (client.Interface, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	mockSignerClient := client.NewMockInterface(mockCtrl)

	mockSignerClient.EXPECT().SignPayload(gomock.Any(), gomock.Any()).Return(nil, err)
	return mockSignerClient, mockCtrl
}
