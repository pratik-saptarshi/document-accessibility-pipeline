// Package cloudtools verifies and bootstraps the pinned cloud build tools
// listed in a build-tools manifest. Verification is digest-pinned and
// fail-closed; installation proceeds only after every entry verifies.
package cloudtools
