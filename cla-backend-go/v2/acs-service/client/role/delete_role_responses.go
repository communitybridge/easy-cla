// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package role

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/models"
)

// DeleteRoleReader is a Reader for the DeleteRole structure.
type DeleteRoleReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteRoleReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 204:
		result := NewDeleteRoleNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewDeleteRoleBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewDeleteRoleUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewDeleteRoleForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewDeleteRoleNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewDeleteRoleNoContent creates a DeleteRoleNoContent with default headers values
func NewDeleteRoleNoContent() *DeleteRoleNoContent {
	return &DeleteRoleNoContent{}
}

/*DeleteRoleNoContent handles this case with default header values.

An empty response
*/
type DeleteRoleNoContent struct {
}

func (o *DeleteRoleNoContent) Error() string {
	return fmt.Sprintf("[DELETE /roles/{id}][%d] deleteRoleNoContent ", 204)
}

func (o *DeleteRoleNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteRoleBadRequest creates a DeleteRoleBadRequest with default headers values
func NewDeleteRoleBadRequest() *DeleteRoleBadRequest {
	return &DeleteRoleBadRequest{}
}

/*DeleteRoleBadRequest handles this case with default header values.

Invalid request
*/
type DeleteRoleBadRequest struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *DeleteRoleBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /roles/{id}][%d] deleteRoleBadRequest  %+v", 400, o.Payload)
}

func (o *DeleteRoleBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteRoleBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteRoleUnauthorized creates a DeleteRoleUnauthorized with default headers values
func NewDeleteRoleUnauthorized() *DeleteRoleUnauthorized {
	return &DeleteRoleUnauthorized{}
}

/*DeleteRoleUnauthorized handles this case with default header values.

Unauthorized
*/
type DeleteRoleUnauthorized struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *DeleteRoleUnauthorized) Error() string {
	return fmt.Sprintf("[DELETE /roles/{id}][%d] deleteRoleUnauthorized  %+v", 401, o.Payload)
}

func (o *DeleteRoleUnauthorized) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteRoleUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteRoleForbidden creates a DeleteRoleForbidden with default headers values
func NewDeleteRoleForbidden() *DeleteRoleForbidden {
	return &DeleteRoleForbidden{}
}

/*DeleteRoleForbidden handles this case with default header values.

Insufficient privilege to execute action
*/
type DeleteRoleForbidden struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *DeleteRoleForbidden) Error() string {
	return fmt.Sprintf("[DELETE /roles/{id}][%d] deleteRoleForbidden  %+v", 403, o.Payload)
}

func (o *DeleteRoleForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteRoleForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteRoleNotFound creates a DeleteRoleNotFound with default headers values
func NewDeleteRoleNotFound() *DeleteRoleNotFound {
	return &DeleteRoleNotFound{}
}

/*DeleteRoleNotFound handles this case with default header values.

Not found
*/
type DeleteRoleNotFound struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *DeleteRoleNotFound) Error() string {
	return fmt.Sprintf("[DELETE /roles/{id}][%d] deleteRoleNotFound  %+v", 404, o.Payload)
}

func (o *DeleteRoleNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *DeleteRoleNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
