## AWS Signer plugin for Notation

[![Build Status](https://github.com/aws/aws-signer-notation-plugin/actions/workflows/build.yml/badge.svg?event=push&branch=main)](https://github.com/aws/aws-signer-notation-plugin/actions/workflows/build.yml?query=workflow%3Abuild+event%3Apush+branch%3Amain)
[![Codecov](https://codecov.io/gh/aws/aws-signer-notation-plugin/branch/main/graph/badge.svg)](https://codecov.io/gh/aws/aws-signer-notation-plugin)
[![Go Reference](https://pkg.go.dev/badge/github.com/aws/aws-signer-notation-plugin.svg)](https://pkg.go.dev/github.com/aws/aws-signer-notation-plugin@main)

The AWS Signer Notation plugin, when used with the [notation](https://github.com/notaryproject/notation) or [notation-go](https://github.com/notaryproject/notation-go), allows you to sign and verify artifacts (such as container images) using AWS Signer. [AWS Signer](https://docs.aws.amazon.com/signer/latest/developerguide/Welcome.html) is a fully managed code-signing service to ensure the trust and integrity of your code. AWS Signer manages the code-signing certificate's public and private keys, and enables central management of the code-signing lifecycle.

The plugin facilitates signing by calling SignPayload API of AWS Signer and supports extended verification by validating trusted identities and verifying the signature revocation status through the GetRevocationStatus API of AWS Signer. The plugin is compliant with the [Notary Specifications](https://github.com/notaryproject/specifications/tree/main).
## Getting Started
To use AWS Signer Notation plugin:

* With Notation CLI - Please refer to the [Notation Container Workflow](https://docs.aws.amazon.com/signer/latest/developerguide/container-workflow.html) for guidance on signing and verifying OCI artifacts.
* With notation-go library - You can use this plugin as a library with notation-go, eliminating the need for invoking plugin executable.

## Building from Source

1. Install go. For more information, refer [Golang documentation](https://golang.org/doc/install).
2. The plugin uses go modules for dependency management. For more information, refer [Go Modules](https://github.com/golang/go/wiki/Modules).
3. Run `make build` to build the AWS Signer Notation plugin.
4. Upon completion of the build process, the plugin executable will be created at `build/bin/notation-com.amazonaws.signer.notation.plugin`.

Now you can use this plugin executable with notation CLI by using the following command:

```
notation plugin install --file ./build/bin/notation-com.amazonaws.signer.notation.plugin
```

### Make Targets
The following targets are available. Each may be run with `make <target>`.

| Make Target        | Description                                                                       |
|:-------------------|:----------------------------------------------------------------------------------|
| `help`             | `help` shows available make targets                                               |
| `build`            | `build` builds the plugin for current environment(e.g. Linux, Darvin and Windows) |
| `generate-mocks`   | generates the mocks required for unit tests                                       |
| `clean`            | `clean` removes build artifacts and auto generated mocks                          |

## Security disclosures
If you think you’ve found a potential security issue, please do not post it in the Issues.  Instead, please follow the instructions [here](https://aws.amazon.com/security/vulnerability-reporting/) or email [AWS security directly](mailto:aws-security@amazon.com).

## License
This project is licensed under the Apache-2.0 License.You can read the license [here](LICENSE)

