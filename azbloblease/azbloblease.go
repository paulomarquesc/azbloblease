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

	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
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
	acquireCommand := flag.NewFlagSet("acquire", flag.ExitOnError)
	renewCommand := flag.NewFlagSet("renew", flag.ExitOnError)

	// Acquire subcommand flag pointers
	acquireSubscriptionIDPtr := acquireCommand.String("subscriptionid", "", "Subscription where the Storage Account is located")
	acquireResourceGroupNamePtr := acquireCommand.String("resourcegroupname", "", "Storage Account Resource Group Name")
	acquireAccountNamePtr := acquireCommand.String("accountname", "", "Storage Account Name")
	acquireBlobContainerPtr := acquireCommand.String("container", "", "Blob container name")
	acquireLeaseDurationPtr := acquireCommand.Int("leaseduration", 60, "Lease duration in seconds, valid values are between 15 and 60, -1 is not supported in this tool")

	// Renew subcommand flag pointers
	renewSubscriptionIDPtr := renewCommand.String("subscriptionid", "", "Subscription where the Storage Account is located")
	renewResourceGroupNamePtr := renewCommand.String("resourcegroupname", "", "Storage Account Resource Group Name")
	renewAccountNamePtr := renewCommand.String("accountname", "", "Storage Account Name")
	renewBlobContainerPtr := renewCommand.String("container", "", "Blob container name")
	renewLeaseIDPtr := renewCommand.String("leaseid", "", "GUID value that represents the acquired lease")
	renewIterationsPtr := renewCommand.Int("iterations", 20, "Lease acquire operation number of retry attempts")
	renewWaitTimeSecPtr := renewCommand.Int("waittimesec", 30, "Time in seconds between iterations to renew current lease, must be between 1 and 59 seconds, ideally half of the time used when acquiring lease")

	flag.Parse()

	if len(os.Args) < 2 {
		utils.PrintHeader(fmt.Sprintf("azbloblease - CLI tool to help on leader elections based on Azure Blob Storage blob leasing process - v%v", config.Version()))

		fmt.Println("")
		fmt.Println("General usage")
		fmt.Println("")
		fmt.Println("\tazbloblease <command> <options>")
		fmt.Println("")

		fmt.Println("List of commands and their options")

		fmt.Println("")
		fmt.Println(fmt.Sprintf("%v - Acquires a lease", acquireCommand.Name()))
		fmt.Println("")
		acquireCommand.PrintDefaults()
		fmt.Println("")
		fmt.Println("\tExample")
		fmt.Println("\t\tazbloblease acquire -accountname \"mystorageaccount\" -container \"azbloblease\" -leaseduration 60 -resourcegroupname \"my-rg\" -subscriptionid \"11111111-1111-1111-1111-111111111111\"")
		fmt.Println("")
		fmt.Println("\tOutputs on stdout")
		fmt.Println("\t\t{\"operation\": \"acquire\",\"leaseId\": \"d3d63201-153b-453b-85ef-6c3bee3082f0\",\"status\":\"Success\",\"errorMessage\": null}")
		fmt.Println("\t\t{\"operation\": \"acquire\",\"leaseId\": null,\"status\":\"Fail\",\"errorMessage\": \"Error Message here ...\"}")

		fmt.Println("")
		fmt.Println(fmt.Sprintf("%v - Renews a lease for # of iterations based on an interval between", renewCommand.Name()))
		fmt.Println("")
		renewCommand.PrintDefaults()
		fmt.Println("")
		fmt.Println("\tExample")
		fmt.Println("\t\tazbloblease renew -accountname \"mystorageaccount\" -container \"azbloblease\" -leaseid \"d3d63201-153b-453b-85ef-6c3bee3082f0\" -resourcegroupname \"my-rg\" -subscriptionid \"11111111-1111-1111-1111-111111111111\" -iterations 10 -waittimesec 30")
		fmt.Println("")
		fmt.Println("\tOutputs on stdout")
		fmt.Println("\t\t{\"operation\": \"renew\",\"leaseId\": \"d3d63201-153b-453b-85ef-6c3bee3082f0\",\"status\":\"Success\",\"errorMessage\": null}")
		fmt.Println("\t\t{\"operation\": \"renew\",\"leaseId\": null,\"status\":\"Fail\",\"errorMessage\": \"Error Message here ...\"}")
		fmt.Println("")
		fmt.Println("\tOutputs on stderr (for monitoring)")
		fmt.Println("\t\t2020/10/29 10:25:14 Renewed lease d3d63201-153b-453b-85ef-6c3bee3082f0, iteration 0, request id 13a48284-e01e-0030-4718-ae52a9000000")
		fmt.Println("\t\t2020/10/29 10:25:44 Renewed lease d3d63201-153b-453b-85ef-6c3bee3082f0, iteration 1, request id 13a4d711-e01e-0030-0d18-ae52a9000000")
		fmt.Println("\t\t2020/10/29 10:26:15 Renewed lease d3d63201-153b-453b-85ef-6c3bee3082f0, iteration 2, request id 13a53849-e01e-0030-0c18-ae52a9000000")
		fmt.Println("\t\t2020/10/29 10:26:45 Renewed lease d3d63201-153b-453b-85ef-6c3bee3082f0, iteration 3, request id 13a5b15f-e01e-0030-7118-ae52a9000000")

		fmt.Println("")
		fmt.Println(fmt.Sprintf("%v - gets tool version", versionCommand.Name()))
		fmt.Println("")
		versionCommand.PrintDefaults()
		fmt.Println("")
		fmt.Println("\tExample")
		fmt.Println("\t\tazbloblease version")
		fmt.Println("")
		fmt.Println("\tOutputs on stdout")
		fmt.Println("\t\t0.1.0")

		exitCode = config.ErrorCode("ErrInvalidArgument")
		return
	}

	// Parsing flags based on subcommand

	switch os.Args[1] {
	case "version":
		versionCommand.Parse(os.Args[2:])
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

	// Acquire subcommnad execution
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

		// Run acquire
		leaseID, err := subcommands.AcquireLease(
			cntx,
			*acquireSubscriptionIDPtr,
			*acquireResourceGroupNamePtr,
			*acquireAccountNamePtr,
			strings.ToLower(*acquireBlobContainerPtr),
			*acquireLeaseDurationPtr,
		)

		// Outputs json result in stdout
		utils.ConsoleOutput(
			utils.BuildResult(acquireCommand.Name(), leaseID, err),
			config.StdoutJSON(),
		)
	}

	// Renew subcommnad execution
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
			acquireCommand.PrintDefaults()
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
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingContainer")
			return
		}

		if *renewLeaseIDPtr == "" {
			fmt.Println(renewCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentMissingLeaseID")
			return
		}

		if *renewIterationsPtr < 1 {
			fmt.Println(renewCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentIterationsCount")
			return
		}

		if *renewWaitTimeSecPtr < 1 || *renewWaitTimeSecPtr > 59 {
			fmt.Println(renewCommand.Name())
			acquireCommand.PrintDefaults()
			exitCode = config.ErrorCode("ErrInvalidArgumentWaitTime")
			return
		}

		// Run renew
		leaseID, err := subcommands.RenewLease(
			cntx,
			*renewSubscriptionIDPtr,
			*renewResourceGroupNamePtr,
			*renewAccountNamePtr,
			strings.ToLower(*renewBlobContainerPtr),
			*renewLeaseIDPtr,
			*renewIterationsPtr,
			*renewWaitTimeSecPtr,
		)

		// Outputs result into stdout
		utils.ConsoleOutput(
			utils.BuildResult(renewCommand.Name(), leaseID, err),
			config.StdoutJSON(),
		)
	}
}
