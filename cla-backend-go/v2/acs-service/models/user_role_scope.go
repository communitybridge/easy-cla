// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// UserRoleScope User RoleScope
//
// User Role Scope is entity to represent list of relation between user and object that user has access from the role
//
// swagger:model UserRoleScope
type UserRoleScope struct {

	// Unix timestamp when data is created
	// Read Only: true
	CreatedAt int64 `json:"created_at,omitempty"`

	// LFID/Username of entity author/creator
	// Read Only: true
	CreatedBy string `json:"created_by,omitempty"`

	// User's email address.
	Email string `json:"email,omitempty"`

	// User's first name.
	Firstname string `json:"firstname,omitempty"`

	// Unique Identifier for user grant.
	GrantID string `json:"grant_id,omitempty"`

	// User's last name.
	Lastname string `json:"lastname,omitempty"`

	// individual -- user is an individual contributor. member -- user works for the organization which has the valid membership. non-member -- user works for the organization who's membership is expired or doest have a valid membership. staff -- user having the email address e.g user@linuxfoundation.org
	// Enum: [individual member non-member staff]
	Level string `json:"level,omitempty"`

	// Unique Identifier of the object
	ObjectID int64 `json:"object_id,omitempty"`

	// The name of object
	ObjectName string `json:"object_name,omitempty"`

	// Global Identifier for a specific object (e.g. project).
	ObjectTypeID int64 `json:"object_type_id,omitempty"`

	// String name for an object type.
	ObjectTypeName string `json:"object_type_name,omitempty"`

	// Unique Identifier for role.
	RoleID string `json:"role_id,omitempty"`

	// Unique Identifier for scope.
	ScopeID string `json:"scope_id,omitempty"`

	// Unix timestamp when data is last updated
	// Read Only: true
	UpdatedAt int64 `json:"updated_at,omitempty"`

	// Username/LFID of user who updated last
	// Read Only: true
	UpdatedBy string `json:"updated_by,omitempty"`

	// Linux Foundation ID for this user.
	Username string `json:"username,omitempty"`
}

// Validate validates this user role scope
func (m *UserRoleScope) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLevel(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var userRoleScopeTypeLevelPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["individual","member","non-member","staff"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		userRoleScopeTypeLevelPropEnum = append(userRoleScopeTypeLevelPropEnum, v)
	}
}

const (

	// UserRoleScopeLevelIndividual captures enum value "individual"
	UserRoleScopeLevelIndividual string = "individual"

	// UserRoleScopeLevelMember captures enum value "member"
	UserRoleScopeLevelMember string = "member"

	// UserRoleScopeLevelNonMember captures enum value "non-member"
	UserRoleScopeLevelNonMember string = "non-member"

	// UserRoleScopeLevelStaff captures enum value "staff"
	UserRoleScopeLevelStaff string = "staff"
)

// prop value enum
func (m *UserRoleScope) validateLevelEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, userRoleScopeTypeLevelPropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *UserRoleScope) validateLevel(formats strfmt.Registry) error {

	if swag.IsZero(m.Level) { // not required
		return nil
	}

	// value enum
	if err := m.validateLevelEnum("level", "body", m.Level); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *UserRoleScope) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UserRoleScope) UnmarshalBinary(b []byte) error {
	var res UserRoleScope
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
