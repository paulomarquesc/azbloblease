// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package subcommands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/lease"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/common"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/models"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/utils"
)

// RenewLease - attempts to renew an Azure blob storage lease
func RenewLease(cntx context.Context, subscriptionID, resourceGroupName, accountName, container, blobName, leaseID, environment, cloudConfigFile string, iterations, waittimesec int, cred azcore.TokenCredential) models.ResponseInfo {

	response := models.ResponseInfo{
		SubscriptionID:     &subscriptionID,
		ResourceGroupName:  &resourceGroupName,
		StorageAccountName: &accountName,
		ContainerName:      &container,
		BlobName:           &blobName,
		Status:             to.Ptr(config.Fail()),
	}

	// Getting storage client
	storageAccountClient, err := common.GetStorageClient(subscriptionID, environment, cloudConfigFile, cred)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while getting storage account client: %v.", err), config.Stderr())
		response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	// Getting blob client
	azBlobClient, err := common.GetBlobClient(cntx, storageAccountClient, accountName, resourceGroupName, cred)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining az blob client: %v", err), config.Stderr())
		response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	blobRelativePath := fmt.Sprintf("%v/%v", container, blobName)
	blobURL := fmt.Sprintf("%v%v", azBlobClient.URL, blobRelativePath)

	blockBlobClient, err := blockblob.NewClient(blobURL, cred, nil)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error occurred trying to create blob client for blob %v, error: %v", blobURL, err), config.Stderr())
		response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	_, err = blockBlobClient.GetProperties(cntx, nil)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error occurred trying to get blob %v, error: %v", blobURL, err), config.Stderr())
		response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	// Renew Lease
	for i := 0; i < iterations; i++ {

		// Getting lease client
		blobLeaseClient, err := lease.NewBlobClient(blockBlobClient, &lease.BlobClientOptions{
			LeaseID: &leaseID,
		})

		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while acquiring lease client: %v", err), config.Stderr())
			response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
		} else {

			// Renew lease
			leaseResponse, err := blobLeaseClient.RenewLease(
				cntx,
				&lease.BlobRenewOptions{},
			)

			if err != nil {
				utils.ConsoleOutput(fmt.Sprintf("an error ocurred while renewing lease: %v.", err), config.Stderr())
				response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
				return response
			}

			renewedLeaseID := *leaseResponse.LeaseID
			diagnosticMessage := fmt.Sprintf("Renewed lease %v, iteration %v, request id %v", renewedLeaseID, i, *leaseResponse.RequestID)
			utils.ConsoleOutput(diagnosticMessage, config.Stderr())
		}

		time.Sleep(time.Duration(waittimesec) * time.Second)
	}

	response.Status = to.Ptr(config.SuccessOnRenew())
	return response
}
