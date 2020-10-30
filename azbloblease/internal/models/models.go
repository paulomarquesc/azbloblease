// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package models

// ResponseInfo object definition
type ResponseInfo struct {
	Operation    *string `json:"operation"`
	LeaseID      *string `json:"leaseId"`
	Status       *string `json:"status"`
	ErrorMessage *string `json:"errorMessage"`
}
