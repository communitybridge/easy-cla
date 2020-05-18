// Code generated by go-swagger; DO NOT EDIT.

// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

package project

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/models"
)

// GetProjectListReader is a Reader for the GetProjectList structure.
type GetProjectListReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetProjectListReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetProjectListOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetProjectListBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetProjectListUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewGetProjectListForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetProjectListNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetProjectListOK creates a GetProjectListOK with default headers values
func NewGetProjectListOK() *GetProjectListOK {
	return &GetProjectListOK{}
}

/*GetProjectListOK handles this case with default header values.

Success
*/
type GetProjectListOK struct {
	Payload []*models.Project
}

func (o *GetProjectListOK) Error() string {
	return fmt.Sprintf("[GET /projects][%d] getProjectListOK  %+v", 200, o.Payload)
}

func (o *GetProjectListOK) GetPayload() []*models.Project {
	return o.Payload
}

func (o *GetProjectListOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetProjectListBadRequest creates a GetProjectListBadRequest with default headers values
func NewGetProjectListBadRequest() *GetProjectListBadRequest {
	return &GetProjectListBadRequest{}
}

/*GetProjectListBadRequest handles this case with default header values.

Invalid request
*/
type GetProjectListBadRequest struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetProjectListBadRequest) Error() string {
	return fmt.Sprintf("[GET /projects][%d] getProjectListBadRequest  %+v", 400, o.Payload)
}

func (o *GetProjectListBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetProjectListBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetProjectListUnauthorized creates a GetProjectListUnauthorized with default headers values
func NewGetProjectListUnauthorized() *GetProjectListUnauthorized {
	return &GetProjectListUnauthorized{}
}

/*GetProjectListUnauthorized handles this case with default header values.

Unauthorized
*/
type GetProjectListUnauthorized struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetProjectListUnauthorized) Error() string {
	return fmt.Sprintf("[GET /projects][%d] getProjectListUnauthorized  %+v", 401, o.Payload)
}

func (o *GetProjectListUnauthorized) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetProjectListUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetProjectListForbidden creates a GetProjectListForbidden with default headers values
func NewGetProjectListForbidden() *GetProjectListForbidden {
	return &GetProjectListForbidden{}
}

/*GetProjectListForbidden handles this case with default header values.

Insufficient privilege to execute action
*/
type GetProjectListForbidden struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetProjectListForbidden) Error() string {
	return fmt.Sprintf("[GET /projects][%d] getProjectListForbidden  %+v", 403, o.Payload)
}

func (o *GetProjectListForbidden) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetProjectListForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetProjectListNotFound creates a GetProjectListNotFound with default headers values
func NewGetProjectListNotFound() *GetProjectListNotFound {
	return &GetProjectListNotFound{}
}

/*GetProjectListNotFound handles this case with default header values.

Not found
*/
type GetProjectListNotFound struct {
	/*Unique request ID to help in tracing and debugging
	 */
	XREQUESTID string

	Payload *models.ErrorResponse
}

func (o *GetProjectListNotFound) Error() string {
	return fmt.Sprintf("[GET /projects][%d] getProjectListNotFound  %+v", 404, o.Payload)
}

func (o *GetProjectListNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetProjectListNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-REQUEST-ID
	o.XREQUESTID = response.GetHeader("X-REQUEST-ID")

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
