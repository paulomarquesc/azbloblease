// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package subcommands

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/common"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/models"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/utils"
)

// CreateLeaseBlob - creates a blob to be used for storage lease process
func CreateLeaseBlob(cntx context.Context, subscriptionID, resourceGroupName, accountName, container, blobName, environment, cloudConfigFile string, cred azcore.TokenCredential) models.ResponseInfo {

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

	// Check if container already exists
	containerClient := azBlobClient.Client.ServiceClient().NewContainerClient(container)

	_, err = containerClient.GetProperties(cntx, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "ContainerNotFound") {
			utils.ConsoleOutput(fmt.Sprintf("an error occurred while checking if container %v exists: %v", container, err), config.Stderr())
			response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
			return response
		}

		// Let's create a new container
		_, err = containerClient.Create(cntx, &azblob.CreateContainerOptions{})
		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error occurred trying to create container %v: %v", container, err), config.Stderr())
			response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
			return response
		}
	}

	blobRelativePath := fmt.Sprintf("%v/%v", container, blobName)
	blobURL := fmt.Sprintf("%v%v", azBlobClient.URL, blobRelativePath)

	blockBlobClient, err := blockblob.NewClient(blobURL, cred, nil)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error occurred trying to create blob client for blob %v, error: %v", blobURL, err), config.Stderr())
		response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error occurred trying to create blob client for blob %v, error: %v", blobURL, err), config.Stderr())
		response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
		return response
	}

	_, err = blockBlobClient.GetProperties(cntx, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "BlobNotFound") {
			utils.ConsoleOutput(fmt.Sprintf("an error occurred while checking if blob %v exists: %v", blobName, err), config.Stderr())
			response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
			return response
		}

		// Perform UploadStream to create new blob for leasing

		// Create some data for the upload stream
		blobSize := 1024 // 1KB
		data := make([]byte, blobSize)
		rand.Read(data)

		_, err = blockBlobClient.UploadStream(cntx, bytes.NewReader(data), &blockblob.UploadStreamOptions{})
		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error occurred while uploading blob stream: %v", err), config.Stderr())
			response.ErrorMessage = to.Ptr(strings.Replace(err.Error(), "\"", "", -1))
			return response
		}
		response.Status = to.Ptr(config.Success())
		return response
	}

	response.Status = to.Ptr(config.SuccessAlreadyExists())
	return response
}
