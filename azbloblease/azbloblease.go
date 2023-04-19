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
	"github.com/Azure/go-autorest/autorest/to"
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
	createLeaseBlobSubscriptionIDPtr := createLeaseBlobCommand.String("subscriptionid", "", "Subscription where the Storage Account is located")
	createLeaseBlobResourceGroupNamePtr := createLeaseBlobCommand.String("resourcegroupname", "", "Storage Account Resource Group Name")
	createLeaseBlobAccountNamePtr := createLeaseBlobCommand.String("accountname", "", "Storage Account Name")
	createLeaseBlobBlobContainerPtr := createLeaseBlobCommand.String("container", "", "Blob container name")
	createLeaseBlobBlobBlobNamePtr := createLeaseBlobCommand.String("blobname", config.BlobName(), "Blob name")
	createLeaseBlobEnvironmentPtr := createLeaseBlobCommand.String("environment", "AZUREPUBLICCLOUD", fmt.Sprintf("Azure cloud type, currently supported ones are: %v", config.ValidEnvironments()))
	createLeaseBlobManagedIdentityId := createLeaseBlobCommand.String("managed-identity-id", "", "uses user managed identities (accepts resource id or client id)")
	createLeaseBlobUseSystemManagedIdentity := createLeaseBlobCommand.Bool("use-system-managed-identity", false, "uses system managed identity")

	// Acquire subcommand flag pointers
	acquireSubscriptionIDPtr := acquireCommand.String("subscriptionid", "", "Subscription where the Storage Account is located")
	acquireResourceGroupNamePtr := acquireCommand.String("resourcegroupname", "", "Storage Account Resource Group Name")
	acquireAccountNamePtr := acquireCommand.String("accountname", "", "Storage Account Name")
	acquireBlobContainerPtr := acquireCommand.String("container", "", "Blob container name")
	acquireBlobNamePtr := acquireCommand.String("blobname", config.BlobName(), "Blob name")
	acquireLeaseDurationPtr := acquireCommand.Int("leaseduration", 60, "Lease duration in seconds, valid values are between 15 and 60, -1 is not supported in this tool")
	acquireRetriesPtr := acquireCommand.Int("retries", 1, "Lease acquire operation, number of retry attempts")
	acquireWaitTimeSecPtr := acquireCommand.Int("waittimesec", 0, "Time in seconds between iterations to renew current lease, must be between 1 and 59 seconds, ideally half of the time used when acquiring lease")
	acquireEnvironmentPtr := acquireCommand.String("environment", "AZUREPUBLICCLOUD", fmt.Sprintf("Azure cloud type, currently supported ones are: %v", config.ValidEnvironments()))
	acquireManagedIdentityId := acquireCommand.String("managed-identity-id", "", "uses user managed identities (accepts resource id or client id)")
	acquireUseSystemManagedIdentity := acquireCommand.Bool("use-system-managed-identity", false, "uses system managed identity")

	// Renew subcommand flag pointers
	renewSubscriptionIDPtr := renewCommand.String("subscriptionid", "", "Subscription where the Storage Account is located")
	renewResourceGroupNamePtr := renewCommand.String("resourcegroupname", "", "Storage Account Resource Group Name")
	renewAccountNamePtr := renewCommand.String("accountname", "", "Storage Account Name")
	renewBlobContainerPtr := renewCommand.String("container", "", "Blob container name")
	renewBlobNamePtr := renewCommand.String("blobname", config.BlobName(), "Blob name")
	renewLeaseIDPtr := renewCommand.String("leaseid", "", "GUID value that represents the acquired lease")
	renewIterationsPtr := renewCommand.Int("iterations", 20, "Lease renew, number of times it will repeat renew operation")
	renewWaitTimeSecPtr := renewCommand.Int("waittimesec", 30, "Time in seconds between iterations to renew current lease, must be between 1 and 59 seconds, ideally half of the time used when acquiring lease")
	renewEnvironmentPtr := renewCommand.String("environment", "AZUREPUBLICCLOUD", fmt.Sprintf("Azure cloud type, currently supported ones are: %v", config.ValidEnvironments()))
	renewManagedIdentityId := renewCommand.String("managed-identity-id", "", "uses user managed identities (accepts resource id or client id)")
	renewUseSystemManagedIdentity := renewCommand.Bool("use-system-managed-identity", false, "uses system managed identity")

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
		if *createLeaseBlobSubscriptionIDPtr == "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingSubscriptionID")
			return
		}

		if *createLeaseBlobResourceGroupNamePtr == "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingResourceGroupName")
			return
		}

		if *createLeaseBlobAccountNamePtr == "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingAccountName")
			return
		}

		if *createLeaseBlobBlobContainerPtr == "" {
			fmt.Println(createLeaseBlobCommand.Name())
			createLeaseBlobCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingContainer")
			return
		}

		if strings.ToUpper(*createLeaseBlobEnvironmentPtr) != "AZUREPUBLICCLOUD" {
			// Checks if valid cloud environment was passed
			_, found := utils.FindInSlice(config.ValidEnvironments(), strings.ToUpper(*createLeaseBlobEnvironmentPtr))
			if !found {
				fmt.Println(createLeaseBlobCommand.Name())
				createLeaseBlobCommand.PrintDefaults()
				exitCode = config.ErrorCode("ErrInvalidCloudType")
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
			*createLeaseBlobSubscriptionIDPtr,
			*createLeaseBlobResourceGroupNamePtr,
			*createLeaseBlobAccountNamePtr,
			strings.ToLower(*createLeaseBlobBlobContainerPtr),
			*createLeaseBlobBlobBlobNamePtr,
			strings.ToUpper(*createLeaseBlobEnvironmentPtr),
			cred,
		)

		// Outputs json result in stdout
		createLeaseBlobResult.Operation = to.StringPtr(createLeaseBlobCommand.Name())
		utils.ConsoleOutput(
			utils.BuildResultResponse(createLeaseBlobResult),
			config.StdoutJSON(),
		)
	}

	// Acquire subcommand execution
	if acquireCommand.Parsed() {

		// Validations
		if *acquireSubscriptionIDPtr == "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingSubscriptionID")
			return
		}

		if *acquireResourceGroupNamePtr == "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingResourceGroupName")
			return
		}

		if *acquireAccountNamePtr == "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingAccountName")
			return
		}

		if *acquireBlobContainerPtr == "" {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingContainer")
			return
		}

		if *acquireLeaseDurationPtr < 15 || *acquireLeaseDurationPtr > 60 {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentInvalidLeaseDuration")
			return
		}

		if *acquireRetriesPtr < 1 {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentRetryCount")
			return
		}

		if *acquireWaitTimeSecPtr < 0 || *acquireWaitTimeSecPtr > 59 {
			fmt.Println(acquireCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentWaitTimeAcquire")
			return
		}

		if strings.ToUpper(*acquireEnvironmentPtr) != "AZUREPUBLICCLOUD" {
			// Checks if valid cloud environment was passed
			_, found := utils.FindInSlice(config.ValidEnvironments(), strings.ToUpper(*acquireEnvironmentPtr))
			if !found {
				fmt.Println(acquireCommand.Name())
				acquireCommand.PrintDefaults()
				exitCode = config.ErrorCode("ErrInvalidCloudType")
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
			*acquireSubscriptionIDPtr,
			*acquireResourceGroupNamePtr,
			*acquireAccountNamePtr,
			strings.ToLower(*acquireBlobContainerPtr),
			*acquireBlobNamePtr,
			strings.ToUpper(*acquireEnvironmentPtr),
			*acquireLeaseDurationPtr,
			*acquireRetriesPtr,
			*acquireWaitTimeSecPtr,
			cred,
		)

		// Outputs json result in stdout
		acquireResult.Operation = to.StringPtr(acquireCommand.Name())
		utils.ConsoleOutput(
			utils.BuildResultResponse(acquireResult),
			config.StdoutJSON(),
		)
	}

	// Renew subcommand execution
	if renewCommand.Parsed() {

		// Validations
		if *renewSubscriptionIDPtr == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingSubscriptionID")
			return
		}

		if *renewResourceGroupNamePtr == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingResourceGroupName")
			return
		}

		if *renewAccountNamePtr == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingAccountName")
			return
		}

		if *renewBlobContainerPtr == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingContainer")
			return
		}

		if *renewLeaseIDPtr == "" {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingLeaseID")
			return
		}

		if *renewIterationsPtr < 1 {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentIterationsCount")
			return
		}

		if *renewWaitTimeSecPtr < 1 || *renewWaitTimeSecPtr > 59 {
			fmt.Println(renewCommand.Name())
			renewCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentWaitTime")
			return
		}

		if strings.ToUpper(*renewEnvironmentPtr) != "AZUREPUBLICCLOUD" {
			// Checks if valid cloud environment was passed
			_, found := utils.FindInSlice(config.ValidEnvironments(), strings.ToUpper(*renewEnvironmentPtr))
			if !found {
				fmt.Println(renewCommand.Name())
				renewCommand.PrintDefaults()
				exitCode = config.ErrorCode("ErrInvalidCloudType")
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
			*renewSubscriptionIDPtr,
			*renewResourceGroupNamePtr,
			*renewAccountNamePtr,
			strings.ToLower(*renewBlobContainerPtr),
			*renewBlobNamePtr,
			*renewLeaseIDPtr,
			strings.ToUpper(*renewEnvironmentPtr),
			*renewIterationsPtr,
			*renewWaitTimeSecPtr,
			cred,
		)

		// Outputs result into stdout
		renewResult.Operation = to.StringPtr(renewCommand.Name())
		utils.ConsoleOutput(
			utils.BuildResultResponse(renewResult),
			config.StdoutJSON(),
		)
	}
}
