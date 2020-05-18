// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package doc

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

// NewGetDocParams creates a new GetDocParams object
// with the default values initialized.
func NewGetDocParams() *GetDocParams {

	return &GetDocParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetDocParamsWithTimeout creates a new GetDocParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetDocParamsWithTimeout(timeout time.Duration) *GetDocParams {

	return &GetDocParams{

		timeout: timeout,
	}
}

// NewGetDocParamsWithContext creates a new GetDocParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetDocParamsWithContext(ctx context.Context) *GetDocParams {

	return &GetDocParams{

		Context: ctx,
	}
}

// NewGetDocParamsWithHTTPClient creates a new GetDocParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetDocParamsWithHTTPClient(client *http.Client) *GetDocParams {

	return &GetDocParams{
		HTTPClient: client,
	}
}

/*GetDocParams contains all the parameters to send to the API endpoint
for the get doc operation typically these are written to a http.Request
*/
type GetDocParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get doc params
func (o *GetDocParams) WithTimeout(timeout time.Duration) *GetDocParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get doc params
func (o *GetDocParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get doc params
func (o *GetDocParams) WithContext(ctx context.Context) *GetDocParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get doc params
func (o *GetDocParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get doc params
func (o *GetDocParams) WithHTTPClient(client *http.Client) *GetDocParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get doc params
func (o *GetDocParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *GetDocParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
