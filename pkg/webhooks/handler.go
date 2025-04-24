// Package webhooks provides functionality for working with Vipps MobilePay webhooks
package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gkhaavik/vipps-mobilepay-sdk/pkg/models"
)

// Handler processes webhook events from Vipps MobilePay
type Handler struct {
	SecretKey string
}

// NewHandler creates a new webhook handler
func NewHandler(secretKey string) *Handler {
	return &Handler{
		SecretKey: secretKey,
	}
}

// ValidateSignature validates the signature of a webhook event
func (h *Handler) ValidateSignature(r *http.Request) error {
	// Get the signature from the header
	signatureHeader := r.Header.Get("X-Vipps-Signature")
	if signatureHeader == "" {
		return fmt.Errorf("missing X-Vipps-Signature header")
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Restore the body for later reading
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// Decode the base64 signature
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureHeader)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}

	// Create a new HMAC with SHA256
	mac := hmac.New(sha256.New, []byte(h.SecretKey))
	mac.Write(body)
	expectedSignature := mac.Sum(nil)

	// Compare signatures in constant time to prevent timing attacks
	if !hmac.Equal(signatureBytes, expectedSignature) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// ParseEvent parses a webhook event from an HTTP request
func (h *Handler) ParseEvent(r *http.Request) (*models.WebhookEvent, error) {
	// Validate the signature if a secret key is provided
	if h.SecretKey != "" {
		if err := h.ValidateSignature(r); err != nil {
			return nil, fmt.Errorf("signature validation failed: %w", err)
		}
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Parse the event
	var event models.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	return &event, nil
}

// HandleHTTP creates an http.HandlerFunc that processes webhook events
func (h *Handler) HandleHTTP(handler func(event *models.WebhookEvent) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the event
		event, err := h.ParseEvent(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse event: %v", err), http.StatusBadRequest)
			return
		}

		// Process the event
		if err := handler(event); err != nil {
			// Return a 5xx error so Vipps MobilePay will retry
			http.Error(w, fmt.Sprintf("Failed to process event: %v", err), http.StatusInternalServerError)
			return
		}

		// Acknowledge the event
		w.WriteHeader(http.StatusOK)
	}
}

// EventProcessor is a function that processes a webhook event
type EventProcessor func(*models.WebhookEvent) error

// Router routes webhook events to different handlers based on event type
type Router struct {
	handlers map[models.PaymentEventName]EventProcessor
	fallback EventProcessor
}

// NewRouter creates a new webhook router
func NewRouter() *Router {
	return &Router{
		handlers: make(map[models.PaymentEventName]EventProcessor),
	}
}

// Handle registers a handler for a specific event type
func (r *Router) Handle(eventName models.PaymentEventName, handler EventProcessor) {
	r.handlers[eventName] = handler
}

// HandleFunc registers a handler function for a specific event type
func (r *Router) HandleFunc(eventName models.PaymentEventName, handlerFunc func(*models.WebhookEvent) error) {
	r.handlers[eventName] = handlerFunc
}

// HandleDefault registers a fallback handler for unhandled event types
func (r *Router) HandleDefault(handler EventProcessor) {
	r.fallback = handler
}

// Process routes an event to the appropriate handler
func (r *Router) Process(event *models.WebhookEvent) error {
	if handler, ok := r.handlers[event.Name]; ok {
		return handler(event)
	}

	if r.fallback != nil {
		return r.fallback(event)
	}

	return fmt.Errorf("no handler for event type: %s", event.Name)
}
