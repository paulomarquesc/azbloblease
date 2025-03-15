// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Tool that is used for leader elections based on Azure Blob Storage blob leases.
// If it successfully acquires a lease it returns the lease id, if not, it will
// return null as lease id and an error message.
//
// Output is a json string in standard output stream with following format:
// {
// 		"operation": "<acquire | renew>",
// 		"leaseID": "<guid>" | null,
// 		"status": "<success | fail>"
// 		"ErrorMessage": "<error message>" | null
// 	}
//
// Notes:
//    - LeaseID and ErrorMessage are mutually exclusive, at any point, when lease id
//      is not empty, ErrorMessage will be null, when lease id is null, ErrorMessage will
//      contain the error description.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/iam"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/subcommands"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/utils"
)

var (
	exitCode = 0
)

func exit(cntx context.Context, exitCode int) {

	if exitCode > 0 {
		os.Exit(exitCode)
	}

}

func main() {
	cntx := context.Background()

	// Cleanup and exit handling
	defer func() { exit(cntx, exitCode); os.Exit(exitCode) }()

	// Flag subcommands
	versionCommand := flag.NewFlagSet("version", flag.ExitOnError)
	createLeaseBlobCommand := flag.NewFlagSet("createleaseblob", flag.ExitOnError)
	acquireCommand := flag.NewFlagSet("acquire", flag.ExitOnError)
	renewCommand := flag.NewFlagSet("renew", flag.ExitOnError)
	// TODO: Implement release command

	// CreateLeaseBlob subcommand flag pointers
	createLeaseBlobSubscriptionID := createLeaseBlobCommand.String("subscriptionid", "", "Subscription where the Storage Account is located")
	createLeaseBlobResourceGroupName := createLeaseBlobCommand.String("resourcegroupname", "", "Storage Account Resource Group Name")
	createLeaseBlobAccountName := createLeaseBlobCommand.String("accountname", "", "Storage Account Name")
	createLeaseBlobBlobContainer := createLeaseBlobCommand.String("container", "", "Blob container name")
	createLeaseBlobBlobBlobName := createLeaseBlobCommand.String("blobname", config.BlobName(), "Blob name")
	createLeaseBlobEnvironment := createLeaseBlobCommand.String("environment", "AZUREPUBLICCLOUD", fmt.Sprintf("Azure cloud type, currently supported ones are: %v", config.ValidEnvironments()))
	createLeaseBlobManagedIdentityId := createLeaseBlobCommand.String("managed-identity-id", "", "uses user managed identities (accepts resource id or client id)")
	createLeaseBlobUseSystemManagedIdentity := createLeaseBlobCommand.Bool("use-system-managed-identity", false, "uses system managed identity")
	createLeaseBlobCustomCloudConfigFile := createLeaseBlobCommand.String("custom-cloudconfig-file", "", "passes a custom cloud configuration to the sdk for use with non-public azure clouds, only used for CUSTOMCLOUD environment")

	// Acquire subcommand flag pointers
	acquireSubscriptionID := acquireCommand.String("subscriptionid", "", "Subscription where the Storage Account is located")
	acquireResourceGroupName := acquireCommand.String("resourcegroupname", "", "Storage Account Resource Group Name")
	acquireAccountName := acquireCommand.String("accountname", "", "Storage Account Name")
	acquireBlobContainer := acquireCommand.String("container", "", "Blob container name")
	acquireBlobName := acquireCommand.String("blobname", config.BlobName(), "Blob name")
	acquireLeaseDuration := acquireCommand.Int("leaseduration", 60, "Lease duration in seconds, valid values are between 15 and 60, -1 is not supported in this tool")
	acquireRetries := acquireCommand.Int("retries", 1, "Lease acquire operation, number of retry attempts")
	acquireWaitTimeSec := acquireCommand.Int("waittimesec", 0, "Time in seconds between iterations to renew current lease, must be between 1 and 59 seconds, ideally half of the time used when acquiring lease")
	acquireEnvironment := acquireCommand.String("environment", "AZUREPUBLICCLOUD", fmt.Sprintf("Azure cloud type, currently supported ones are: %v", config.ValidEnvironments()))
	acquireManagedIdentityId := acquireCommand.String("managed-identity-id", "", "uses user managed identities (accepts resource id or client id)")
	acquireUseSystemManagedIdentity := acquireCommand.Bool("use-system-managed-identity", false, "uses system managed identity")
	acquireCustomCloudConfigFile := acquireCommand.String("custom-cloudconfig-file", "", "passes a custom cloud configuration to the sdk for use with non-public azure clouds, only used for CUSTOMCLOUD environment")

	// Renew subcommand flag pointers
	renewSubscriptionID := renewCommand.String("subscriptionid", "", "Subscription where the Storage Account is located")
	renewResourceGroupName := renewCommand.String("resourcegroupname", "", "Storage Account Resource Group Name")
	renewAccountName := renewCommand.String("accountname", "", "Storage Account Name")
	renewBlobContainer := renewCommand.String("container", "", "Blob container name")
	renewBlobName := renewCommand.String("blobname", config.BlobName(), "Blob name")
	renewLeaseID := renewCommand.String("leaseid", "", "GUID value that represents the acquired lease")
	renewIterations := renewCommand.Int("iterations", 20, "Lease renew, number of times it will repeat renew operation")
	renewWaitTimeSec := renewCommand.Int("waittimesec", 30, "Time in seconds between iterations to renew current lease, must be between 1 and 59 seconds, ideally half of the time used when acquiring lease")
	renewEnvironment := renewCommand.String("environment", "AZUREPUBLICCLOUD", fmt.Sprintf("Azure cloud type, currently supported ones are: %v", config.ValidEnvironments()))
	renewManagedIdentityId := renewCommand.String("managed-identity-id", "", "uses user managed identities (accepts resource id or client id)")
	renewUseSystemManagedIdentity := renewCommand.Bool("use-system-managed-identity", false, "uses system managed identity")
	renewCustomCloudConfigFile := renewCommand.String("custom-cloudconfig-file", "", "passes a custom cloud configuration to the sdk for use with non-public azure clouds, only used for CUSTOMCLOUD environment")

	flag.Parse()

	if len(os.Args) < 2 {
		utils.PrintHeader(fmt.Sprintf("azbloblease - CLI tool to help on leader elections based on Azure Blob Storage blob leasing process - v%v", config.Version()))

		utils.PrintUsage(createLeaseBlobCommand, acquireCommand, renewCommand, versionCommand)

		exitCode = config.ErrorCode("ErrInvalidArgument")
		return
	}

	// Parsing flags based on subcommand
	switch os.Args[1] {
	case "version":
		versionCommand.Parse(os.Args[2:])
	case "createleaseblob":
		createLeaseBlobCommand.Parse(os.Args[2:])
	case "acquire":
		acquireCommand.Parse(os.Args[2:])
	case "renew":
		renewCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		exitCode = config.ErrorCode("ErrInvalidArgument")
		return
	}

	// Executing chosen subcommand

	// Version subcommand execution
	if versionCommand.Parsed() {
		fmt.Println(config.Version())
		exitCode = 0
		return
	}

	// Azure authentication
	var cred azcore.TokenCredential
	var err error

	// CreateLeaseBlob subcommand execution
	if createLeaseBlobCommand.Parsed() {

		// Validations
		if *createLeaseBlobSubscriptionID == "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingSubscriptionID")
			return
		}

		if *createLeaseBlobResourceGroupName == "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingResourceGroupName")
			return
		}

		if *createLeaseBlobAccountName == "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingAccountName")
			return
		}

		if *createLeaseBlobBlobContainer == "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingContainer")
			return
		}

		if strings.ToUpper(*createLeaseBlobEnvironment) != "AZUREPUBLICCLOUD" {
			// Checks if valid cloud environment was passed
			_, found := utils.FindInSlice(config.ValidEnvironments(), strings.ToUpper(*createLeaseBlobEnvironment))
			if !found {
				fmt.Println(createLeaseBlobCommand.Name())
				createLeaseBlobCommand.PrintDefaults()
				exitCode = config.ErrorCode("ErrInvalidCloudType")
				return
			}
		}

		if strings.ToUpper(*createLeaseBlobEnvironment) != "CUSTOMCLOUD" && *createLeaseBlobCustomCloudConfigFile != "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrCloudConfigFileOnlyForCustomCloud")
			return
		}

		if strings.ToUpper(*createLeaseBlobEnvironment) == "CUSTOMCLOUD" && *createLeaseBlobCustomCloudConfigFile == "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrCloudConfigFileRequiredForCustomCloud")
			return
		}

		if strings.ToUpper(*createLeaseBlobEnvironment) == "CUSTOMCLOUD" && *createLeaseBlobCustomCloudConfigFile != "" {
			// Checks if custom cloud config file exists
			if _, err := os.Stat(*createLeaseBlobCustomCloudConfigFile); os.IsNotExist(err) {
				fmt.Println(createLeaseBlobCommand.Name())
				createLeaseBlobCommand.PrintDefaults()
				exitCode = config.ErrorCode("ErrCloudConfigFileNotFound")
				return
			}
		}

		// Azure authentication
		cred, err = iam.GetTokenCredentials(*createLeaseBlobManagedIdentityId, *createLeaseBlobUseSystemManagedIdentity)
		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining token credential: %v", err), config.Stderr())
			exitCode = config.ErrorCode("ErrAuthentication")
			return
		}

		// Run createLeaseBlob
		createLeaseBlobResult := subcommands.CreateLeaseBlob(
			cntx,
			*createLeaseBlobSubscriptionID,
			*createLeaseBlobResourceGroupName,
			*createLeaseBlobAccountName,
			strings.ToLower(*createLeaseBlobBlobContainer),
			*createLeaseBlobBlobBlobName,
			strings.ToUpper(*createLeaseBlobEnvironment),
			*createLeaseBlobCustomCloudConfigFile,
			cred,
		)

		// Outputs json result in stdout
		createLeaseBlobResult.Operation = to.Ptr(createLeaseBlobCommand.Name())
		utils.ConsoleOutput(
			utils.BuildResultResponse(createLeaseBlobResult),
			config.StdoutJSON(),
		)
	}

	// Acquire subcommand execution
	if acquireCommand.Parsed() {

		// Validations
		if *acquireSubscriptionID == "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingSubscriptionID")
			return
		}

		if *acquireResourceGroupName == "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingResourceGroupName")
			return
		}

		if *acquireAccountName == "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingAccountName")
			return
		}

		if *acquireBlobContainer == "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingContainer")
			return
		}

		if *acquireLeaseDuration < 15 || *acquireLeaseDuration > 60 {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentInvalidLeaseDuration")
			return
		}

		if *acquireRetries < 1 {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentRetryCount")
			return
		}

		if *acquireWaitTimeSec < 0 || *acquireWaitTimeSec > 59 {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentWaitTimeAcquire")
			return
		}

		if strings.ToUpper(*acquireEnvironment) != "AZUREPUBLICCLOUD" {
			// Checks if valid cloud environment was passed
			_, found := utils.FindInSlice(config.ValidEnvironments(), strings.ToUpper(*acquireEnvironment))
			if !found {
				fmt.Println(acquireCommand.Name())
				acquireCommand.PrintDefaults()
				exitCode = config.ErrorCode("ErrInvalidCloudType")
				return
			}
		}

		if strings.ToUpper(*acquireEnvironment) != "CUSTOMCLOUD" && *acquireCustomCloudConfigFile != "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrCloudConfigFileOnlyForCustomCloud")
			return
		}

		if strings.ToUpper(*acquireEnvironment) == "CUSTOMCLOUD" && *acquireCustomCloudConfigFile == "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrCloudConfigFileRequiredForCustomCloud")
			return
		}

		if strings.ToUpper(*acquireEnvironment) == "CUSTOMCLOUD" && *acquireCustomCloudConfigFile != "" {
			// Checks if custom cloud config file exists
			if _, err := os.Stat(*acquireCustomCloudConfigFile); os.IsNotExist(err) {
				fmt.Println(acquireCommand.Name())
				acquireCommand.PrintDefaults()
				exitCode = config.ErrorCode("ErrCloudConfigFileNotFound")
				return
			}
		}

		// Azure authentication
		cred, err = iam.GetTokenCredentials(*acquireManagedIdentityId, *acquireUseSystemManagedIdentity)
		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining token credential: %v", err), config.Stderr())
			exitCode = config.ErrorCode("ErrAuthentication")
			return
		}

		// Run acquire
		acquireResult := subcommands.AcquireLease(
			cntx,
			*acquireSubscriptionID,
			*acquireResourceGroupName,
			*acquireAccountName,
			strings.ToLower(*acquireBlobContainer),
			*acquireBlobName,
			strings.ToUpper(*acquireEnvironment),
			*acquireCustomCloudConfigFile,
			*acquireLeaseDuration,
			*acquireRetries,
			*acquireWaitTimeSec,
			cred,
		)

		// Outputs json result in stdout
		acquireResult.Operation = to.Ptr(acquireCommand.Name())
		utils.ConsoleOutput(
			utils.BuildResultResponse(acquireResult),
			config.StdoutJSON(),
		)
	}

	// Renew subcommand execution
	if renewCommand.Parsed() {

		// Validations
		if *renewSubscriptionID == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingSubscriptionID")
			return
		}

		if *renewResourceGroupName == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingResourceGroupName")
			return
		}

		if *renewAccountName == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingAccountName")
			return
		}

		if *renewBlobContainer == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingContainer")
			return
		}

		if *renewLeaseID == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingLeaseID")
			return
		}

		if *renewIterations < 1 {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentIterationsCount")
			return
		}

		if *renewWaitTimeSec < 1 || *renewWaitTimeSec > 59 {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentWaitTime")
			return
		}

		if strings.ToUpper(*renewEnvironment) != "AZUREPUBLICCLOUD" {
			// Checks if valid cloud environment was passed
			_, found := utils.FindInSlice(config.ValidEnvironments(), strings.ToUpper(*renewEnvironment))
			if !found {
				fmt.Println(renewCommand.Name())
				renewCommand.PrintDefaults()
				exitCode = config.ErrorCode("ErrInvalidCloudType")
				return
			}
		}

		if strings.ToUpper(*renewEnvironment) != "CUSTOMCLOUD" && *renewCustomCloudConfigFile != "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrCloudConfigFileOnlyForCustomCloud")
			return
		}

		if strings.ToUpper(*renewEnvironment) == "CUSTOMCLOUD" && *renewCustomCloudConfigFile == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrCloudConfigFileRequiredForCustomCloud")
			return
		}

		if strings.ToUpper(*renewEnvironment) == "CUSTOMCLOUD" && *renewCustomCloudConfigFile != "" {
			// Checks if custom cloud config file exists
			if _, err := os.Stat(*renewCustomCloudConfigFile); os.IsNotExist(err) {
				fmt.Println(renewCommand.Name())
				renewCommand.PrintDefaults()
				exitCode = config.ErrorCode("ErrCloudConfigFileNotFound")
				return
			}
		}

		// Azure authentication
		cred, err = iam.GetTokenCredentials(*renewManagedIdentityId, *renewUseSystemManagedIdentity)
		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining token credential: %v", err), config.Stderr())
			exitCode = config.ErrorCode("ErrAuthentication")
			return
		}

		// Run renew
		renewResult := subcommands.RenewLease(
			cntx,
			*renewSubscriptionID,
			*renewResourceGroupName,
			*renewAccountName,
			strings.ToLower(*renewBlobContainer),
			*renewBlobName,
			*renewLeaseID,
			strings.ToUpper(*renewEnvironment),
			*renewCustomCloudConfigFile,
			*renewIterations,
			*renewWaitTimeSec,
			cred,
		)

		// Outputs result into stdout
		renewResult.Operation = to.Ptr(renewCommand.Name())
		utils.ConsoleOutput(
			utils.BuildResultResponse(renewResult),
			config.StdoutJSON(),
		)
	}
}
