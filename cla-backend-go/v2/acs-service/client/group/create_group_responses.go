// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package group

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/models"
)

// CreateGroupReader is a Reader for the CreateGroup structure.
type CreateGroupReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateGroupReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewCreateGroupCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewCreateGroupBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewCreateGroupUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewCreateGroupForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewCreateGroupNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 409:
		result := NewCreateGroupConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewCreateGroupCreated creates a CreateGroupCreated with default headers values
func NewCreateGroupCreated() *CreateGroupCreated {
	return &CreateGroupCreated{}
}

/*CreateGroupCreated handles this case with default header values.

Created
*/
type CreateGroupCreated struct {
	/*Unique HttpRequest ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.Group
}

func (o *CreateGroupCreated) Error() string {
	return fmt.Sprintf("[POST /groups][%d] createGroupCreated  %+v", 201, o.Payload)
}

func (o *CreateGroupCreated) GetPayload() *models.Group {
	return o.Payload
}

func (o *CreateGroupCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.Group)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateGroupBadRequest creates a CreateGroupBadRequest with default headers values
func NewCreateGroupBadRequest() *CreateGroupBadRequest {
	return &CreateGroupBadRequest{}
}

/*CreateGroupBadRequest handles this case with default header values.

Invalid request
*/
type CreateGroupBadRequest struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateGroupBadRequest) Error() string {
	return fmt.Sprintf("[POST /groups][%d] createGroupBadRequest  %+v", 400, o.Payload)
}

func (o *CreateGroupBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateGroupBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateGroupUnauthorized creates a CreateGroupUnauthorized with default headers values
func NewCreateGroupUnauthorized() *CreateGroupUnauthorized {
	return &CreateGroupUnauthorized{}
}

/*CreateGroupUnauthorized handles this case with default header values.

Unauthorized
*/
type CreateGroupUnauthorized struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateGroupUnauthorized) Error() string {
	return fmt.Sprintf("[POST /groups][%d] createGroupUnauthorized  %+v", 401, o.Payload)
}

func (o *CreateGroupUnauthorized) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateGroupUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateGroupForbidden creates a CreateGroupForbidden with default headers values
func NewCreateGroupForbidden() *CreateGroupForbidden {
	return &CreateGroupForbidden{}
}

/*CreateGroupForbidden handles this case with default header values.

Insufficient privilege to execute action
*/
type CreateGroupForbidden struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateGroupForbidden) Error() string {
	return fmt.Sprintf("[POST /groups][%d] createGroupForbidden  %+v", 403, o.Payload)
}

func (o *CreateGroupForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateGroupForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateGroupNotFound creates a CreateGroupNotFound with default headers values
func NewCreateGroupNotFound() *CreateGroupNotFound {
	return &CreateGroupNotFound{}
}

/*CreateGroupNotFound handles this case with default header values.

Not found
*/
type CreateGroupNotFound struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateGroupNotFound) Error() string {
	return fmt.Sprintf("[POST /groups][%d] createGroupNotFound  %+v", 404, o.Payload)
}

func (o *CreateGroupNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateGroupNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateGroupConflict creates a CreateGroupConflict with default headers values
func NewCreateGroupConflict() *CreateGroupConflict {
	return &CreateGroupConflict{}
}

/*CreateGroupConflict handles this case with default header values.

Duplicate Resource
*/
type CreateGroupConflict struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *CreateGroupConflict) Error() string {
	return fmt.Sprintf("[POST /groups][%d] createGroupConflict  %+v", 409, o.Payload)
}

func (o *CreateGroupConflict) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateGroupConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
