// Package models contains the data structures used in the Vipps MobilePay API
package models

// Amount represents a monetary amount with currency
type Amount struct {
	Currency string `json:"currency"` // NOK, DKK, or EUR
	Value    int    `json:"value"`    // In minor units (øre, cent), e.g. 10.00 NOK = 1000
}

// Customer represents a customer identified by phone number, QR code, or token
type Customer struct {
	PhoneNumber   *string     `json:"phoneNumber,omitempty"`   // Country code + number, e.g. "4712345678"
	PersonalQR    *PersonalQR `json:"personalQr,omitempty"`    // Personal QR code info
	CustomerToken *string     `json:"customerToken,omitempty"` // Customer token
}

// PersonalQR represents a personal QR code
type PersonalQR struct {
	QR string `json:"qr"` // QR code value
}

// PaymentMethod represents the payment method configuration
type PaymentMethod struct {
	Type           string   `json:"type"`                     // Usually "WALLET"
	BlockedSources []string `json:"blockedSources,omitempty"` // Payment sources to block
}

// IndustryData contains additional compliance data
type IndustryData struct {
	AirlineData *AirlineData `json:"airlineData,omitempty"`
}

// AirlineData contains data specific to airline transactions
type AirlineData struct {
	// Airline-specific fields can be added here as needed
}

// Receipt represents a payment receipt
type Receipt struct {
	LineItems []LineItem `json:"lineItems,omitempty"`
}

// LineItem represents an item in a receipt
type LineItem struct {
	Name        string `json:"name"`                  // Name of the item
	Description string `json:"description,omitempty"` // Description of the item
	Quantity    int    `json:"quantity"`              // Number of items
	Amount      Amount `json:"amount"`                // Price per item
	Discount    Amount `json:"discount,omitempty"`    // Discount amount
	VatAmount   Amount `json:"vatAmount,omitempty"`   // VAT amount
	VatPercent  int    `json:"vatPercent,omitempty"`  // VAT percentage
}

// Profile represents user profile information requested
type Profile struct {
	Scope string `json:"scope,omitempty"` // Space-separated list of profile data to request
	Sub   string `json:"sub,omitempty"`   // User's sub ID
}

// QRFormat specifies formatting options for QR codes
type QRFormat struct {
	Format string `json:"format,omitempty"` // Format of the QR code, e.g. "IMAGE_URL"
}

// ProblemDetail represents a standard RFC 7807 problem detail
type ProblemDetail struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code,omitempty"`
}

// Metadata is a map of key-value pairs for storing additional information
type Metadata map[string]string

// AggregateAmount represents aggregated amounts for different payment states
type AggregateAmount struct {
	AuthorizedAmount Amount `json:"authorizedAmount"`
	CapturedAmount   Amount `json:"capturedAmount"`
	RefundedAmount   Amount `json:"refundedAmount"`
	CancelledAmount  Amount `json:"cancelledAmount"`
}
