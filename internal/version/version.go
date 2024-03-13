// Package version provides utility methods for AWS Signer's plugin version.
package version

// Version value will be set during the build.
var Version = "unknown"

// GetVersion returns the plugin version in Semantic Versioning 2.0.0 format.
func GetVersion() string {
	return Version
}
