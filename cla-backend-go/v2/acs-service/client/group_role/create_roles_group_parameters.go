// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package group_role

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

// NewCreateRolesGroupParams creates a new CreateRolesGroupParams object
// with the default values initialized.
func NewCreateRolesGroupParams() *CreateRolesGroupParams {
	var ()
	return &CreateRolesGroupParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewCreateRolesGroupParamsWithTimeout creates a new CreateRolesGroupParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewCreateRolesGroupParamsWithTimeout(timeout time.Duration) *CreateRolesGroupParams {
	var ()
	return &CreateRolesGroupParams{

		timeout: timeout,
	}
}

// NewCreateRolesGroupParamsWithContext creates a new CreateRolesGroupParams object
// with the default values initialized, and the ability to set a context for a request
func NewCreateRolesGroupParamsWithContext(ctx context.Context) *CreateRolesGroupParams {
	var ()
	return &CreateRolesGroupParams{

		Context: ctx,
	}
}

// NewCreateRolesGroupParamsWithHTTPClient creates a new CreateRolesGroupParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewCreateRolesGroupParamsWithHTTPClient(client *http.Client) *CreateRolesGroupParams {
	var ()
	return &CreateRolesGroupParams{
		HTTPClient: client,
	}
}

/*CreateRolesGroupParams contains all the parameters to send to the API endpoint
for the create roles group operation typically these are written to a http.Request
*/
type CreateRolesGroupParams struct {

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
	/*Grant
	  One more roles

	*/
	Grant *models.CreateGroupRole
	/*ID
	  The group id.

	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the create roles group params
func (o *CreateRolesGroupParams) WithTimeout(timeout time.Duration) *CreateRolesGroupParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the create roles group params
func (o *CreateRolesGroupParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the create roles group params
func (o *CreateRolesGroupParams) WithContext(ctx context.Context) *CreateRolesGroupParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the create roles group params
func (o *CreateRolesGroupParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the create roles group params
func (o *CreateRolesGroupParams) WithHTTPClient(client *http.Client) *CreateRolesGroupParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the create roles group params
func (o *CreateRolesGroupParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithEmptyHeader adds the emptyHeader to the create roles group params
func (o *CreateRolesGroupParams) WithEmptyHeader(emptyHeader string) *CreateRolesGroupParams {
	o.SetEmptyHeader(emptyHeader)
	return o
}

// SetEmptyHeader adds the emptyHeader to the create roles group params
func (o *CreateRolesGroupParams) SetEmptyHeader(emptyHeader string) {
	o.EmptyHeader = emptyHeader
}

// WithXEMAIL adds the xEMAIL to the create roles group params
func (o *CreateRolesGroupParams) WithXEMAIL(xEMAIL *string) *CreateRolesGroupParams {
	o.SetXEMAIL(xEMAIL)
	return o
}

// SetXEMAIL adds the xEMAIL to the create roles group params
func (o *CreateRolesGroupParams) SetXEMAIL(xEMAIL *string) {
	o.XEMAIL = xEMAIL
}

// WithXREQUESTID adds the xREQUESTID to the create roles group params
func (o *CreateRolesGroupParams) WithXREQUESTID(xREQUESTID *string) *CreateRolesGroupParams {
	o.SetXREQUESTID(xREQUESTID)
	return o
}

// SetXREQUESTID adds the xREQUESTId to the create roles group params
func (o *CreateRolesGroupParams) SetXREQUESTID(xREQUESTID *string) {
	o.XREQUESTID = xREQUESTID
}

// WithXUSERNAME adds the xUSERNAME to the create roles group params
func (o *CreateRolesGroupParams) WithXUSERNAME(xUSERNAME *string) *CreateRolesGroupParams {
	o.SetXUSERNAME(xUSERNAME)
	return o
}

// SetXUSERNAME adds the xUSERNAME to the create roles group params
func (o *CreateRolesGroupParams) SetXUSERNAME(xUSERNAME *string) {
	o.XUSERNAME = xUSERNAME
}

// WithGrant adds the grant to the create roles group params
func (o *CreateRolesGroupParams) WithGrant(grant *models.CreateGroupRole) *CreateRolesGroupParams {
	o.SetGrant(grant)
	return o
}

// SetGrant adds the grant to the create roles group params
func (o *CreateRolesGroupParams) SetGrant(grant *models.CreateGroupRole) {
	o.Grant = grant
}

// WithID adds the id to the create roles group params
func (o *CreateRolesGroupParams) WithID(id string) *CreateRolesGroupParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the create roles group params
func (o *CreateRolesGroupParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *CreateRolesGroupParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	if o.Grant != nil {
		if err := r.SetBodyParam(o.Grant); err != nil {
			return err
		}
	}

	// path param id
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
