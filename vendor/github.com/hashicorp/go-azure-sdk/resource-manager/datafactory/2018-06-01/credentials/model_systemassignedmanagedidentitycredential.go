package credentials

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type SystemAssignedManagedIdentityCredential struct {
	Annotations    *[]interface{}                 `json:"annotations,omitempty"`
	Description    *string                        `json:"description,omitempty"`
	Type           string                         `json:"type"`
	TypeProperties *ManagedIdentityTypeProperties `json:"typeProperties,omitempty"`
}
