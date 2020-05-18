// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package policy

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

// NewDeletePolicyParams creates a new DeletePolicyParams object
// with the default values initialized.
func NewDeletePolicyParams() *DeletePolicyParams {
	var ()
	return &DeletePolicyParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewDeletePolicyParamsWithTimeout creates a new DeletePolicyParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewDeletePolicyParamsWithTimeout(timeout time.Duration) *DeletePolicyParams {
	var ()
	return &DeletePolicyParams{

		timeout: timeout,
	}
}

// NewDeletePolicyParamsWithContext creates a new DeletePolicyParams object
// with the default values initialized, and the ability to set a context for a request
func NewDeletePolicyParamsWithContext(ctx context.Context) *DeletePolicyParams {
	var ()
	return &DeletePolicyParams{

		Context: ctx,
	}
}

// NewDeletePolicyParamsWithHTTPClient creates a new DeletePolicyParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewDeletePolicyParamsWithHTTPClient(client *http.Client) *DeletePolicyParams {
	var ()
	return &DeletePolicyParams{
		HTTPClient: client,
	}
}

/*DeletePolicyParams contains all the parameters to send to the API endpoint
for the delete policy operation typically these are written to a http.Request
*/
type DeletePolicyParams struct {

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
	  The id of the policy.

	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the delete policy params
func (o *DeletePolicyParams) WithTimeout(timeout time.Duration) *DeletePolicyParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete policy params
func (o *DeletePolicyParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the delete policy params
func (o *DeletePolicyParams) WithContext(ctx context.Context) *DeletePolicyParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete policy params
func (o *DeletePolicyParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete policy params
func (o *DeletePolicyParams) WithHTTPClient(client *http.Client) *DeletePolicyParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete policy params
func (o *DeletePolicyParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithEmptyHeader adds the emptyHeader to the delete policy params
func (o *DeletePolicyParams) WithEmptyHeader(emptyHeader string) *DeletePolicyParams {
	o.SetEmptyHeader(emptyHeader)
	return o
}

// SetEmptyHeader adds the emptyHeader to the delete policy params
func (o *DeletePolicyParams) SetEmptyHeader(emptyHeader string) {
	o.EmptyHeader = emptyHeader
}

// WithXEMAIL adds the xEMAIL to the delete policy params
func (o *DeletePolicyParams) WithXEMAIL(xEMAIL *string) *DeletePolicyParams {
	o.SetXEMAIL(xEMAIL)
	return o
}

// SetXEMAIL adds the xEMAIL to the delete policy params
func (o *DeletePolicyParams) SetXEMAIL(xEMAIL *string) {
	o.XEMAIL = xEMAIL
}

// WithXREQUESTID adds the xREQUESTID to the delete policy params
func (o *DeletePolicyParams) WithXREQUESTID(xREQUESTID *string) *DeletePolicyParams {
	o.SetXREQUESTID(xREQUESTID)
	return o
}

// SetXREQUESTID adds the xREQUESTId to the delete policy params
func (o *DeletePolicyParams) SetXREQUESTID(xREQUESTID *string) {
	o.XREQUESTID = xREQUESTID
}

// WithXUSERNAME adds the xUSERNAME to the delete policy params
func (o *DeletePolicyParams) WithXUSERNAME(xUSERNAME *string) *DeletePolicyParams {
	o.SetXUSERNAME(xUSERNAME)
	return o
}

// SetXUSERNAME adds the xUSERNAME to the delete policy params
func (o *DeletePolicyParams) SetXUSERNAME(xUSERNAME *string) {
	o.XUSERNAME = xUSERNAME
}

// WithID adds the id to the delete policy params
func (o *DeletePolicyParams) WithID(id string) *DeletePolicyParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the delete policy params
func (o *DeletePolicyParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *DeletePolicyParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
