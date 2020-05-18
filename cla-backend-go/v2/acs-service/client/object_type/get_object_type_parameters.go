// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package object_type

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

// NewGetObjectTypeParams creates a new GetObjectTypeParams object
// with the default values initialized.
func NewGetObjectTypeParams() *GetObjectTypeParams {
	var ()
	return &GetObjectTypeParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetObjectTypeParamsWithTimeout creates a new GetObjectTypeParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetObjectTypeParamsWithTimeout(timeout time.Duration) *GetObjectTypeParams {
	var ()
	return &GetObjectTypeParams{

		timeout: timeout,
	}
}

// NewGetObjectTypeParamsWithContext creates a new GetObjectTypeParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetObjectTypeParamsWithContext(ctx context.Context) *GetObjectTypeParams {
	var ()
	return &GetObjectTypeParams{

		Context: ctx,
	}
}

// NewGetObjectTypeParamsWithHTTPClient creates a new GetObjectTypeParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetObjectTypeParamsWithHTTPClient(client *http.Client) *GetObjectTypeParams {
	var ()
	return &GetObjectTypeParams{
		HTTPClient: client,
	}
}

/*GetObjectTypeParams contains all the parameters to send to the API endpoint
for the get object type operation typically these are written to a http.Request
*/
type GetObjectTypeParams struct {

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
	  The object type id or object type name, example: 2 or organization

	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get object type params
func (o *GetObjectTypeParams) WithTimeout(timeout time.Duration) *GetObjectTypeParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get object type params
func (o *GetObjectTypeParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get object type params
func (o *GetObjectTypeParams) WithContext(ctx context.Context) *GetObjectTypeParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get object type params
func (o *GetObjectTypeParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get object type params
func (o *GetObjectTypeParams) WithHTTPClient(client *http.Client) *GetObjectTypeParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get object type params
func (o *GetObjectTypeParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithEmptyHeader adds the emptyHeader to the get object type params
func (o *GetObjectTypeParams) WithEmptyHeader(emptyHeader string) *GetObjectTypeParams {
	o.SetEmptyHeader(emptyHeader)
	return o
}

// SetEmptyHeader adds the emptyHeader to the get object type params
func (o *GetObjectTypeParams) SetEmptyHeader(emptyHeader string) {
	o.EmptyHeader = emptyHeader
}

// WithXEMAIL adds the xEMAIL to the get object type params
func (o *GetObjectTypeParams) WithXEMAIL(xEMAIL *string) *GetObjectTypeParams {
	o.SetXEMAIL(xEMAIL)
	return o
}

// SetXEMAIL adds the xEMAIL to the get object type params
func (o *GetObjectTypeParams) SetXEMAIL(xEMAIL *string) {
	o.XEMAIL = xEMAIL
}

// WithXREQUESTID adds the xREQUESTID to the get object type params
func (o *GetObjectTypeParams) WithXREQUESTID(xREQUESTID *string) *GetObjectTypeParams {
	o.SetXREQUESTID(xREQUESTID)
	return o
}

// SetXREQUESTID adds the xREQUESTId to the get object type params
func (o *GetObjectTypeParams) SetXREQUESTID(xREQUESTID *string) {
	o.XREQUESTID = xREQUESTID
}

// WithXUSERNAME adds the xUSERNAME to the get object type params
func (o *GetObjectTypeParams) WithXUSERNAME(xUSERNAME *string) *GetObjectTypeParams {
	o.SetXUSERNAME(xUSERNAME)
	return o
}

// SetXUSERNAME adds the xUSERNAME to the get object type params
func (o *GetObjectTypeParams) SetXUSERNAME(xUSERNAME *string) {
	o.XUSERNAME = xUSERNAME
}

// WithID adds the id to the get object type params
func (o *GetObjectTypeParams) WithID(id string) *GetObjectTypeParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the get object type params
func (o *GetObjectTypeParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *GetObjectTypeParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
