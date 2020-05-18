// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// OrgUsernameRoleScope OrgUsernameRoleScope is entity to represent set of role_id, role_name and array of scopes
//
// swagger:model OrgUsernameRoleScope
type OrgUsernameRoleScope struct {

	// metadata
	Metadata *ListMetadata `json:"metadata,omitempty"`

	// userroles
	Userroles *UsernameRoleScope `json:"userroles,omitempty"`
}

// Validate validates this org username role scope
func (m *OrgUsernameRoleScope) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateMetadata(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUserroles(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *OrgUsernameRoleScope) validateMetadata(formats strfmt.Registry) error {

	if swag.IsZero(m.Metadata) { // not required
		return nil
	}

	if m.Metadata != nil {
		if err := m.Metadata.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("metadata")
			}
			return err
		}
	}

	return nil
}

func (m *OrgUsernameRoleScope) validateUserroles(formats strfmt.Registry) error {

	if swag.IsZero(m.Userroles) { // not required
		return nil
	}

	if m.Userroles != nil {
		if err := m.Userroles.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("userroles")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *OrgUsernameRoleScope) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *OrgUsernameRoleScope) UnmarshalBinary(b []byte) error {
	var res OrgUsernameRoleScope
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
