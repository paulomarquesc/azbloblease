// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package models

import "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

// ResponseInfo object definition
type ResponseInfo struct {
	SubscriptionID     *string `json:"subscriptionId"`
	ResourceGroupName  *string `json:"resourceGroupName"`
	StorageAccountName *string `json:"storageAccountName"`
	ContainerName      *string `json:"containerName"`
	BlobName           *string `json:"blobName"`
	Operation          *string `json:"operation"`
	LeaseID            *string `json:"leaseId"`
	Status             *string `json:"status"`
	ErrorMessage       *string `json:"errorMessage"`
}

// Endpoints object definition
type Endpoints struct {
	ActiveDirectoryAuthorityHost string `json:"activeDirectory"`
	ResourceManagerEndpoint      string `json:"resourceManager"`
	ResourceManagerAudience      string `json:"activeDirectoryResourceId"`
}

// CloudConfigInfo object definition, used to map the output of az cloud show -n <cloud name> -o json
type CloudConfigInfo struct {
	Endpoints Endpoints `json:"endpoints"`
}

// AzBlobClient object definition
type AzBlobClient struct {
	Client              *azblob.Client
	URL                 string
	SharedKeyCredential azblob.SharedKeyCredential
}
