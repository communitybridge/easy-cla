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
	"github.com/go-openapi/validate"
)

// CreateObjectType Create Object Type
//
// ObjectType is entity to represent data that is needed to create object type
//
// swagger:model CreateObjectType
type CreateObjectType struct {

	// Restricted to alphanum and these special characters: `+=,.@-_. Max Length: 128.
	// Required: true
	// Pattern: ^[\w+\+=,\.@\-_]{0,128}$
	Name *string `json:"name"`
}

// Validate validates this create object type
func (m *CreateObjectType) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateName(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *CreateObjectType) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	if err := validate.Pattern("name", "body", string(*m.Name), `^[\w+\+=,\.@\-_]{0,128}$`); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *CreateObjectType) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CreateObjectType) UnmarshalBinary(b []byte) error {
	var res CreateObjectType
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
