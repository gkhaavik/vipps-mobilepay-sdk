package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gkhaavik/vipps-mobilepay-sdk/pkg/client"
	"github.com/gkhaavik/vipps-mobilepay-sdk/pkg/models"
	"github.com/gkhaavik/vipps-mobilepay-sdk/pkg/utils"
	"github.com/gkhaavik/vipps-mobilepay-sdk/pkg/webhooks"
)

func main() {
	// Create a new client
	vippsClient, err := utils.NewClientFromEnv()

	if err != nil {
		log.Fatalf("Failed to create Vipps client: %v", err)
	}

	// Get access token
	if err := vippsClient.GetAccessToken(); err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}

	// Create webhook client
	webhookClient := client.NewWebhook(vippsClient)

	// Register a webhook (usually you'd do this once during setup)
	// For this example, we'll just check if there are existing webhooks
	existingWebhooks, err := webhookClient.GetAll()
	if err != nil {
		log.Fatalf("Failed to get webhooks: %v", err)
	}

	var secretKey string

	if len(existingWebhooks) > 0 {
		// remove all webhooks
		for _, webhook := range existingWebhooks {
			if err := webhookClient.Delete(webhook.ID); err != nil {
				log.Printf("Failed to remove webhook %s: %v", webhook.ID, err)
			} else {
				fmt.Printf("Webhook %s removed successfully\n", webhook.ID)
			}
		}
	}

	// Register a new webhook
	webhookReq := models.WebhookRegistrationRequest{
		URL: utils.WebhookURL, // Replace with your actual webhook endpoint
		Events: []string{
			string(models.WebhookEventPaymentAuthorized),
			string(models.WebhookEventPaymentCaptured),
			string(models.WebhookEventPaymentRefunded),
		},
	}

	webhook, err := webhookClient.Register(webhookReq)
	if err != nil {
		log.Fatalf("Failed to register webhook: %v", err)
	}

	secretKey = webhook.Secret

	fmt.Printf("Webhook registered successfully! ID: %s\n", webhook.ID)

	// Create a webhook handler
	// In a production environment, you would get this from your webhook registration
	handler := webhooks.NewHandler(secretKey)

	// Create a webhook router
	router := webhooks.NewRouter()

	// Register handlers for different event types
	router.HandleFunc(models.EventAuthorized, handleAuthorized)
	router.HandleFunc(models.EventCaptured, handleCaptured)
	router.HandleFunc(models.EventRefunded, handleRefunded)

	// Register a default handler for other events
	router.HandleDefault(func(event *models.WebhookEvent) error {
		fmt.Printf("Received unhandled event: %s\n", event.Name)
		return nil
	})

	// Set up HTTP server with the webhook handler
	http.HandleFunc("/webhook", handler.HandleHTTP(router.Process))

	// Start server in a goroutine
	server := &http.Server{
		Addr: ":8080",
	}

	go func() {
		fmt.Println("Starting webhook server on :8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("Shutting down server...")
	server.Close()
}

func handleAuthorized(event *models.WebhookEvent) error {
	fmt.Printf("Payment authorized: Reference: %s, Amount: %.2f %s\n",
		event.Reference,
		float64(event.Amount.Value)/100,
		event.Amount.Currency)

	// trigger other business logic based on the authorized payment

	return nil
}

func handleCaptured(event *models.WebhookEvent) error {
	fmt.Printf("Payment captured: Reference: %s, Amount: %.2f %s\n",
		event.Reference,
		float64(event.Amount.Value)/100,
		event.Amount.Currency)

	// In a real application, you would mark the order as paid

	return nil
}

func handleRefunded(event *models.WebhookEvent) error {
	fmt.Printf("Payment refunded: Reference: %s, Amount: %.2f %s\n",
		event.Reference,
		float64(event.Amount.Value)/100,
		event.Amount.Currency)

	// In a real application, you would process the refund in your system

	return nil
}
