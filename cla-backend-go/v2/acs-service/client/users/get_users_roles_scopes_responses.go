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

// GetUsersRolesScopesReader is a Reader for the GetUsersRolesScopes structure.
type GetUsersRolesScopesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetUsersRolesScopesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetUsersRolesScopesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetUsersRolesScopesBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetUsersRolesScopesUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewGetUsersRolesScopesForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetUsersRolesScopesNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetUsersRolesScopesOK creates a GetUsersRolesScopesOK with default headers values
func NewGetUsersRolesScopesOK() *GetUsersRolesScopesOK {
	return &GetUsersRolesScopesOK{}
}

/*GetUsersRolesScopesOK handles this case with default header values.

Success
*/
type GetUsersRolesScopesOK struct {
	Payload *models.UsernameRoleScope
}

func (o *GetUsersRolesScopesOK) Error() string {
	return fmt.Sprintf("[GET /users/rolescopes][%d] getUsersRolesScopesOK  %+v", 200, o.Payload)
}

func (o *GetUsersRolesScopesOK) GetPayload() *models.UsernameRoleScope {
	return o.Payload
}

func (o *GetUsersRolesScopesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.UsernameRoleScope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetUsersRolesScopesBadRequest creates a GetUsersRolesScopesBadRequest with default headers values
func NewGetUsersRolesScopesBadRequest() *GetUsersRolesScopesBadRequest {
	return &GetUsersRolesScopesBadRequest{}
}

/*GetUsersRolesScopesBadRequest handles this case with default header values.

Invalid request
*/
type GetUsersRolesScopesBadRequest struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetUsersRolesScopesBadRequest) Error() string {
	return fmt.Sprintf("[GET /users/rolescopes][%d] getUsersRolesScopesBadRequest  %+v", 400, o.Payload)
}

func (o *GetUsersRolesScopesBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetUsersRolesScopesBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetUsersRolesScopesUnauthorized creates a GetUsersRolesScopesUnauthorized with default headers values
func NewGetUsersRolesScopesUnauthorized() *GetUsersRolesScopesUnauthorized {
	return &GetUsersRolesScopesUnauthorized{}
}

/*GetUsersRolesScopesUnauthorized handles this case with default header values.

Unauthorized
*/
type GetUsersRolesScopesUnauthorized struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetUsersRolesScopesUnauthorized) Error() string {
	return fmt.Sprintf("[GET /users/rolescopes][%d] getUsersRolesScopesUnauthorized  %+v", 401, o.Payload)
}

func (o *GetUsersRolesScopesUnauthorized) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetUsersRolesScopesUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetUsersRolesScopesForbidden creates a GetUsersRolesScopesForbidden with default headers values
func NewGetUsersRolesScopesForbidden() *GetUsersRolesScopesForbidden {
	return &GetUsersRolesScopesForbidden{}
}

/*GetUsersRolesScopesForbidden handles this case with default header values.

Insufficient privilege to execute action
*/
type GetUsersRolesScopesForbidden struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetUsersRolesScopesForbidden) Error() string {
	return fmt.Sprintf("[GET /users/rolescopes][%d] getUsersRolesScopesForbidden  %+v", 403, o.Payload)
}

func (o *GetUsersRolesScopesForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetUsersRolesScopesForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetUsersRolesScopesNotFound creates a GetUsersRolesScopesNotFound with default headers values
func NewGetUsersRolesScopesNotFound() *GetUsersRolesScopesNotFound {
	return &GetUsersRolesScopesNotFound{}
}

/*GetUsersRolesScopesNotFound handles this case with default header values.

Not found
*/
type GetUsersRolesScopesNotFound struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetUsersRolesScopesNotFound) Error() string {
	return fmt.Sprintf("[GET /users/rolescopes][%d] getUsersRolesScopesNotFound  %+v", 404, o.Payload)
}

func (o *GetUsersRolesScopesNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetUsersRolesScopesNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
