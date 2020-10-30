// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Package that provides some general functions.

package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/models"
)

// PrintHeader prints a header message
func PrintHeader(header string) {
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))
}

// ConsoleOutput writes to stdout.
func ConsoleOutput(message string, logger *log.Logger) {
	logger.Println(message)
}

// Contains checks if there is a string already in an existing splice of strings
func Contains(array []string, element string) bool {
	for _, e := range array {
		if e == element {
			return true
		}
	}
	return false
}

// FindInSlice returns index greater than -1 and true if item is found
// Code from https://golangcode.com/check-if-element-exists-in-slice/
func FindInSlice(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// BuildResult returns the json formatted result
func BuildResult(operation, leaseID string, err error) string {
	status := config.Success()
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
		status = config.Fail()
	}

	response := models.ResponseInfo{
		Operation:    to.StringPtr(operation),
		LeaseID:      to.StringPtr(leaseID),
		Status:       to.StringPtr(status),
		ErrorMessage: to.StringPtr(errorMessage),
	}

	responseJSON, _ := json.MarshalIndent(response, "", "    ")

	return strings.Replace(string(responseJSON), "\"\"", "null", -1)
}
