# Vipps MobilePay SDK for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/zenfulcode/vipps-mobilepay-sdk.svg)](https://pkg.go.dev/github.com/zenfulcode/vipps-mobilepay-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/zenfulcode/vipps-mobilepay-sdk)](https://goreportcard.com/report/github.com/zenfulcode/vipps-mobilepay-sdk)
[![License](https://img.shields.io/github/license/zenfulcode/vipps-mobilepay-sdk)](LICENSE)

A comprehensive Go SDK for the Vipps MobilePay ePayment API. This SDK allows you to easily integrate Vipps MobilePay payments into your Go applications.

## Features

- Complete API coverage for the Vipps MobilePay ePayment API
- Authentication handling with automatic token refresh
- Payment creation, capture, refund, and cancellation
- Webhook event handling and verification
- Support for both test and production environments
- Idempotency key support for safe retries
- Comprehensive error handling

## Installation

```bash
go get github.com/zenfulcode/vipps-mobilepay-sdk
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/zenfulcode/vipps-mobilepay-sdk/pkg/client"
	"github.com/zenfulcode/vipps-mobilepay-sdk/pkg/models"
	"github.com/google/uuid"
)

func main() {
	// Create a client
	vippsClient := client.NewClient(
		"your-client-id",
		"your-client-secret",
		"your-subscription-key",
		"your-msn",
		true, // Test mode
	)

	// Get access token
	if err := vippsClient.GetAccessToken(); err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}

	// Create payment client
	paymentClient := client.NewPayment(vippsClient)

	// Generate a unique reference
	reference := fmt.Sprintf("order-%s", uuid.New().String())

	// Create a payment
	phoneNumber := "4712345678"
	req := models.CreatePaymentRequest{
		Amount: models.Amount{
			Currency: "NOK",
			Value:    1000, // 10.00 NOK
		},
		Customer: &models.Customer{
			PhoneNumber: &phoneNumber,
		},
		PaymentMethod: &models.PaymentMethod{
			Type: "WALLET",
		},
		Reference:          reference,
		ReturnURL:          "https://example.com/return?order=" + reference,
		UserFlow:           models.UserFlowWebRedirect,
		PaymentDescription: "Test payment",
	}

	resp, err := paymentClient.Create(req)
	if err != nil {
		log.Fatalf("Failed to create payment: %v", err)
	}

	fmt.Printf("Payment created! Redirect URL: %s\n", resp.RedirectURL)
}
```

## Documentation

### Client Setup

```go
// Create a client for test environment
vippsClient := client.NewClient(
	"your-client-id",
	"your-client-secret",
	"your-subscription-key",
	"your-msn",
	true, // Test mode
)

// Create a client for production environment
prodClient := client.NewClient(
	"your-client-id",
	"your-client-secret",
	"your-subscription-key",
	"your-msn",
	false, // Production mode
)

// Optional: Set system information
vippsClient.SetSystemInfo(
	"MyShopSystem",    // System name
	"1.0.0",           // System version
	"MyShopPlugin",    // Plugin name
	"2.0.0"            // Plugin version
)

// Optional: Set HTTP client timeout
vippsClient.SetTimeout(60 * time.Second)
```

### Payment Operations

```go
// Create payment client
paymentClient := client.NewPayment(vippsClient)

// Create a payment
response, err := paymentClient.Create(createPaymentRequest)

// Get payment details
payment, err := paymentClient.Get("payment-reference")

// Get payment events
events, err := paymentClient.GetEvents("payment-reference")

// Capture a payment
captureReq := models.ModificationRequest{
	ModificationAmount: models.Amount{
		Currency: "NOK",
		Value:    1000,
	},
}
captureResponse, err := paymentClient.Capture("payment-reference", captureReq)

// Refund a payment
refundReq := models.ModificationRequest{
	ModificationAmount: models.Amount{
		Currency: "NOK",
		Value:    500, // Partial refund of 5.00 NOK
	},
}
refundResponse, err := paymentClient.Refund("payment-reference", refundReq)

// Cancel a payment
cancelReq := &models.CancelModificationRequest{
	CancelTransactionOnly: false,
}
cancelResponse, err := paymentClient.Cancel("payment-reference", cancelReq)

// Force approve a payment (test environment only)
err := paymentClient.ForceApprove("payment-reference", "4712345678")
```

### Webhook Management

```go
// Create webhook client
webhookClient := client.NewWebhook(vippsClient)

// Register a new webhook
webhookReq := models.WebhookRegistrationRequest{
	URL: "https://example.com/webhook",
	Events: []string{
		string(models.WebhookEventPaymentAuthorized),
		string(models.WebhookEventPaymentCaptured),
	},
}
webhook, err := webhookClient.Register(webhookReq)

// Get all webhooks
webhooks, err := webhookClient.GetAll()

// Get a specific webhook
webhook, err := webhookClient.Get("webhook-id")

// Delete a webhook
err := webhookClient.Delete("webhook-id")
```

### Handling Webhook Events

```go
// Create a webhook handler with your secret key
secretKey := "your-webhook-secret-key" // From webhook registration
handler := webhooks.NewHandler(secretKey)

// Create a webhook router
router := webhooks.NewRouter()

// Register handlers for different event types
router.HandleFunc(models.EventAuthorized, func(event *models.WebhookEvent) error {
	fmt.Printf("Payment authorized: %s\n", event.Reference)
	return nil
})

router.HandleFunc(models.EventCaptured, func(event *models.WebhookEvent) error {
	fmt.Printf("Payment captured: %s\n", event.Reference)
	return nil
})

// Set up HTTP server with the webhook handler
http.HandleFunc("/webhook", handler.HandleHTTP(router.Process))
http.ListenAndServe(":8080", nil)
```

## Complete Examples

See the `examples` directory for complete examples:

- `examples/payment/main.go`: Shows how to create and manage payments
- `examples/webhook/main.go`: Shows how to handle webhook events

## Error Handling

The SDK provides detailed error information, including HTTP status codes and error messages from the API:

```go
resp, err := paymentClient.Create(req)
if err != nil {
	// err contains the error type, status code, and message
	fmt.Printf("Error: %v\n", err)
	return
}
```

## Testing

For testing your payment integration, you can use the test environment and the force approve functionality:

```go
// Force approve a payment in the test environment
err := paymentClient.ForceApprove(reference, "4712345678")
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Resources

- [Vipps MobilePay Developer Portal](https://developer.vippsmobilepay.com/)
- [ePayment API Documentation](https://developer.vippsmobilepay.com/docs/APIs/epayment-api/)
- [Webhooks API Documentation](https://developer.vippsmobilepay.com/docs/APIs/webhooks-api/)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
