## AWS Signer Plugin for Notation

[![Build Status](https://github.com/aws/aws-signer-notation-plugin/actions/workflows/build.yml/badge.svg?event=push&branch=main)](https://github.com/aws/aws-signer-notation-plugin/actions/workflows/build.yml?query=workflow%3Abuild+event%3Apush+branch%3Amain)
[![Codecov](https://codecov.io/gh/aws/aws-signer-notation-plugin/branch/main/graph/badge.svg)](https://codecov.io/gh/aws/aws-signer-notation-plugin)
[![Go Reference](https://pkg.go.dev/badge/github.com/aws/aws-signer-notation-plugin.svg)](https://pkg.go.dev/github.com/aws/aws-signer-notation-plugin@main)

[Notation](https://github.com/notaryproject/notation) is an open source tool developed by the [Notary Project](https://notaryproject.dev/), which supports signing and verifying container images and other artifacts. The AWS Signer Notation plugin, allows users of Notation ([notation CLI](https://github.com/notaryproject/notation) and [notation-go](https://github.com/notaryproject/notation-go)) to sign and verify artifacts (such as container images) using AWS Signer. [AWS Signer](https://docs.aws.amazon.com/signer/latest/developerguide/Welcome.html) is a fully managed code-signing service to ensure the trust and integrity of your code. AWS Signer manages the code-signing certificates, secures private keys, and manages key rotation without requiring users to take any action.

The plugin is compliant with the [Notary Project specification](https://github.com/notaryproject/specifications/tree/main). It uses the AWS Signer _SignPayload_ API for signing, and _GetRevocationStatus_ API for signature verification.

## Getting Started
To use AWS Signer Notation plugin:

* Notation CLI  - Please refer [AWS Signer documentation](https://docs.aws.amazon.com/signer/latest/developerguide/container-workflow.html) for guidance on signing and verifying OCI artifacts.
* notation-go library -  You can use this plugin as library with notation-go, eliminating the need for invoking plugin executable. Please refer the provided [examples](https://github.com/aws/aws-signer-notation-plugin/tree/main/examples) on how to use plugin as library with notation-go.

## Building from Source

1. Install go. For more information, refer [go documentation](https://golang.org/doc/install).
2. The plugin uses go modules for dependency management. For more information, refer [go modules](https://github.com/golang/go/wiki/Modules).
3. Run `make build` to build the AWS Signer Notation plugin.
4. Upon completion of the build process, the plugin executable will be created at `build/bin/notation-com.amazonaws.signer.notation.plugin`.

Now you can use this plugin executable with notation CLI by using the following command:

`notation plugin install --file ./build/bin/notation-com.amazonaws.signer.notation.plugin`

### Make Targets
The following targets are available. Each may be run with `make <target>`.

| Make Target      | Description                                                                           |
|:-----------------|:--------------------------------------------------------------------------------------|
| `help`           | shows available make targets                                                          |
| `build`          | builds the plugin executable for current environment (e.g. Linux, Darwin and Windows) |
| `test`           | runs all the unit tests using `go test`                                               |
| `generate-mocks` | generates the mocks required for unit tests                                           |
| `clean`          | removes build artifacts and auto generated mocks.                                     |

## Troubleshooting for Local Plugin Builds

If you encounter `gomock` import path errors during `make build`, ensure you're using the correct `mockgen` version and import path:

- Replace all instances of `github.com/golang/mock/gomock` with `go.uber.org/mock/gomock`
- You can use this command:

```bash
find . -name '*.go' -exec sed -i '' 's|github.com/golang/mock/gomock|go.uber.org/mock/gomock|g' {} +

## Security disclosures
To report a potential security issue, please do not create a new Issue in the repository. Instead, please report using the instructions [here](https://aws.amazon.com/security/vulnerability-reporting/) or email [AWS security directly](mailto:aws-security@amazon.com).

## License
This project is licensed under the [Apache-2.0](LICENSE) License.

