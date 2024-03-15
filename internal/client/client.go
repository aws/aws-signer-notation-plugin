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

// Package client creates AWS service like AWS Signer client required by plugin.
package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-signer-notation-plugin/internal/logger"
	"github.com/aws/aws-signer-notation-plugin/internal/version"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/smithy-go/logging"
	"github.com/aws/smithy-go/middleware"
	"github.com/notaryproject/notation-plugin-framework-go/plugin"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
)

const (
	configKeyAwsProfile     = "aws-profile"
	configKeyAwsRegion      = "aws-region"
	configKeySignerEndpoint = "aws-signer-endpoint-url"
)

// NewAWSSigner creates new AWS Signer client from given pluginConfig
func NewAWSSigner(ctx context.Context, pluginConfig map[string]string) (*signer.Client, error) {
	log := logger.GetLogger(ctx)
	log.Debugln("Initializing Signer Client")
	loadOptions := getLoadOptions(ctx, pluginConfig)

	// Use default config for aws credentials
	defaultConfig, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, plugin.NewGenericError(err.Error())
	}
	s, err := signer.NewFromConfig(defaultConfig), nil

	log.Debugln("Initialized Signer Client")
	return s, err
}

func getLoadOptions(ctx context.Context, pluginConfig map[string]string) []func(*config.LoadOptions) error {
	log := logger.GetLogger(ctx)
	var loadOptions []func(*config.LoadOptions) error
	if customEndpoint, ok := pluginConfig[configKeySignerEndpoint]; ok {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == signer.ServiceID && customEndpoint != "" {
				log.Debug("AWS Signer endpoint override: " + customEndpoint)
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           customEndpoint,
					SigningRegion: region,
				}, nil
			}
			// returning EndpointNotFoundError will allow the service to fall back to its default resolution
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})
		loadOptions = append(loadOptions, config.WithEndpointResolverWithOptions(customResolver))
	}

	if region, ok := pluginConfig[configKeyAwsRegion]; ok {
		loadOptions = append(loadOptions, config.WithRegion(region))
		log.Debugf("AWS Signer region override: %s\n", region)
	}

	if credentialProfile, ok := pluginConfig[configKeyAwsProfile]; ok {
		loadOptions = append(loadOptions, config.WithSharedConfigProfile(credentialProfile))
		log.Debugf("AWS Signer credential profile: %s\n", credentialProfile)
	}

	loadOptions = append(loadOptions, config.WithAPIOptions([]func(*middleware.Stack) error{
		awsmiddleware.AddUserAgentKeyValue("aws-signer-caller", "NotationPlugin/"+version.Version),
	}))

	if log.IsDebug() {
		loadOptions = append(loadOptions, config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody))
		loadOptions = append(loadOptions, config.WithLogConfigurationWarnings(true))
		loadOptions = append(loadOptions, config.WithLogger(logging.LoggerFunc(func(_ logging.Classification, format string, v ...interface{}) {
			log.Debugf("AWS call %s\n", fmt.Sprintf(format, v))
		})))
	}

	return loadOptions
}
