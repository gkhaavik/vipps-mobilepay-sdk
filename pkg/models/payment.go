package models

import "time"

// PaymentUserFlow defines the flow for bringing users to the payment app
type PaymentUserFlow string

const (
	// UserFlowPushMessage sends push notification to user's app
	UserFlowPushMessage PaymentUserFlow = "PUSH_MESSAGE"
	// UserFlowWebRedirect redirects user to the landing page
	UserFlowWebRedirect PaymentUserFlow = "WEB_REDIRECT"
	// UserFlowNativeRedirect redirects between apps (not recommended)
	UserFlowNativeRedirect PaymentUserFlow = "NATIVE_REDIRECT"
	// UserFlowQR returns a QR code for the payment
	UserFlowQR PaymentUserFlow = "QR"
)

// CustomerInteraction defines how the customer interacts with the merchant
type CustomerInteraction string

const (
	// CustomerPresent means the customer is physically present
	CustomerPresent CustomerInteraction = "CUSTOMER_PRESENT"
	// CustomerNotPresent means the customer is not physically present
	CustomerNotPresent CustomerInteraction = "CUSTOMER_NOT_PRESENT"
)

// PaymentState represents the current state of a payment
type PaymentState string

const (
	// PaymentStateCreated means the payment has been initiated but not acted upon
	PaymentStateCreated PaymentState = "CREATED"
	// PaymentStateAuthorized means the payment has been accepted by the user
	PaymentStateAuthorized PaymentState = "AUTHORIZED"
	// PaymentStateAborted means the payment was actively stopped by the user
	PaymentStateAborted PaymentState = "ABORTED"
	// PaymentStateExpired means the user did not act on the payment within the time limit
	PaymentStateExpired PaymentState = "EXPIRED"
	// PaymentStateTerminated means the merchant canceled the payment before authorization
	PaymentStateTerminated PaymentState = "TERMINATED"
)

// PaymentEventName represents the type of payment event
type PaymentEventName string

const (
	// EventCreated indicates a payment was created
	EventCreated PaymentEventName = "CREATED"
	// EventAuthorized indicates a payment was authorized
	EventAuthorized PaymentEventName = "AUTHORIZED"
	// EventAborted indicates a payment was aborted by the user
	EventAborted PaymentEventName = "ABORTED"
	// EventExpired indicates a payment expired
	EventExpired PaymentEventName = "EXPIRED"
	// EventCancelled indicates a payment was cancelled by the merchant
	EventCancelled PaymentEventName = "CANCELLED"
	// EventCaptured indicates a payment was captured by the merchant
	EventCaptured PaymentEventName = "CAPTURED"
	// EventRefunded indicates a payment was refunded by the merchant
	EventRefunded PaymentEventName = "REFUNDED"
	// EventTerminated indicates a payment was terminated by the merchant
	EventTerminated PaymentEventName = "TERMINATED"
)

// CreatePaymentRequest represents a request to create a new payment
type CreatePaymentRequest struct {
	Amount              Amount              `json:"amount"`                        // Required: payment amount
	Customer            *Customer           `json:"customer,omitempty"`            // Customer identification
	MinimumUserAge      *int                `json:"minimumUserAge,omitempty"`      // Min age required for user (0-100)
	CustomerInteraction CustomerInteraction `json:"customerInteraction,omitempty"` // Default: CUSTOMER_NOT_PRESENT
	IndustryData        *IndustryData       `json:"industryData,omitempty"`        // Additional compliance data
	PaymentMethod       *PaymentMethod      `json:"paymentMethod"`                 // Required: payment method configuration
	Profile             *Profile            `json:"profile,omitempty"`             // User profile information to request
	Reference           string              `json:"reference"`                     // Required: unique identifier for the payment
	ReturnURL           string              `json:"returnUrl,omitempty"`           // URL to return to after payment
	UserFlow            PaymentUserFlow     `json:"userFlow"`                      // Required: how to bring user to payment
	ExpiresAt           *time.Time          `json:"expiresAt,omitempty"`           // When the payment expires (long-living payments)
	QRFormat            *QRFormat           `json:"qrFormat,omitempty"`            // QR code format options
	PaymentDescription  string              `json:"paymentDescription,omitempty"`  // Description shown to the user
	Receipt             *Receipt            `json:"receipt,omitempty"`             // Receipt information
	Metadata            Metadata            `json:"metadata,omitempty"`            // Additional metadata
	ReceiptURL          string              `json:"receiptUrl,omitempty"`          // URL to view or download receipt
}

// CreatePaymentResponse represents the response after creating a payment
type CreatePaymentResponse struct {
	RedirectURL string `json:"redirectUrl"`          // URL for continuing the payment flow
	Reference   string `json:"reference"`            // Unique reference for the payment
	QRImageURL  string `json:"qrImageUrl,omitempty"` // URL to QR image if UserFlow is QR
}

// GetPaymentResponse represents the response when getting payment details
type GetPaymentResponse struct {
	Aggregate       *AggregateAmount `json:"aggregate"`                 // Aggregated amounts
	Amount          Amount           `json:"amount"`                    // Original payment amount
	State           PaymentState     `json:"state"`                     // Current payment state
	PaymentMethod   *PaymentMethod   `json:"paymentMethod,omitempty"`   // Payment method used
	Profile         *Profile         `json:"profile,omitempty"`         // User profile information
	PSPReference    string           `json:"pspReference"`              // Reference from payment service provider
	RedirectURL     string           `json:"redirectUrl,omitempty"`     // URL for continuing the payment flow
	Reference       string           `json:"reference"`                 // Unique reference for the payment
	Metadata        Metadata         `json:"metadata,omitempty"`        // Additional metadata
	CardBin         string           `json:"cardBin,omitempty"`         // First 6 digits of card if card payment
	CustomerName    string           `json:"customerName,omitempty"`    // Customer name if available
	CustomerPhone   string           `json:"customerPhone,omitempty"`   // Customer phone if available
	CustomerEmail   string           `json:"customerEmail,omitempty"`   // Customer email if available
	CustomerAddress string           `json:"customerAddress,omitempty"` // Customer address if available
}

// PaymentEvent represents an event in a payment's history
type PaymentEvent struct {
	Reference      string           `json:"reference"`                // Payment reference
	PSPReference   string           `json:"pspReference"`             // PSP reference for this event
	Name           PaymentEventName `json:"name"`                     // Type of event
	Amount         Amount           `json:"amount"`                   // Amount for this event
	Timestamp      time.Time        `json:"timestamp"`                // When the event occurred
	IdempotencyKey string           `json:"idempotencyKey,omitempty"` // Idempotency key if applicable
	Success        bool             `json:"success"`                  // Whether the operation succeeded
}

// ModificationRequest represents a request to modify a payment
type ModificationRequest struct {
	ModificationAmount Amount `json:"modificationAmount"` // Amount to capture, refund, etc.
}

// CancelModificationRequest represents a request to cancel a payment
type CancelModificationRequest struct {
	CancelTransactionOnly bool `json:"cancelTransactionOnly,omitempty"` // Only cancel if not authorized
}

// AdjustmentResponse represents the response to a payment modification
type AdjustmentResponse struct {
	Amount       Amount          `json:"amount"`       // Current payment amount
	State        PaymentState    `json:"state"`        // Current payment state
	Aggregate    AggregateAmount `json:"aggregate"`    // Aggregated amounts
	PSPReference string          `json:"pspReference"` // Reference from payment service provider
	Reference    string          `json:"reference"`    // Unique reference for the payment
}
