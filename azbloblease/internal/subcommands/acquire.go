// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// TODO: This sample tool downloads Azure Key Vault Managed certificates
// they can be self-signed certs or certs generated by CAs integrated
// with AKV. The format can be be PKCS12 or PEM.

package subcommands

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/common"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/models"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/utils"
)

// AcquireLease - acquires an Azure blob storage lease
func AcquireLease(cntx context.Context, subscriptionID, resourceGroupName, accountName, container, blobName string, leaseDuration, retries, waittimesec int) models.ResponseInfo {

	//-------------------------------------
	// Operations based on storage mgmt sdk
	//-------------------------------------
	response := models.ResponseInfo{
		SubscriptionID:     &subscriptionID,
		ResourceGroupName:  &resourceGroupName,
		StorageAccountName: &accountName,
		ContainerName:      &container,
		BlobName:           &blobName,
		Status:             to.StringPtr(config.Fail()),
	}

	// Getting storage client
	storageAccountMgmtClient, err := common.GetStorageAccountsClient(subscriptionID)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining storage client/authorizer: %v.", err), config.Stderr())
		response.ErrorMessage = to.StringPtr(err.Error())
		return response
	}

	// Getting Storage Account Key
	accountKey, err := common.GetAccountKey(
		cntx,
		storageAccountMgmtClient,
		resourceGroupName,
		accountName,
	)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while executing GetAccountKey: %v.", err), config.Stderr())
		response.ErrorMessage = to.StringPtr(err.Error())
		return response
	}

	//-----------------------------------
	// Operations based on azblob package
	//-----------------------------------

	// Create a credential object; this is used to access account while using azblob module.
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining azblob credential: %v.", err), config.Stderr())
		response.ErrorMessage = to.StringPtr(err.Error())
		return response
	}

	// Creating azblob request pipeline
	requestPipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	blobEndppointURL, err := url.Parse(
		common.GetAccountBlobEndpoint(cntx, storageAccountMgmtClient, resourceGroupName, accountName),
	)

	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining blob endpoint: %v.", err), config.Stderr())
		response.ErrorMessage = to.StringPtr(err.Error())
		return response
	}

	// Create an ServiceURL object that wraps the service URL and a request pipeline.
	serviceURL := azblob.NewServiceURL(*blobEndppointURL, requestPipeline)

	// Create a URL that references a container in Azure Storage account to create the lease
	// This returns a ContainerURL object that wraps the container's URL and a request pipeline (inherited from serviceURL)
	containerURL := serviceURL.NewContainerURL(container)

	// Create a URL that references the blob used to acquire the lock
	// This returns a BlockBlobURL object that wraps the blob's URL and a request pipeline (inherited from containerURL)
	blobURL := containerURL.NewBlockBlobURL(blobName)

	// AcquireLease

	// Generating LeaseID
	proposedLeaseID := uuid.New()
	for i := 0; i < retries; i++ {
		_, err := blobURL.AcquireLease(
			cntx,
			proposedLeaseID.String(),
			int32(leaseDuration),
			azblob.ModifiedAccessConditions{},
		)

		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while acquiring lease: %v.", err), config.Stderr())
			response.ErrorMessage = to.StringPtr(err.Error())
		} else {
			response.ErrorMessage = nil
			break
		}

		time.Sleep(time.Duration(waittimesec) * time.Second)
	}

	if response.ErrorMessage == nil {
		response.Status = to.StringPtr(config.Success())
		response.LeaseID = to.StringPtr(proposedLeaseID.String())
	}

	return response
}
