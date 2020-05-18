// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package meta

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

// NewGetMetaParams creates a new GetMetaParams object
// with the default values initialized.
func NewGetMetaParams() *GetMetaParams {

	return &GetMetaParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetMetaParamsWithTimeout creates a new GetMetaParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetMetaParamsWithTimeout(timeout time.Duration) *GetMetaParams {

	return &GetMetaParams{

		timeout: timeout,
	}
}

// NewGetMetaParamsWithContext creates a new GetMetaParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetMetaParamsWithContext(ctx context.Context) *GetMetaParams {

	return &GetMetaParams{

		Context: ctx,
	}
}

// NewGetMetaParamsWithHTTPClient creates a new GetMetaParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetMetaParamsWithHTTPClient(client *http.Client) *GetMetaParams {

	return &GetMetaParams{
		HTTPClient: client,
	}
}

/*GetMetaParams contains all the parameters to send to the API endpoint
for the get meta operation typically these are written to a http.Request
*/
type GetMetaParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get meta params
func (o *GetMetaParams) WithTimeout(timeout time.Duration) *GetMetaParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get meta params
func (o *GetMetaParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get meta params
func (o *GetMetaParams) WithContext(ctx context.Context) *GetMetaParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get meta params
func (o *GetMetaParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get meta params
func (o *GetMetaParams) WithHTTPClient(client *http.Client) *GetMetaParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get meta params
func (o *GetMetaParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *GetMetaParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
