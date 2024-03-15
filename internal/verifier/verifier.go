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

// Package verifier verified provides functionality to verify signatures generated using AWS Signer
// in accordance with the NotaryProject Plugin contract.
package verifier

import (
	"context"
	"crypto/sha512"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-signer-notation-plugin/internal/client"
	"github.com/aws/aws-signer-notation-plugin/internal/logger"
	"github.com/aws/aws-signer-notation-plugin/internal/slices"

	"github.com/notaryproject/notation-plugin-framework-go/plugin"
)

const (
	wildcardIdentity       = "*"
	errMsgWildcardIdentity = "The AWSSigner plugin does not support wildcard identity in the trust policy."

	attrSigningProfileVersion = "com.amazonaws.signer.signingProfileVersion"
	attrSigningJob            = "com.amazonaws.signer.signingJob"
	signingSchemeAuthority    = "notary.x509.signingAuthority"

	errMsgCertificateParse = "unable to parse certificates in certificate chain."
	errMsgAttributeParse   = "unable to parse attribute %q."

	reasonTrustedIdentityFailure    = "Signature publisher doesn't match any trusted identities."
	reasonTrustedIdentitySuccessFmt = "Signature publisher matched %q trusted identity."
	reasonNotRevoked                = "Signature is not revoked."
	reasonRevokedResourceFmt        = "Resource(s) %s have been revoked."
	reasonRevokedCertificate        = "Certificate(s) have been revoked."

	platformNotation = "Notation-OCI-SHA384-ECDSA"
)

var verificationCapabilities = []plugin.Capability{
	plugin.CapabilityTrustedIdentityVerifier,
	plugin.CapabilityRevocationCheckVerifier}

// Verifier verifies signature generated using AWS Signer.
type Verifier struct {
	awssigner client.Interface
}

// New returns Verifier given an AWS Signer client.
func New(s client.Interface) *Verifier {
	return &Verifier{awssigner: s}
}

// Verify provides extended verification (including trusted-identity and revocation check)
// for signatures generated using AWS Signer.
func (v *Verifier) Verify(ctx context.Context, request *plugin.VerifySignatureRequest) (*plugin.VerifySignatureResponse, error) {
	log := logger.GetLogger(ctx)
	log.Debug("validating VerifySignatureRequest")
	if err := validate(request); err != nil {
		log.Debugf("validate VerifySignatureRequest error :%s", err)
		return nil, err
	}

	response := plugin.VerifySignatureResponse{
		VerificationResults: make(map[plugin.Capability]*plugin.VerificationResult),
	}
	if slices.Contains(request.TrustPolicy.SignatureVerification, plugin.CapabilityTrustedIdentityVerifier) {
		log.Debug("validating trusted identity")
		if err := validateTrustedIdentity(request, &response); err != nil {
			log.Debugf("validate trusted identity error :%v", err)
			return nil, err
		}
		log.Debugf("verification response: %+v\n", response)
	}
	if slices.Contains(request.TrustPolicy.SignatureVerification, plugin.CapabilityRevocationCheckVerifier) {
		log.Debug("validating revocation status")
		if err := v.validateRevocation(ctx, request, &response); err != nil {
			log.Debugf("validate revocation status error :%v", err)
			return nil, err
		}
		log.Debugf("verification response: %+v\n", response)
	}

	// marking both signing-job ARN and signing-profile-version arn as processed attributes here because the plugin should
	// return both of them as processed even if the revocation call was skipped
	response.ProcessedAttributes = slices.AppendIfNotPresent(response.ProcessedAttributes, attrSigningProfileVersion)
	response.ProcessedAttributes = slices.AppendIfNotPresent(response.ProcessedAttributes, attrSigningJob)
	return &response, nil
}

func validate(req *plugin.VerifySignatureRequest) error {
	if req.ContractVersion != plugin.ContractVersion {
		return plugin.NewUnsupportedContractVersionError(req.ContractVersion)
	}

	if slices.Contains(req.TrustPolicy.TrustedIdentities, wildcardIdentity) {
		return plugin.NewValidationError(errMsgWildcardIdentity)
	}

	for _, value := range req.TrustPolicy.SignatureVerification {
		if !pluginCapabilitySupported(value) {
			return plugin.NewValidationErrorf("'%s' is not a supported plugin capability", value)
		}
	}

	critcAttr := req.Signature.CriticalAttributes
	if critcAttr.AuthenticSigningTime.IsZero() {
		return plugin.NewValidationError("missing authenticSigningTime")
	}

	if !strings.EqualFold(critcAttr.SigningScheme, signingSchemeAuthority) {
		return plugin.NewUnsupportedError(fmt.Sprintf("'%s' signing scheme", req.Signature.CriticalAttributes.SigningScheme))
	}

	return nil
}

