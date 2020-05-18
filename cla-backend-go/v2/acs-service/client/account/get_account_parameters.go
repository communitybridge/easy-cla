// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package account

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

// NewGetAccountParams creates a new GetAccountParams object
// with the default values initialized.
func NewGetAccountParams() *GetAccountParams {
	var ()
	return &GetAccountParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetAccountParamsWithTimeout creates a new GetAccountParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetAccountParamsWithTimeout(timeout time.Duration) *GetAccountParams {
	var ()
	return &GetAccountParams{

		timeout: timeout,
	}
}

// NewGetAccountParamsWithContext creates a new GetAccountParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetAccountParamsWithContext(ctx context.Context) *GetAccountParams {
	var ()
	return &GetAccountParams{

		Context: ctx,
	}
}

// NewGetAccountParamsWithHTTPClient creates a new GetAccountParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetAccountParamsWithHTTPClient(client *http.Client) *GetAccountParams {
	var ()
	return &GetAccountParams{
		HTTPClient: client,
	}
}

/*GetAccountParams contains all the parameters to send to the API endpoint
for the get account operation typically these are written to a http.Request
*/
type GetAccountParams struct {

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
	  The salesforce ID of account, example: 003q000000x1au8AAA

	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get account params
func (o *GetAccountParams) WithTimeout(timeout time.Duration) *GetAccountParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get account params
func (o *GetAccountParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get account params
func (o *GetAccountParams) WithContext(ctx context.Context) *GetAccountParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get account params
func (o *GetAccountParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get account params
func (o *GetAccountParams) WithHTTPClient(client *http.Client) *GetAccountParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get account params
func (o *GetAccountParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithEmptyHeader adds the emptyHeader to the get account params
func (o *GetAccountParams) WithEmptyHeader(emptyHeader string) *GetAccountParams {
	o.SetEmptyHeader(emptyHeader)
	return o
}

// SetEmptyHeader adds the emptyHeader to the get account params
func (o *GetAccountParams) SetEmptyHeader(emptyHeader string) {
	o.EmptyHeader = emptyHeader
}

// WithXEMAIL adds the xEMAIL to the get account params
func (o *GetAccountParams) WithXEMAIL(xEMAIL *string) *GetAccountParams {
	o.SetXEMAIL(xEMAIL)
	return o
}

// SetXEMAIL adds the xEMAIL to the get account params
func (o *GetAccountParams) SetXEMAIL(xEMAIL *string) {
	o.XEMAIL = xEMAIL
}

// WithXREQUESTID adds the xREQUESTID to the get account params
func (o *GetAccountParams) WithXREQUESTID(xREQUESTID *string) *GetAccountParams {
	o.SetXREQUESTID(xREQUESTID)
	return o
}

// SetXREQUESTID adds the xREQUESTId to the get account params
func (o *GetAccountParams) SetXREQUESTID(xREQUESTID *string) {
	o.XREQUESTID = xREQUESTID
}

// WithXUSERNAME adds the xUSERNAME to the get account params
func (o *GetAccountParams) WithXUSERNAME(xUSERNAME *string) *GetAccountParams {
	o.SetXUSERNAME(xUSERNAME)
	return o
}

// SetXUSERNAME adds the xUSERNAME to the get account params
func (o *GetAccountParams) SetXUSERNAME(xUSERNAME *string) {
	o.XUSERNAME = xUSERNAME
}

// WithID adds the id to the get account params
func (o *GetAccountParams) WithID(id string) *GetAccountParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the get account params
func (o *GetAccountParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *GetAccountParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
