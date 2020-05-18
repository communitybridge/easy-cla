// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package object_type

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/models"
)

// GetObjectTypeReader is a Reader for the GetObjectType structure.
type GetObjectTypeReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetObjectTypeReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetObjectTypeOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetObjectTypeBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetObjectTypeUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewGetObjectTypeForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetObjectTypeNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetObjectTypeOK creates a GetObjectTypeOK with default headers values
func NewGetObjectTypeOK() *GetObjectTypeOK {
	return &GetObjectTypeOK{}
}

/*GetObjectTypeOK handles this case with default header values.

Success
*/
type GetObjectTypeOK struct {
	Payload *models.ObjectType
}

func (o *GetObjectTypeOK) Error() string {
	return fmt.Sprintf("[GET /object-types/{id}][%d] getObjectTypeOK  %+v", 200, o.Payload)
}

func (o *GetObjectTypeOK) GetPayload() *models.ObjectType {
	return o.Payload
}

func (o *GetObjectTypeOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ObjectType)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetObjectTypeBadRequest creates a GetObjectTypeBadRequest with default headers values
func NewGetObjectTypeBadRequest() *GetObjectTypeBadRequest {
	return &GetObjectTypeBadRequest{}
}

/*GetObjectTypeBadRequest handles this case with default header values.

Invalid request
*/
type GetObjectTypeBadRequest struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetObjectTypeBadRequest) Error() string {
	return fmt.Sprintf("[GET /object-types/{id}][%d] getObjectTypeBadRequest  %+v", 400, o.Payload)
}

func (o *GetObjectTypeBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetObjectTypeBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetObjectTypeUnauthorized creates a GetObjectTypeUnauthorized with default headers values
func NewGetObjectTypeUnauthorized() *GetObjectTypeUnauthorized {
	return &GetObjectTypeUnauthorized{}
}

/*GetObjectTypeUnauthorized handles this case with default header values.

Unauthorized
*/
type GetObjectTypeUnauthorized struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetObjectTypeUnauthorized) Error() string {
	return fmt.Sprintf("[GET /object-types/{id}][%d] getObjectTypeUnauthorized  %+v", 401, o.Payload)
}

func (o *GetObjectTypeUnauthorized) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetObjectTypeUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetObjectTypeForbidden creates a GetObjectTypeForbidden with default headers values
func NewGetObjectTypeForbidden() *GetObjectTypeForbidden {
	return &GetObjectTypeForbidden{}
}

/*GetObjectTypeForbidden handles this case with default header values.

Insufficient privilege to execute action
*/
type GetObjectTypeForbidden struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetObjectTypeForbidden) Error() string {
	return fmt.Sprintf("[GET /object-types/{id}][%d] getObjectTypeForbidden  %+v", 403, o.Payload)
}

func (o *GetObjectTypeForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetObjectTypeForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetObjectTypeNotFound creates a GetObjectTypeNotFound with default headers values
func NewGetObjectTypeNotFound() *GetObjectTypeNotFound {
	return &GetObjectTypeNotFound{}
}

/*GetObjectTypeNotFound handles this case with default header values.

Not found
*/
type GetObjectTypeNotFound struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetObjectTypeNotFound) Error() string {
	return fmt.Sprintf("[GET /object-types/{id}][%d] getObjectTypeNotFound  %+v", 404, o.Payload)
}

func (o *GetObjectTypeNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetObjectTypeNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
