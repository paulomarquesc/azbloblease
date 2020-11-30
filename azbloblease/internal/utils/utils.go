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

	"github.com/Azure/go-autorest/autorest/azure"
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

// BuildResultResponse returns the json formatted result
func BuildResultResponse(result models.ResponseInfo) string {
	responseJSON, _ := json.MarshalIndent(result, "", "    ")
	return strings.Replace(string(responseJSON), "\"\"", "null", -1)
}

// Environment returns an `azure.Environment{...}` for the current cloud.
func Environment(CloudType string) *azure.Environment {

	env, err := azure.EnvironmentFromName(CloudType)
	if err != nil {
		panic(fmt.Sprintf(
			"invalid cloud name '%s' specified, cannot continue\n", CloudType))
	}

	return &env
}
