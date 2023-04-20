// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Sample package that is used to obtain an authorizer token
// and to return unmarshall the Azure authentication file
// created by az ad sp create create-for-rbac command-line
// into an AzureAuthInfo object.

package iam

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func GetTokenCredentials(managedIdentityId string, useSystemManagedIdentity bool) (azcore.TokenCredential, error) {
	var cred azcore.TokenCredential
	var err error

	if managedIdentityId == "" && !useSystemManagedIdentity {
		cred, err = azidentity.NewDefaultAzureCredential(nil)
	} else if useSystemManagedIdentity {
		fmt.Println("Using NewManagedIdentityCredential")
		cred, err = azidentity.NewManagedIdentityCredential(nil)
	} else if managedIdentityId != "" {
		fmt.Println("Using NewManagedIdentityCredential for user assigned managed identity")
		opts := azidentity.ManagedIdentityCredentialOptions{}

		if strings.Contains(managedIdentityId, "/") {
			opts = azidentity.ManagedIdentityCredentialOptions{
				ID: azidentity.ResourceID(managedIdentityId),
			}
		} else {
			opts = azidentity.ManagedIdentityCredentialOptions{
				ID: azidentity.ClientID(managedIdentityId),
			}
		}

		cred, err = azidentity.NewManagedIdentityCredential(&opts)
	} else {
		return nil, fmt.Errorf("authentication method not supported")
	}

	if err != nil {
		return nil, fmt.Errorf("an error ocurred: %v", err)
	}

	return cred, nil
}
