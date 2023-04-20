// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Package that provides some general functions.

package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/models"
)

// PrintHeader prints a header message
func PrintHeader(header string) {
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))
}

func PrintUsage(createLeaseBlobCommand, acquireCommand, renewCommand, versionCommand *flag.FlagSet) {
	fmt.Println("")
	fmt.Println("General usage")
	fmt.Println("")
	fmt.Println("\tazbloblease <command> <options>")
	fmt.Println("")

	fmt.Println("List of commands and their options")

	fmt.Println("")
	fmt.Printf(fmt.Sprintf("%v - Creates a blob to be used for the lease process\n", createLeaseBlobCommand.Name()))
	fmt.Println("")
	createLeaseBlobCommand.PrintDefaults()
	fmt.Println("")
	fmt.Println("\tExample")
	fmt.Println("\t\tazbloblease createleaseblob -accountname \"mystorageaccount\" -container \"azbloblease\" -blobname \"myblob\" -resourcegroupname \"my-rg\" -subscriptionid \"11111111-1111-1111-1111-111111111111\"")
	fmt.Println("")
	fmt.Println("\tOutputs")
	fmt.Println("\t\tstdout - json response after createleaseblob process is executed")
	fmt.Println("\t\tstderr - error messages")

	fmt.Println("")
	fmt.Printf(fmt.Sprintf("%v - Acquires a lease\n", acquireCommand.Name()))
	fmt.Println("")
	acquireCommand.PrintDefaults()
	fmt.Println("")
	fmt.Println("\tExample")
	fmt.Println("\t\tazbloblease acquire -accountname \"mystorageaccount\" -container \"azbloblease\" -blobname \"myblob\" -leaseduration 60 -resourcegroupname \"my-rg\" -subscriptionid \"11111111-1111-1111-1111-111111111111\"")
	fmt.Println("")
	fmt.Println("\tOutputs")
	fmt.Println("\t\tstdout - json response after acquire process is executed")
	fmt.Println("\t\tstderr - error messages")

	fmt.Println("")
	fmt.Printf(fmt.Sprintf("%v - Renews a lease for # of iterations based on an interval between\n", renewCommand.Name()))
	fmt.Println("")
	renewCommand.PrintDefaults()
	fmt.Println("")
	fmt.Println("\tExample")
	fmt.Println("\t\tazbloblease renew -accountname \"mystorageaccount\" -container \"azbloblease\" -blobname \"myblob\" -leaseid \"d3d63201-153b-453b-85ef-6c3bee3082f0\" -resourcegroupname \"my-rg\" -subscriptionid \"11111111-1111-1111-1111-111111111111\" -iterations 10 -waittimesec 30")
	fmt.Println("")
	fmt.Println("\tOutputs")
	fmt.Println("\t\tstdout - json response after all renew iteration operations complete")
	fmt.Println("\t\tstderr - diagnostic messages in every iteration and error messages")

	fmt.Println("")
	fmt.Printf(fmt.Sprintf("%v - gets tool version\n", versionCommand.Name()))
	fmt.Println("")
	versionCommand.PrintDefaults()
	fmt.Println("")
	fmt.Println("\tExample")
	fmt.Println("\t\tazbloblease version")
	fmt.Println("")
	fmt.Println("\tOutputs")
	fmt.Println("\t\tstdout - tool version")
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

// ImportCloudConfigJson imports the cloud config json file and returns a struct
func ImportCloudConfigJson(path string) (*models.CloudConfigInfo, error) {
	infoJSON, err := ioutil.ReadFile(path)
	if err != nil {
		ConsoleOutput(fmt.Sprintf("failed to read file: %v", err), config.Stderr())
		return &models.CloudConfigInfo{}, err
	}

	// Converting json to struct
	var info models.CloudConfigInfo
	json.Unmarshal(infoJSON, &info)
	return &info, nil
}
