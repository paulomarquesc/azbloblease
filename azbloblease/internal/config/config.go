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
	version              = "2.0.3"
	blobName             = "azblobleaseblob"
	success              = "Success"
	fail                 = "Fail"
	successAlreadyExists = "SuccessAlreadyExists"
	successRenew         = "SuccessOnRenew"
)

// Variables locally and globally scoped
var (
	userAgent         = "azblobleaseclient"                                                                      // UserAgent - add identification to clients
	stdout            = log.New(os.Stdout, "", log.LstdFlags)                                                    // Stdout - standard stream output for logs
	stdoutJSON        = log.New(os.Stdout, "", 0)                                                                // stdoutJSON - standard output without adding prefixes
	stderr            = log.New(os.Stderr, "", log.LstdFlags)                                                    // StdErr - Error stream output for logs
	validEnvironments = []string{"AZUREPUBLICCLOUD", "AZUREUSGOVERNMENTCLOUD", "AZURECHINACLOUD", "CUSTOMCLOUD"} // validEnvironments supported Azure cloud types

	errorCodes = map[string]int{
		"InvalidErrorCode":                           10,  // Used when an error name passed to GetErrorCode is invalid
		"ErrInvalidArgument":                         100, // Generic invalid argument return code
		"ErrInvalidArgumentMissingResourceGroupName": 110, // Missing resource group name
		"ErrInvalidArgumentMissingAccountName":       120, // Missing storage account name
		"ErrInvalidArgumentMissingContainer":         130, // Missing container name
		"ErrInvalidArgumentInvalidLeaseDuration":     140, // Invalid Lease Duration (needs to be between 15-60)
		"ErrInvalidArgumentMissingLeaseID":           150, // Missing lease ID
		"ErrInvalidArgumentMissingSubscriptionID":    160, // Missing subscription ID
		"ErrInvalidCloudType":                        170, // An invalid cloud type was passed
		"ErrCloudConfigFileOnlyForCustomCloud":       180, // Cloud config file is only supported for custom cloud
		"ErrCloudConfigFileNotFound":                 181, // Cloud config file not found
		"ErrCloudConfigFileRequiredForCustomCloud":   182, // Cloud config file is required for custom cloud
		"ErrAuthentication":                          300, // Error code related to issues getting authenticated
		"ErrInvalidArgumentIterationsCount":          500, // Iterations cannot be less then 1
		"ErrInvalidArgumentRetryCount":               510, // Retry count on acquire cannot be less then 1
		"ErrInvalidArgumentWaitTime":                 520, // Invalid wait time between renew iteration, valid values are between 1 and 59 seconds
		"ErrInvalidArgumentWaitTimeAcquire":          530, // Invalid wait time between acquire retry attempt, valid values are between 0 and 59 seconds
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

// ValidEnvironments returns currently supported Azure Cloud types
func ValidEnvironments() []string {
	return validEnvironments
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