func validateTrustedIdentity(request *plugin.VerifySignatureRequest, response *plugin.VerifySignatureResponse) error {
	signatureIdentity, err := getValueAsString(request.Signature.CriticalAttributes.ExtendedAttributes, attrSigningProfileVersion)
	if err != nil {
		return err
	}

	var trustedArns []string
	for _, identity := range request.TrustPolicy.TrustedIdentities {
		if _, ok := isSigningProfileArn(identity); ok {
			trustedArns = append(trustedArns, identity)
		}
	}

	result := &plugin.VerificationResult{
		Success: false,
		Reason:  reasonTrustedIdentityFailure,
	}

	var profileMatch bool
	for _, identity := range request.TrustPolicy.TrustedIdentities {
		if arn, ok := isSigningProfileArn(identity); ok {
			s := strings.Split(arn.Resource, "/")
			if len(s) == 3 { // if profile arn
				lastIndex := strings.LastIndex(signatureIdentity, "/")
				if lastIndex != -1 && strings.EqualFold(signatureIdentity[:lastIndex], identity) {
					profileMatch = true
				}
			} else if len(s) == 4 { // if profile version arn
				if strings.EqualFold(signatureIdentity, identity) {
					profileMatch = true
				}
			}
			if profileMatch {
				result.Success = true
				result.Reason = fmt.Sprintf(reasonTrustedIdentitySuccessFmt, identity)
				break
			}
		}
	}

	response.VerificationResults[plugin.CapabilityTrustedIdentityVerifier] = result
	return nil
}

func isSigningProfileArn(s string) (arn.ARN, bool) {
	if a, err := arn.Parse(s); err == nil {
		return a, a.Service == "signer" && strings.HasPrefix(a.Resource, "/signing-profiles/")
	}

	return arn.ARN{}, false
}

func getValueAsString(m map[string]interface{}, k string) (string, error) {
	if val, ok := m[k]; ok {
		if s, ok := val.(string); ok {
			return s, nil
		}
	}

	return "", plugin.NewValidationErrorf(errMsgAttributeParse, k)
}

func (v *Verifier) validateRevocation(ctx context.Context, request *plugin.VerifySignatureRequest, response *plugin.VerifySignatureResponse) error {
	profileVersionArn, err := getValueAsString(request.Signature.CriticalAttributes.ExtendedAttributes, attrSigningProfileVersion)
	if err != nil {
		return err
	}

	jobArn, err := getValueAsString(request.Signature.CriticalAttributes.ExtendedAttributes, attrSigningJob)
	if err != nil {
		return err
	}

	certHashes, err := hashCertificates(request.Signature.CertificateChain)
	if err != nil {
		return plugin.NewValidationError(errMsgCertificateParse)
	}

	input := &signer.GetRevocationStatusInput{
		CertificateHashes:  certHashes,
		JobArn:             aws.String(jobArn),
		PlatformId:         aws.String(platformNotation),
		ProfileVersionArn:  aws.String(profileVersionArn),
		SignatureTimestamp: request.Signature.CriticalAttributes.AuthenticSigningTime,
	}

	result := &plugin.VerificationResult{
		Success: true,
		Reason:  reasonNotRevoked,
	}
	output, err := v.awssigner.GetRevocationStatus(ctx, input)
	if err != nil {
		result.Success = false
		result.Reason = fmt.Sprintf("GetRevocationStatus call failed with error: %+v", err)
	} else {
		if len(output.RevokedEntities) > 0 {
			result.Success = false
			result.Reason = getRevocationResultReason(output.RevokedEntities)
		}
	}

	response.VerificationResults[plugin.CapabilityRevocationCheckVerifier] = result
	return nil
}

func getRevocationResultReason(revokedEntities []string) string {
	var resources string
	var certRevoked bool
	for _, resource := range revokedEntities {
		if strings.HasPrefix(resource, "arn") {
			if resources == "" {
				resources += resource
			} else {
				resources = resources + ", " + resource
			}
		} else {
			certRevoked = true
		}
	}

	var reason string
	if resources != "" {
		reason = fmt.Sprintf(reasonRevokedResourceFmt, resources)
	}
	if certRevoked {
		reason += reasonRevokedCertificate
	}

	return reason
}

func hashCertificates(certStrings [][]byte) ([]string, error) {
	var certHashes []string
	for _, certString := range certStrings {
		// notation always passes cert in DER format
		cert, err := x509.ParseCertificate(certString)
		if err != nil {
			return nil, err
		}

		certHashes = append(certHashes, hashCertificate(*cert))
	}

	for i := range certHashes {
		if i == len(certHashes)-1 {
			certHashes[i] = certHashes[i] + certHashes[i]
		} else {
			certHashes[i] = certHashes[i] + certHashes[i+1]
		}
	}

	return certHashes, nil
}

func hashCertificate(cert x509.Certificate) string {
	h := sha512.New384()
	h.Write(cert.RawTBSCertificate)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func pluginCapabilitySupported(capability plugin.Capability) bool {
	return slices.Contains(verificationCapabilities, capability)
}
