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

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/config"
	"github.com/paulomarquesc/azbloblease/azbloblease/internal/utils"
)

// GetAuthorizer gets an authorization token
func GetAuthorizer() (autorest.Authorizer, error) {

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while obtaining authorizer: %v.", err), config.Stderr())
		return nil, err
	}

	return authorizer, nil
}
