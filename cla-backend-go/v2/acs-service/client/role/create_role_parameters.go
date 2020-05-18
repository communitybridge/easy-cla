// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package role

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/models"
)

// NewCreateRoleParams creates a new CreateRoleParams object
// with the default values initialized.
func NewCreateRoleParams() *CreateRoleParams {
	var ()
	return &CreateRoleParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewCreateRoleParamsWithTimeout creates a new CreateRoleParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewCreateRoleParamsWithTimeout(timeout time.Duration) *CreateRoleParams {
	var ()
	return &CreateRoleParams{

		timeout: timeout,
	}
}

// NewCreateRoleParamsWithContext creates a new CreateRoleParams object
// with the default values initialized, and the ability to set a context for a request
func NewCreateRoleParamsWithContext(ctx context.Context) *CreateRoleParams {
	var ()
	return &CreateRoleParams{

		Context: ctx,
	}
}

// NewCreateRoleParamsWithHTTPClient creates a new CreateRoleParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewCreateRoleParamsWithHTTPClient(client *http.Client) *CreateRoleParams {
	var ()
	return &CreateRoleParams{
		HTTPClient: client,
	}
}

/*CreateRoleParams contains all the parameters to send to the API endpoint
for the create role operation typically these are written to a http.Request
*/
type CreateRoleParams struct {

	/*EmptyHeader
	  The access control list header value encoded as base64 - assigned by the API Gateway based on user/request permissions

	*/
	EmptyHeader string
	/*XEMAIL
	  Email of the person who is requesting an access

	*/
	XEMAIL *string
	/*XREQUESTID
	  The unique request ID value - assigned/set by the API Gateway based on the login session

	*/
	XREQUESTID *string
	/*XUSERNAME
	  Username of the person who is requesting an access

	*/
	XUSERNAME *string
	/*Role
	  Creates a new role in the ACS.

	*/
	Role *models.RoleCommon

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the create role params
func (o *CreateRoleParams) WithTimeout(timeout time.Duration) *CreateRoleParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the create role params
func (o *CreateRoleParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the create role params
func (o *CreateRoleParams) WithContext(ctx context.Context) *CreateRoleParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the create role params
func (o *CreateRoleParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the create role params
func (o *CreateRoleParams) WithHTTPClient(client *http.Client) *CreateRoleParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the create role params
func (o *CreateRoleParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithEmptyHeader adds the emptyHeader to the create role params
func (o *CreateRoleParams) WithEmptyHeader(emptyHeader string) *CreateRoleParams {
	o.SetEmptyHeader(emptyHeader)
	return o
}

// SetEmptyHeader adds the emptyHeader to the create role params
func (o *CreateRoleParams) SetEmptyHeader(emptyHeader string) {
	o.EmptyHeader = emptyHeader
}

// WithXEMAIL adds the xEMAIL to the create role params
func (o *CreateRoleParams) WithXEMAIL(xEMAIL *string) *CreateRoleParams {
	o.SetXEMAIL(xEMAIL)
	return o
}

// SetXEMAIL adds the xEMAIL to the create role params
func (o *CreateRoleParams) SetXEMAIL(xEMAIL *string) {
	o.XEMAIL = xEMAIL
}

// WithXREQUESTID adds the xREQUESTID to the create role params
func (o *CreateRoleParams) WithXREQUESTID(xREQUESTID *string) *CreateRoleParams {
	o.SetXREQUESTID(xREQUESTID)
	return o
}

// SetXREQUESTID adds the xREQUESTId to the create role params
func (o *CreateRoleParams) SetXREQUESTID(xREQUESTID *string) {
	o.XREQUESTID = xREQUESTID
}

// WithXUSERNAME adds the xUSERNAME to the create role params
func (o *CreateRoleParams) WithXUSERNAME(xUSERNAME *string) *CreateRoleParams {
	o.SetXUSERNAME(xUSERNAME)
	return o
}

// SetXUSERNAME adds the xUSERNAME to the create role params
func (o *CreateRoleParams) SetXUSERNAME(xUSERNAME *string) {
	o.XUSERNAME = xUSERNAME
}

// WithRole adds the role to the create role params
func (o *CreateRoleParams) WithRole(role *models.RoleCommon) *CreateRoleParams {
	o.SetRole(role)
	return o
}

// SetRole adds the role to the create role params
func (o *CreateRoleParams) SetRole(role *models.RoleCommon) {
	o.Role = role
}

// WriteToRequest writes these params to a swagger request
func (o *CreateRoleParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// header param Empty-Header
	if err := r.SetHeaderParam("Empty-Header", o.EmptyHeader); err != nil {
		return err
	}

	if o.XEMAIL != nil {

		// header param X-EMAIL
		if err := r.SetHeaderParam("X-EMAIL", *o.XEMAIL); err != nil {
			return err
		}

	}

	if o.XREQUESTID != nil {

		// header param X-REQUEST-ID
		if err := r.SetHeaderParam("X-REQUEST-ID", *o.XREQUESTID); err != nil {
			return err
		}

	}

	if o.XUSERNAME != nil {

		// header param X-USERNAME
		if err := r.SetHeaderParam("X-USERNAME", *o.XUSERNAME); err != nil {
			return err
		}

	}

	if o.Role != nil {
		if err := r.SetBodyParam(o.Role); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
