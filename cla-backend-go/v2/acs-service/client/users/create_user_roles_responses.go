// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/models"
)

// CreateUserRolesReader is a Reader for the CreateUserRoles structure.
type CreateUserRolesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateUserRolesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewCreateUserRolesCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewCreateUserRolesBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewCreateUserRolesUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewCreateUserRolesForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewCreateUserRolesNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 409:
		result := NewCreateUserRolesConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewCreateUserRolesCreated creates a CreateUserRolesCreated with default headers values
func NewCreateUserRolesCreated() *CreateUserRolesCreated {
	return &CreateUserRolesCreated{}
}

/*CreateUserRolesCreated handles this case with default header values.

Created
*/
type CreateUserRolesCreated struct {
	Payload []*models.UserRoleScope
}

func (o *CreateUserRolesCreated) Error() string {
	return fmt.Sprintf("[POST /users/roles][%d] createUserRolesCreated  %+v", 201, o.Payload)
}

func (o *CreateUserRolesCreated) GetPayload() []*models.UserRoleScope {
	return o.Payload
}

func (o *CreateUserRolesCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateUserRolesBadRequest creates a CreateUserRolesBadRequest with default headers values
func NewCreateUserRolesBadRequest() *CreateUserRolesBadRequest {
	return &CreateUserRolesBadRequest{}
}

/*CreateUserRolesBadRequest handles this case with default header values.

Invalid request
*/
type CreateUserRolesBadRequest struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateUserRolesBadRequest) Error() string {
	return fmt.Sprintf("[POST /users/roles][%d] createUserRolesBadRequest  %+v", 400, o.Payload)
}

func (o *CreateUserRolesBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateUserRolesBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateUserRolesUnauthorized creates a CreateUserRolesUnauthorized with default headers values
func NewCreateUserRolesUnauthorized() *CreateUserRolesUnauthorized {
	return &CreateUserRolesUnauthorized{}
}

/*CreateUserRolesUnauthorized handles this case with default header values.

Unauthorized
*/
type CreateUserRolesUnauthorized struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateUserRolesUnauthorized) Error() string {
	return fmt.Sprintf("[POST /users/roles][%d] createUserRolesUnauthorized  %+v", 401, o.Payload)
}

func (o *CreateUserRolesUnauthorized) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateUserRolesUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateUserRolesForbidden creates a CreateUserRolesForbidden with default headers values
func NewCreateUserRolesForbidden() *CreateUserRolesForbidden {
	return &CreateUserRolesForbidden{}
}

/*CreateUserRolesForbidden handles this case with default header values.

Insufficient privilege to execute action
*/
type CreateUserRolesForbidden struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateUserRolesForbidden) Error() string {
	return fmt.Sprintf("[POST /users/roles][%d] createUserRolesForbidden  %+v", 403, o.Payload)
}

func (o *CreateUserRolesForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateUserRolesForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateUserRolesNotFound creates a CreateUserRolesNotFound with default headers values
func NewCreateUserRolesNotFound() *CreateUserRolesNotFound {
	return &CreateUserRolesNotFound{}
}

/*CreateUserRolesNotFound handles this case with default header values.

Not found
*/
type CreateUserRolesNotFound struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateUserRolesNotFound) Error() string {
	return fmt.Sprintf("[POST /users/roles][%d] createUserRolesNotFound  %+v", 404, o.Payload)
}

func (o *CreateUserRolesNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateUserRolesNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateUserRolesConflict creates a CreateUserRolesConflict with default headers values
func NewCreateUserRolesConflict() *CreateUserRolesConflict {
	return &CreateUserRolesConflict{}
}

/*CreateUserRolesConflict handles this case with default header values.

Duplicate Resource
*/
type CreateUserRolesConflict struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateUserRolesConflict) Error() string {
	return fmt.Sprintf("[POST /users/roles][%d] createUserRolesConflict  %+v", 409, o.Payload)
}

func (o *CreateUserRolesConflict) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateUserRolesConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
