package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gkhaavik/vipps-mobilepay-sdk/pkg/client"
	"github.com/gkhaavik/vipps-mobilepay-sdk/pkg/models"
	"github.com/google/uuid"
)

func main() {
	// Create a new client
	vippsClient := client.NewClient(
		"your-client-id",        // ClientID
		"your-client-secret",    // ClientSecret
		"your-subscription-key", // Subscription key
		"your-msn",              // Merchant Serial Number
		true,                    // Test mode
	)

	// Set system info (optional)
	vippsClient.SetSystemInfo("MyShop", "1.0.0", "MyShopPlugin", "2.0.0")

	// Get access token
	if err := vippsClient.GetAccessToken(); err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}

	// Create payment client
	paymentClient := client.NewPayment(vippsClient)

	// Create a unique reference for the payment
	reference := fmt.Sprintf("order-%s", uuid.New().String())

	// Create payment request
	phoneNumber := "4712345678" // Customer's phone number with country code
	req := models.CreatePaymentRequest{
		Amount: models.Amount{
			Currency: "NOK",
			Value:    1000, // 10.00 NOK (amount in minor units)
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

	// Create payment
	resp, err := paymentClient.Create(req)
	if err != nil {
		log.Fatalf("Failed to create payment: %v", err)
	}

	// Print redirect URL
	fmt.Printf("Payment created successfully!\n")
	fmt.Printf("Payment reference: %s\n", resp.Reference)
	fmt.Printf("Redirect URL: %s\n", resp.RedirectURL)

	// In a real application, redirect the customer to the redirect URL
	fmt.Println("\nSimulating the user completing the payment...")
	time.Sleep(2 * time.Second) // In a real app, the user would complete the payment

	// Check payment status
	payment, err := paymentClient.Get(reference)
	if err != nil {
		log.Fatalf("Failed to get payment: %v", err)
	}

	fmt.Printf("\nPayment status: %s\n", payment.State)

	// In a test environment, you can force approve a payment
	if vippsClient.TestMode {
		fmt.Println("\nForce approving the payment (test mode only)...")
		if err := paymentClient.ForceApprove(reference, phoneNumber); err != nil {
			log.Fatalf("Failed to force approve payment: %v", err)
		}
		fmt.Println("Payment force approved successfully!")

		// Check payment status again
		payment, err = paymentClient.Get(reference)
		if err != nil {
			log.Fatalf("Failed to get payment: %v", err)
		}
		fmt.Printf("Payment status: %s\n", payment.State)
	}

	// If payment is authorized, capture the payment
	if payment.State == models.PaymentStateAuthorized {
		fmt.Println("\nCapturing payment...")
		captureReq := models.ModificationRequest{
			ModificationAmount: models.Amount{
				Currency: "NOK",
				Value:    1000, // Full amount
			},
		}

		captureResp, err := paymentClient.Capture(reference, captureReq)
		if err != nil {
			log.Fatalf("Failed to capture payment: %v", err)
		}

		fmt.Println("Payment captured successfully!")
		fmt.Printf("Captured amount: %.2f NOK\n", float64(captureResp.Aggregate.CapturedAmount.Value)/100)
	}

	// Get payment events
	events, err := paymentClient.GetEvents(reference)
	if err != nil {
		log.Fatalf("Failed to get payment events: %v", err)
	}

	fmt.Println("\nPayment events:")
	for _, event := range events {
		fmt.Printf("- %s: %.2f NOK at %s\n",
			event.Name,
			float64(event.Amount.Value)/100,
			event.Timestamp.Format(time.RFC3339))
	}
}
