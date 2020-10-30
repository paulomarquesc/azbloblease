// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package config

import (
	"log"
	"os"
)

// Constants
const (
	version              = "1.0.0"
	blobName             = "azblobleaseblob"
	success              = "Success"
	fail                 = "Fail"
	successAlreadyExists = "SuccessAlreadyExists"
	successRenew         = "SuccessOnRenew"
)

// Variables locally and globally scoped
var (
	userAgent  = "azblobleaseclient"                   // UserAgent - add identification to clients
	stdout     = log.New(os.Stdout, "", log.LstdFlags) // Stdout - standard stream output for logs
	stdoutJSON = log.New(os.Stdout, "", 0)             // stdoutJSON - standard output without adding prefixes
	stderr     = log.New(os.Stderr, "", log.LstdFlags) // StdErr - Error stream output for logs

	errorCodes = map[string]int{
		"InvalidErrorCode":                           10,  // Used when an error name passed to GetErrorCode is invalid
		"ErrInvalidArgument":                         100, // Generic invalid argument return code
		"ErrInvalidArgumentMissingResourceGroupName": 110, // Missing resource group name
		"ErrInvalidArgumentMissingAccountName":       120, // Missing storage account name
		"ErrInvalidArgumentMissingContainer":         130, // Missing container name
		"ErrInvalidArgumentInvalidLeaseDuration":     140, // Invalid Lease Duration (needs to be between 15-60)
		"ErrInvalidArgumentMissingLeaseID":           150, // Missing lease ID
		"ErrInvalidArgumentMissingSubscriptionID":    160, // Missing subscription ID
		"ErrAuthorizer":                              300, // Error code related to issues getting authorizer
		"ErrInvalidArgumentIterationsCount":          500, // Iterations cannot be less then 1
		"ErrInvalidArgumentWaitTime":                 501, // Invalid wait time between renew iteration, valid values are between 1 and 59
	}
)

// ErrorCode returns error code based on error name
func ErrorCode(errorName string) int {
	if _, validChoice := errorCodes[errorName]; !validChoice {
		os.Exit(errorCodes["InvalidErrorCode"])
	}

	return errorCodes[errorName]
}

// UserAgent returns the user agent string
func UserAgent() string {
	return userAgent
}

// Stderr returns error stream logger
func Stderr() *log.Logger {
	return stderr
}

// Stdout returns error stream logger
func Stdout() *log.Logger {
	return stderr
}

// StdoutJSON returns stdout stream logger without prefixes
func StdoutJSON() *log.Logger {
	return stdoutJSON
}

// Version returns current utility version
func Version() string {
	return version
}

// BlobName returns the blob name to be used when acquiring lease
func BlobName() string {
	return blobName
}

// Success returns success string
func Success() string {
	return success
}

// SuccessAlreadyExists returns success already exists string
func SuccessAlreadyExists() string {
	return successAlreadyExists
}

// SuccessOnRenew returns success status code on succeeding renewal process
func SuccessOnRenew() string {
	return successRenew
}

// Fail returns fail string
func Fail() string {
	return fail
}
