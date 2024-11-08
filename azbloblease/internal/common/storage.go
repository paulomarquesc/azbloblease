// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package common

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/models"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/utils"
)

// GetAccountProperties returns storage account properties
func GetAccountProperties(cntx context.Context, accountsClient armstorage.AccountsClient, resourceGroupName, accountName string) (armstorage.AccountsClientGetPropertiesResponse, error) {

	storageAccountProps, err := accountsClient.GetProperties(
		cntx,
		resourceGroupName,
		accountName,
		nil,
	)

	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while getting storage account properties: %v", err), config.Stderr())
		return armstorage.AccountsClientGetPropertiesResponse{}, err
	}

	return storageAccountProps, nil
}

// GetAccountBlobEndpoint gets the url of the blobendpoint, needed by azblob package
func GetAccountBlobEndpoint(cntx context.Context, accountsClient *armstorage.AccountsClient, resourceGroupName, accountName string) string {
	// Getting Storage Account Properties to identify the blob endpoint
	storageAccountProps, err := GetAccountProperties(
		cntx,
		*accountsClient,
		resourceGroupName,
		accountName,
	)

	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining account properties: %v.", err), config.Stderr())
		return ""
	}

	return *storageAccountProps.Properties.PrimaryEndpoints.Blob
}

// GetStorageClient gets a storage client
func GetStorageClient(subscriptionID, environment, cloudConfigFile string, cred azcore.TokenCredential) (armstorage.AccountsClient, error) {

	// Getting storage client
	cloudConfig := cloud.Configuration{}

	if environment == "AZUREUSGOVERNMENTCLOUD" {
		cloudConfig = cloud.AzureGovernment
	} else if environment == "AZURECHINACLOUD" {
		cloudConfig = cloud.AzureChina
	} else if environment == "CUSTOMCLOUD" {

		// This is the mapping between values expected on cloud.Configuration
		// and the output of az cloud show -n AzureCloud -o json
		//
		// ActiveDirectoryAuthorityHost = endpoints.activeDirectory (e.g."https://login.microsoftonline.us")
		// Endpoint = endpoints.resourceManager (e.g. "https://management.usgovcloudapi.net")
		// Audience = endpoints.activeDirectoryResourceId (e.g. "https://management.core.usgovcloudapi.net")

		if cloudConfigFile != "" {
			cloudInfo, err := utils.ImportCloudConfigJson(cloudConfigFile)
			if err != nil {
				return armstorage.AccountsClient{}, fmt.Errorf("an error ocurred while importing cloud config information from json file: %v", err)
			}

			cloudConfig = cloud.Configuration{
				ActiveDirectoryAuthorityHost: cloudInfo.Endpoints.ActiveDirectoryAuthorityHost,
				Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
					cloud.ResourceManager: {
						Endpoint: cloudInfo.Endpoints.ResourceManagerEndpoint,
						Audience: cloudInfo.Endpoints.ResourceManagerEndpoint,
					},
				},
			}
		}

	} else {
		cloudConfig = cloud.AzurePublic
	}

	options := arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: cloudConfig,
		},
	}

	storageClientFactory, err := armstorage.NewClientFactory(subscriptionID, cred, &options)
	if err != nil {
		return armstorage.AccountsClient{}, fmt.Errorf("an error ocurred while storage account client: %v", err)
	}

	return *storageClientFactory.NewAccountsClient(), nil
}

// GetBlobClient gets a blob client
func GetBlobClient(cntx context.Context, storageAccountClient armstorage.AccountsClient, accountName, resourceGroupName string, cred azcore.TokenCredential) (models.AzBlobClient, error) {
	result := models.AzBlobClient{}

	// Getting blob endpoint
	blobEndppointURL, err := url.Parse(
		GetAccountBlobEndpoint(cntx, &storageAccountClient, resourceGroupName, accountName),
	)

	if err != nil {
		return result, fmt.Errorf("an error ocurred while obtaining blob endpoint url: %v", err)
	}

	url := blobEndppointURL.String()

	// Getting a blob client to be used in container operations
	blobClient, err := azblob.NewClient(url, cred, nil)
	if err != nil {
		return result, fmt.Errorf("an error ocurred while obtaining az blob client: %v", err)
	}

	return models.AzBlobClient{Client: blobClient, URL: url}, nil
}
