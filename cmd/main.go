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
