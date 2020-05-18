// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package action

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

// NewUpdateActionParams creates a new UpdateActionParams object
// with the default values initialized.
func NewUpdateActionParams() *UpdateActionParams {
	var ()
	return &UpdateActionParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewUpdateActionParamsWithTimeout creates a new UpdateActionParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewUpdateActionParamsWithTimeout(timeout time.Duration) *UpdateActionParams {
	var ()
	return &UpdateActionParams{

		timeout: timeout,
	}
}

// NewUpdateActionParamsWithContext creates a new UpdateActionParams object
// with the default values initialized, and the ability to set a context for a request
func NewUpdateActionParamsWithContext(ctx context.Context) *UpdateActionParams {
	var ()
	return &UpdateActionParams{

		Context: ctx,
	}
}

// NewUpdateActionParamsWithHTTPClient creates a new UpdateActionParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewUpdateActionParamsWithHTTPClient(client *http.Client) *UpdateActionParams {
	var ()
	return &UpdateActionParams{
		HTTPClient: client,
	}
}

/*UpdateActionParams contains all the parameters to send to the API endpoint
for the update action operation typically these are written to a http.Request
*/
type UpdateActionParams struct {

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
	/*Action
	  The action properties to update

	*/
	Action *models.UpdateAction
	/*ID
	  The action id.

	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the update action params
func (o *UpdateActionParams) WithTimeout(timeout time.Duration) *UpdateActionParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the update action params
func (o *UpdateActionParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the update action params
func (o *UpdateActionParams) WithContext(ctx context.Context) *UpdateActionParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the update action params
func (o *UpdateActionParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the update action params
func (o *UpdateActionParams) WithHTTPClient(client *http.Client) *UpdateActionParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the update action params
func (o *UpdateActionParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithEmptyHeader adds the emptyHeader to the update action params
func (o *UpdateActionParams) WithEmptyHeader(emptyHeader string) *UpdateActionParams {
	o.SetEmptyHeader(emptyHeader)
	return o
}

// SetEmptyHeader adds the emptyHeader to the update action params
func (o *UpdateActionParams) SetEmptyHeader(emptyHeader string) {
	o.EmptyHeader = emptyHeader
}

// WithXEMAIL adds the xEMAIL to the update action params
func (o *UpdateActionParams) WithXEMAIL(xEMAIL *string) *UpdateActionParams {
	o.SetXEMAIL(xEMAIL)
	return o
}

// SetXEMAIL adds the xEMAIL to the update action params
func (o *UpdateActionParams) SetXEMAIL(xEMAIL *string) {
	o.XEMAIL = xEMAIL
}

// WithXREQUESTID adds the xREQUESTID to the update action params
func (o *UpdateActionParams) WithXREQUESTID(xREQUESTID *string) *UpdateActionParams {
	o.SetXREQUESTID(xREQUESTID)
	return o
}

// SetXREQUESTID adds the xREQUESTId to the update action params
func (o *UpdateActionParams) SetXREQUESTID(xREQUESTID *string) {
	o.XREQUESTID = xREQUESTID
}

// WithXUSERNAME adds the xUSERNAME to the update action params
func (o *UpdateActionParams) WithXUSERNAME(xUSERNAME *string) *UpdateActionParams {
	o.SetXUSERNAME(xUSERNAME)
	return o
}

// SetXUSERNAME adds the xUSERNAME to the update action params
func (o *UpdateActionParams) SetXUSERNAME(xUSERNAME *string) {
	o.XUSERNAME = xUSERNAME
}

// WithAction adds the action to the update action params
func (o *UpdateActionParams) WithAction(action *models.UpdateAction) *UpdateActionParams {
	o.SetAction(action)
	return o
}

// SetAction adds the action to the update action params
func (o *UpdateActionParams) SetAction(action *models.UpdateAction) {
	o.Action = action
}

// WithID adds the id to the update action params
func (o *UpdateActionParams) WithID(id string) *UpdateActionParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the update action params
func (o *UpdateActionParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *UpdateActionParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	if o.Action != nil {
		if err := r.SetBodyParam(o.Action); err != nil {
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
