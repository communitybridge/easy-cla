// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// CreateUpdatePolicy CreateUpdatePolicy is entity to represent data that is required to create or update policy
//
// swagger:model CreateUpdatePolicy
type CreateUpdatePolicy struct {

	// Description of role, Max Length: 250
	Description string `json:"description,omitempty"`

	// Unique ID reference of the policy
	// Read Only: true
	PolicyID string `json:"policy_id,omitempty"`

	// Restricted to alphanum and these special characters: +=,.@-_. Max Length: 128.
	// Pattern: ^[\w+\+=,\.@\-_]{0,128}$
	PolicyName string `json:"policy_name,omitempty"`

	// Policy type name
	PolicyType string `json:"policy_type,omitempty"`

	// Array of statements that need to be added/updated in policy
	Statement []*CreateUpdateStatement `json:"statement"`
}

// Validate validates this create update policy
func (m *CreateUpdatePolicy) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validatePolicyName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatement(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *CreateUpdatePolicy) validatePolicyName(formats strfmt.Registry) error {

	if swag.IsZero(m.PolicyName) { // not required
		return nil
	}

	if err := validate.Pattern("policy_name", "body", string(m.PolicyName), `^[\w+\+=,\.@\-_]{0,128}$`); err != nil {
		return err
	}

	return nil
}

func (m *CreateUpdatePolicy) validateStatement(formats strfmt.Registry) error {

	if swag.IsZero(m.Statement) { // not required
		return nil
	}

	for i := 0; i < len(m.Statement); i++ {
		if swag.IsZero(m.Statement[i]) { // not required
			continue
		}

		if m.Statement[i] != nil {
			if err := m.Statement[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("statement" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *CreateUpdatePolicy) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CreateUpdatePolicy) UnmarshalBinary(b []byte) error {
	var res CreateUpdatePolicy
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
