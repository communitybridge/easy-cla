// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package group

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
	"github.com/go-openapi/swag"
)

// NewGetGroupUsersParams creates a new GetGroupUsersParams object
// with the default values initialized.
func NewGetGroupUsersParams() *GetGroupUsersParams {
	var (
		limitDefault  = int64(100)
		offsetDefault = int64(0)
	)
	return &GetGroupUsersParams{
		Limit:  &limitDefault,
		Offset: &offsetDefault,

		timeout: cr.DefaultTimeout,
	}
}

// NewGetGroupUsersParamsWithTimeout creates a new GetGroupUsersParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetGroupUsersParamsWithTimeout(timeout time.Duration) *GetGroupUsersParams {
	var (
		limitDefault  = int64(100)
		offsetDefault = int64(0)
	)
	return &GetGroupUsersParams{
		Limit:  &limitDefault,
		Offset: &offsetDefault,

		timeout: timeout,
	}
}

// NewGetGroupUsersParamsWithContext creates a new GetGroupUsersParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetGroupUsersParamsWithContext(ctx context.Context) *GetGroupUsersParams {
	var (
		limitDefault  = int64(100)
		offsetDefault = int64(0)
	)
	return &GetGroupUsersParams{
		Limit:  &limitDefault,
		Offset: &offsetDefault,

		Context: ctx,
	}
}

// NewGetGroupUsersParamsWithHTTPClient creates a new GetGroupUsersParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetGroupUsersParamsWithHTTPClient(client *http.Client) *GetGroupUsersParams {
	var (
		limitDefault  = int64(100)
		offsetDefault = int64(0)
	)
	return &GetGroupUsersParams{
		Limit:      &limitDefault,
		Offset:     &offsetDefault,
		HTTPClient: client,
	}
}

/*GetGroupUsersParams contains all the parameters to send to the API endpoint
for the get group users operation typically these are written to a http.Request
*/
type GetGroupUsersParams struct {

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
	  The group ID

	*/
	ID string
	/*Limit
	  The maximum number of results per page, value must be a positive integer value

	*/
	Limit *int64
	/*Offset
	  The page offset for fetching subsequent pages of results, value must be a non-negative integer value

	*/
	Offset *int64
	/*Search
	  An optional search for limiting query results

	*/
	Search *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get group users params
func (o *GetGroupUsersParams) WithTimeout(timeout time.Duration) *GetGroupUsersParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get group users params
func (o *GetGroupUsersParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get group users params
func (o *GetGroupUsersParams) WithContext(ctx context.Context) *GetGroupUsersParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get group users params
func (o *GetGroupUsersParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get group users params
func (o *GetGroupUsersParams) WithHTTPClient(client *http.Client) *GetGroupUsersParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get group users params
func (o *GetGroupUsersParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithEmptyHeader adds the emptyHeader to the get group users params
func (o *GetGroupUsersParams) WithEmptyHeader(emptyHeader string) *GetGroupUsersParams {
	o.SetEmptyHeader(emptyHeader)
	return o
}

// SetEmptyHeader adds the emptyHeader to the get group users params
func (o *GetGroupUsersParams) SetEmptyHeader(emptyHeader string) {
	o.EmptyHeader = emptyHeader
}

// WithXEMAIL adds the xEMAIL to the get group users params
func (o *GetGroupUsersParams) WithXEMAIL(xEMAIL *string) *GetGroupUsersParams {
	o.SetXEMAIL(xEMAIL)
	return o
}

// SetXEMAIL adds the xEMAIL to the get group users params
func (o *GetGroupUsersParams) SetXEMAIL(xEMAIL *string) {
	o.XEMAIL = xEMAIL
}

// WithXREQUESTID adds the xREQUESTID to the get group users params
func (o *GetGroupUsersParams) WithXREQUESTID(xREQUESTID *string) *GetGroupUsersParams {
	o.SetXREQUESTID(xREQUESTID)
	return o
}

// SetXREQUESTID adds the xREQUESTId to the get group users params
func (o *GetGroupUsersParams) SetXREQUESTID(xREQUESTID *string) {
	o.XREQUESTID = xREQUESTID
}

// WithXUSERNAME adds the xUSERNAME to the get group users params
func (o *GetGroupUsersParams) WithXUSERNAME(xUSERNAME *string) *GetGroupUsersParams {
	o.SetXUSERNAME(xUSERNAME)
	return o
}

// SetXUSERNAME adds the xUSERNAME to the get group users params
func (o *GetGroupUsersParams) SetXUSERNAME(xUSERNAME *string) {
	o.XUSERNAME = xUSERNAME
}

// WithID adds the id to the get group users params
func (o *GetGroupUsersParams) WithID(id string) *GetGroupUsersParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the get group users params
func (o *GetGroupUsersParams) SetID(id string) {
	o.ID = id
}

// WithLimit adds the limit to the get group users params
func (o *GetGroupUsersParams) WithLimit(limit *int64) *GetGroupUsersParams {
	o.SetLimit(limit)
	return o
}

// SetLimit adds the limit to the get group users params
func (o *GetGroupUsersParams) SetLimit(limit *int64) {
	o.Limit = limit
}

// WithOffset adds the offset to the get group users params
func (o *GetGroupUsersParams) WithOffset(offset *int64) *GetGroupUsersParams {
	o.SetOffset(offset)
	return o
}

// SetOffset adds the offset to the get group users params
func (o *GetGroupUsersParams) SetOffset(offset *int64) {
	o.Offset = offset
}

// WithSearch adds the search to the get group users params
func (o *GetGroupUsersParams) WithSearch(search *string) *GetGroupUsersParams {
	o.SetSearch(search)
	return o
}

// SetSearch adds the search to the get group users params
func (o *GetGroupUsersParams) SetSearch(search *string) {
	o.Search = search
}

// WriteToRequest writes these params to a swagger request
func (o *GetGroupUsersParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	if o.Limit != nil {

		// query param limit
		var qrLimit int64
		if o.Limit != nil {
			qrLimit = *o.Limit
		}
		qLimit := swag.FormatInt64(qrLimit)
		if qLimit != "" {
			if err := r.SetQueryParam("limit", qLimit); err != nil {
				return err
			}
		}

	}

	if o.Offset != nil {

		// query param offset
		var qrOffset int64
		if o.Offset != nil {
			qrOffset = *o.Offset
		}
		qOffset := swag.FormatInt64(qrOffset)
		if qOffset != "" {
			if err := r.SetQueryParam("offset", qOffset); err != nil {
				return err
			}
		}

	}

	if o.Search != nil {

		// query param search
		var qrSearch string
		if o.Search != nil {
			qrSearch = *o.Search
		}
		qSearch := qrSearch
		if qSearch != "" {
			if err := r.SetQueryParam("search", qSearch); err != nil {
				return err
			}
		}

	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
