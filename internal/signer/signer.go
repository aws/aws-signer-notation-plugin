//  Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License"). You may
//  not use this file except in compliance with the License. A copy of the
//  License is located at
//
// 	http://aws.amazon.com/apache2.0
//
//  or in the "license" file accompanying this file. This file is distributed
//  on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language governing
//  permissions and limitations under the License.

// Package signer provides functionality to generate signatures using AWS Signer
// in accordance with the NotaryProject Plugin contract.
package signer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-signer-notation-plugin/internal/client"
	"github.com/aws/aws-signer-notation-plugin/internal/logger"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/smithy-go"

	"github.com/notaryproject/notation-plugin-framework-go/plugin"
)

const (
	mediaTypeJwsEnvelope               = "application/jose+json"
	errorMsgMalformedSigningProfileFmt = "%s is not a valid AWS Signer signing profile or signing profile version ARN."
	errorMSGExpiryPassed               = "AWSSigner plugin doesn't support -e (--expiry) argument. Please use signing profile to set signature expiry."
)

// Signer generates signature generated using AWS Signer.
type Signer struct {
	awssigner client.Interface
}

// New returns Signer given an AWS Signer client.
func New(s client.Interface) *Signer {
	return &Signer{awssigner: s}
}

// GenerateEnvelope generates signature envelope by calling AWS Signer
func (s *Signer) GenerateEnvelope(ctx context.Context, request *plugin.GenerateEnvelopeRequest) (*plugin.GenerateEnvelopeResponse, error) {
	log := logger.GetLogger(ctx)

	log.Debug("validating request")
	if err := validate(request); err != nil {
		return nil, err
	}
	log.Debug("succeeded request validation")

	log.Debug("validating signing profile")
	signingProfileArn, err := arn.Parse(request.KeyID)
	if err != nil {
		return nil, plugin.NewValidationErrorf(errorMsgMalformedSigningProfileFmt, request.KeyID)
	}
	signingProfileName, err := getProfileName(signingProfileArn)
	if err != nil {
		return nil, err
	}
	log.Debug("succeeded signing profile validation")

	log.Debug("calling AWS Signer's SignPayload API")
	input := &signer.SignPayloadInput{
		Payload:       request.Payload,
		ProfileName:   &signingProfileName,
		PayloadFormat: &request.PayloadType,
		ProfileOwner:  &signingProfileArn.AccountID,
	}
	output, err := s.awssigner.SignPayload(ctx, input)
	if err != nil {
		log.Debugf("failed AWS Signer's SignPayload API call with error: %v", err)
		return nil, parseAwsError(err)
	}

	res := &plugin.GenerateEnvelopeResponse{
		SignatureEnvelope:     output.Signature,
		SignatureEnvelopeType: request.SignatureEnvelopeType,
		Annotations:           output.Metadata}
	log.Debugf("succeeded AWS Signer's SignPayload API call. output: %s", res)

	return res, nil
}

func getProfileName(arn arn.ARN) (string, error) {
	//resource name will be in format /signing-profiles/ProfileName
	profileArnParts := strings.Split(arn.Resource, "/")
	if len(profileArnParts) != 3 {
		return "", plugin.NewValidationErrorf(errorMsgMalformedSigningProfileFmt, arn)
	}
	return profileArnParts[2], nil
}

func validate(request *plugin.GenerateEnvelopeRequest) error {
	if request.ExpiryDurationInSeconds != 0 {
		return plugin.NewError(plugin.ErrorCodeValidation, errorMSGExpiryPassed)
	}
	if request.ContractVersion != plugin.ContractVersion {
		return plugin.NewUnsupportedContractVersionError(request.ContractVersion)
	}
	if request.SignatureEnvelopeType != mediaTypeJwsEnvelope {
		return plugin.NewUnsupportedError(fmt.Sprintf("envelope type %q", request.SignatureEnvelopeType))
	}
	return nil
}

// ParseAwsError converts error from SignPayload API to plugin error
func parseAwsError(err error) *plugin.Error {
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		var re *http.ResponseError
		errMsgSuffix := ""
		if errors.As(err, &re) {
			errMsgSuffix = fmt.Sprintf(" RequestID: %s.", re.ServiceRequestID())
		}
		errMsg := fmt.Sprintf("Failed to call AWSSigner. Error: %s.%s", apiError.ErrorMessage(), errMsgSuffix)
		switch apiError.ErrorCode() {
		case "NotFoundException", "ResourceNotFoundException", "ValidationException", "BadRequestException":
			return plugin.NewValidationError(errMsg)
		case "ThrottlingException":
			return plugin.NewError(plugin.ErrorCodeThrottled, errMsg)
		case "AccessDeniedException":
			return plugin.NewError(plugin.ErrorCodeAccessDenied, errMsg)
		default:
			return plugin.NewGenericError(errMsg)
		}
	}
	return plugin.NewGenericError(err.Error())
}
