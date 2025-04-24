package models

import "time"

// WebhookEvent represents the structure of a webhook event
type WebhookEvent struct {
	MSN            string           `json:"msn"`                      // The merchant serial number
	Reference      string           `json:"reference"`                // The payment reference
	PSPReference   string           `json:"pspReference"`             // The PSP reference
	Name           PaymentEventName `json:"name"`                     // The event type
	Amount         Amount           `json:"amount"`                   // The amount for the event
	Timestamp      time.Time        `json:"timestamp"`                // When the event occurred
	IdempotencyKey string           `json:"idempotencyKey,omitempty"` // Idempotency key if applicable
	Success        bool             `json:"success"`                  // Whether the operation succeeded
}

// WebhookRegistration represents a webhook registration
type WebhookRegistration struct {
	ID        string   `json:"id"`                  // The unique identifier for this webhook
	URL       string   `json:"url"`                 // The callback URL where notifications are sent
	Events    []string `json:"events"`              // List of event types to subscribe to
	Created   string   `json:"created,omitempty"`   // When the webhook was registered
	Status    string   `json:"status,omitempty"`    // The status of the webhook (active, etc.)
	MSN       string   `json:"msn,omitempty"`       // The merchant serial number
	SecretKey string   `json:"secretKey,omitempty"` // The secret key for validating signatures
}

// WebhookRegistrationRequest represents a request to register a webhook
type WebhookRegistrationRequest struct {
	URL    string   `json:"url"`    // The callback URL where notifications are sent
	Events []string `json:"events"` // List of event types to subscribe to
}

// WebhookEventType defines the available webhook event types
type WebhookEventType string

const (
	// WebhookEventPaymentCreated is sent when a payment is created
	WebhookEventPaymentCreated WebhookEventType = "epayments.payment.created.v1"
	// WebhookEventPaymentAborted is sent when a payment is aborted by the user
	WebhookEventPaymentAborted WebhookEventType = "epayments.payment.aborted.v1"
	// WebhookEventPaymentExpired is sent when a payment expires
	WebhookEventPaymentExpired WebhookEventType = "epayments.payment.expired.v1"
	// WebhookEventPaymentCancelled is sent when a payment is cancelled by the merchant
	WebhookEventPaymentCancelled WebhookEventType = "epayments.payment.cancelled.v1"
	// WebhookEventPaymentCaptured is sent when a payment is captured by the merchant
	WebhookEventPaymentCaptured WebhookEventType = "epayments.payment.captured.v1"
	// WebhookEventPaymentRefunded is sent when a payment is refunded by the merchant
	WebhookEventPaymentRefunded WebhookEventType = "epayments.payment.refunded.v1"
	// WebhookEventPaymentAuthorized is sent when a payment is authorized by the user
	WebhookEventPaymentAuthorized WebhookEventType = "epayments.payment.authorized.v1"
	// WebhookEventPaymentTerminated is sent when a payment is terminated by the merchant
	WebhookEventPaymentTerminated WebhookEventType = "epayments.payment.terminated.v1"
)
