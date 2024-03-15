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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-signer-notation-plugin/internal/logger"
	"github.com/aws/aws-signer-notation-plugin/plugin"

	"github.com/notaryproject/notation-plugin-framework-go/cli"
)

const debugFlag = "AWS_SIGNER_NOTATION_PLUGIN_DEBUG"

func main() {
	awsPlugin := plugin.NewAWSSignerForCLI()
	ctx := context.Background()

	var pluginCli *cli.CLI
	var err error
	if os.Getenv(debugFlag) == "true" {
		log, logErr := logger.New()
		if logErr != nil {
			os.Exit(100)
		}
		defer log.Close()
		ctx = log.UpdateContext(ctx)
		pluginCli, err = cli.NewWithLogger(awsPlugin, log)
	} else {
		pluginCli, err = cli.New(awsPlugin)
	}
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create executable: %v\n", err)
		os.Exit(101)
	}
	pluginCli.Execute(ctx, os.Args)
}
