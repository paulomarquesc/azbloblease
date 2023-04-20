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
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/lease"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/common"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/models"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/utils"
)

// AcquireLease - acquires an Azure blob storage lease
func AcquireLease(cntx context.Context, subscriptionID, resourceGroupName, accountName, container, blobName, environment, cloudConfigFile string, leaseDuration, retries, waittimesec int, cred azcore.TokenCredential) models.ResponseInfo {

	response := models.ResponseInfo{
		SubscriptionID:     &subscriptionID,
		ResourceGroupName:  &resourceGroupName,
		StorageAccountName: &accountName,
		ContainerName:      &container,
		BlobName:           &blobName,
		Status:             to.StringPtr(config.Fail()),
	}

	// Getting storage client and keys
	storageAccountClient, err := common.GetStorageClient(subscriptionID, environment, cloudConfigFile, cred)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while getting storage account client: %v.", err), config.Stderr())
		response.ErrorMessage = to.StringPtr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	accountKeys, err := common.GetStorageAccountKey(cntx, storageAccountClient, resourceGroupName, accountName)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while getting storage account keys: %v.", err), config.Stderr())
		response.ErrorMessage = to.StringPtr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	// Getting blob client
	azBlobClient, err := common.GetBlobClient(cntx, "", accountName, resourceGroupName, *accountKeys.Keys[0].Value, storageAccountClient, cred)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining az blob client: %v", err), config.Stderr())
		response.ErrorMessage = to.StringPtr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	blobRelativePath := fmt.Sprintf("%v/%v", container, blobName)
	blobURL := fmt.Sprintf("%v%v", azBlobClient.URL, blobRelativePath)

	blockBlobClient, err := blockblob.NewClientWithSharedKeyCredential(blobURL, &azBlobClient.SharedKeyCredential, nil)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error occurred trying to create blob client for blob %v, error: %v", blobURL, err), config.Stderr())
		response.ErrorMessage = to.StringPtr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	_, err = blockBlobClient.GetProperties(cntx, nil)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error occurred trying to get blob %v, error: %v", blobURL, err), config.Stderr())
		response.ErrorMessage = to.StringPtr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	// AcquireLease

	// Generating LeaseID
	proposedLeaseID := uuid.New().String()
	for i := 0; i < retries; i++ {

		// Getting lease client
		blobLeaseClient, err := lease.NewBlobClient(blockBlobClient, &lease.BlobClientOptions{
			LeaseID: &proposedLeaseID,
		})

		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while acquiring lease client: %v", err), config.Stderr())
			response.ErrorMessage = to.StringPtr(strings.Replace(err.Error(), "\"", "", -1))
		} else {

			// Acquiring lease
			_, err := blobLeaseClient.AcquireLease(
				cntx,
				int32(leaseDuration),
				&lease.BlobAcquireOptions{},
			)

			if err != nil {
				utils.ConsoleOutput(fmt.Sprintf("an error ocurred while acquiring lease: %v.", err), config.Stderr())
				response.ErrorMessage = to.StringPtr(strings.Replace(err.Error(), "\"", "", -1))
			} else {
				response.ErrorMessage = nil
				break
			}

		}

		time.Sleep(time.Duration(waittimesec) * time.Second)
	}

	if response.ErrorMessage == nil {
		response.Status = to.StringPtr(config.Success())
		response.LeaseID = to.StringPtr(proposedLeaseID)
	}

	return response
}
