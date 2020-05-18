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
)

// NewDeleteRoleParams creates a new DeleteRoleParams object
// with the default values initialized.
func NewDeleteRoleParams() *DeleteRoleParams {
	var ()
	return &DeleteRoleParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewDeleteRoleParamsWithTimeout creates a new DeleteRoleParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewDeleteRoleParamsWithTimeout(timeout time.Duration) *DeleteRoleParams {
	var ()
	return &DeleteRoleParams{

		timeout: timeout,
	}
}

// NewDeleteRoleParamsWithContext creates a new DeleteRoleParams object
// with the default values initialized, and the ability to set a context for a request
func NewDeleteRoleParamsWithContext(ctx context.Context) *DeleteRoleParams {
	var ()
	return &DeleteRoleParams{

		Context: ctx,
	}
}

// NewDeleteRoleParamsWithHTTPClient creates a new DeleteRoleParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewDeleteRoleParamsWithHTTPClient(client *http.Client) *DeleteRoleParams {
	var ()
	return &DeleteRoleParams{
		HTTPClient: client,
	}
}

/*DeleteRoleParams contains all the parameters to send to the API endpoint
for the delete role operation typically these are written to a http.Request
*/
type DeleteRoleParams struct {

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
	/*ID
	  The role id.

	*/
	ID strfmt.UUID

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the delete role params
func (o *DeleteRoleParams) WithTimeout(timeout time.Duration) *DeleteRoleParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete role params
func (o *DeleteRoleParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the delete role params
func (o *DeleteRoleParams) WithContext(ctx context.Context) *DeleteRoleParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete role params
func (o *DeleteRoleParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete role params
func (o *DeleteRoleParams) WithHTTPClient(client *http.Client) *DeleteRoleParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete role params
func (o *DeleteRoleParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithEmptyHeader adds the emptyHeader to the delete role params
func (o *DeleteRoleParams) WithEmptyHeader(emptyHeader string) *DeleteRoleParams {
	o.SetEmptyHeader(emptyHeader)
	return o
}

// SetEmptyHeader adds the emptyHeader to the delete role params
func (o *DeleteRoleParams) SetEmptyHeader(emptyHeader string) {
	o.EmptyHeader = emptyHeader
}

// WithXEMAIL adds the xEMAIL to the delete role params
func (o *DeleteRoleParams) WithXEMAIL(xEMAIL *string) *DeleteRoleParams {
	o.SetXEMAIL(xEMAIL)
	return o
}

// SetXEMAIL adds the xEMAIL to the delete role params
func (o *DeleteRoleParams) SetXEMAIL(xEMAIL *string) {
	o.XEMAIL = xEMAIL
}

// WithXREQUESTID adds the xREQUESTID to the delete role params
func (o *DeleteRoleParams) WithXREQUESTID(xREQUESTID *string) *DeleteRoleParams {
	o.SetXREQUESTID(xREQUESTID)
	return o
}

// SetXREQUESTID adds the xREQUESTId to the delete role params
func (o *DeleteRoleParams) SetXREQUESTID(xREQUESTID *string) {
	o.XREQUESTID = xREQUESTID
}

// WithXUSERNAME adds the xUSERNAME to the delete role params
func (o *DeleteRoleParams) WithXUSERNAME(xUSERNAME *string) *DeleteRoleParams {
	o.SetXUSERNAME(xUSERNAME)
	return o
}

// SetXUSERNAME adds the xUSERNAME to the delete role params
func (o *DeleteRoleParams) SetXUSERNAME(xUSERNAME *string) {
	o.XUSERNAME = xUSERNAME
}

// WithID adds the id to the delete role params
func (o *DeleteRoleParams) WithID(id strfmt.UUID) *DeleteRoleParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the delete role params
func (o *DeleteRoleParams) SetID(id strfmt.UUID) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *DeleteRoleParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	// path param id
	if err := r.SetPathParam("id", o.ID.String()); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
