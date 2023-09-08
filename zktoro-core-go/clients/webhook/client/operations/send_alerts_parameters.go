// Code generated by go-swagger; DO NOT EDIT.

package operations

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

	"zktoro/zktoro-core-go/clients/webhook/client/models"
)

// NewSendAlertsParams creates a new SendAlertsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewSendAlertsParams() *SendAlertsParams {
	return &SendAlertsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewSendAlertsParamsWithTimeout creates a new SendAlertsParams object
// with the ability to set a timeout on a request.
func NewSendAlertsParamsWithTimeout(timeout time.Duration) *SendAlertsParams {
	return &SendAlertsParams{
		timeout: timeout,
	}
}

// NewSendAlertsParamsWithContext creates a new SendAlertsParams object
// with the ability to set a context for a request.
func NewSendAlertsParamsWithContext(ctx context.Context) *SendAlertsParams {
	return &SendAlertsParams{
		Context: ctx,
	}
}

// NewSendAlertsParamsWithHTTPClient creates a new SendAlertsParams object
// with the ability to set a custom HTTPClient for a request.
func NewSendAlertsParamsWithHTTPClient(client *http.Client) *SendAlertsParams {
	return &SendAlertsParams{
		HTTPClient: client,
	}
}

/*
SendAlertsParams contains all the parameters to send to the API endpoint

	for the send alerts operation.

	Typically these are written to a http.Request.
*/
type SendAlertsParams struct {

	/* Authorization.

	   Webhook request authorization
	*/
	Authorization *string

	// Payload.
	Payload *models.AlertBatch

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the send alerts params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *SendAlertsParams) WithDefaults() *SendAlertsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the send alerts params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *SendAlertsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the send alerts params
func (o *SendAlertsParams) WithTimeout(timeout time.Duration) *SendAlertsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the send alerts params
func (o *SendAlertsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the send alerts params
func (o *SendAlertsParams) WithContext(ctx context.Context) *SendAlertsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the send alerts params
func (o *SendAlertsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the send alerts params
func (o *SendAlertsParams) WithHTTPClient(client *http.Client) *SendAlertsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the send alerts params
func (o *SendAlertsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAuthorization adds the authorization to the send alerts params
func (o *SendAlertsParams) WithAuthorization(authorization *string) *SendAlertsParams {
	o.SetAuthorization(authorization)
	return o
}

// SetAuthorization adds the authorization to the send alerts params
func (o *SendAlertsParams) SetAuthorization(authorization *string) {
	o.Authorization = authorization
}

// WithPayload adds the payload to the send alerts params
func (o *SendAlertsParams) WithPayload(payload *models.AlertBatch) *SendAlertsParams {
	o.SetPayload(payload)
	return o
}

// SetPayload adds the payload to the send alerts params
func (o *SendAlertsParams) SetPayload(payload *models.AlertBatch) {
	o.Payload = payload
}

// WriteToRequest writes these params to a swagger request
func (o *SendAlertsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Authorization != nil {

		// header param Authorization
		if err := r.SetHeaderParam("Authorization", *o.Authorization); err != nil {
			return err
		}
	}
	if o.Payload != nil {
		if err := r.SetBodyParam(o.Payload); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
