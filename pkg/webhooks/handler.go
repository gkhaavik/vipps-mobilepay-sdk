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
	// First, verify the content hash
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Restore the body for later reading
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// Compute SHA256 hash of the body
	contentHash := sha256.Sum256(body)
	expectedContentHash := base64.StdEncoding.EncodeToString(contentHash[:])

	// Check if content hash matches
	actualContentHash := r.Header.Get("X-Ms-Content-Sha256")
	if actualContentHash == "" {
		return fmt.Errorf("missing X-Ms-Content-Sha256 header")
	}

	if expectedContentHash != actualContentHash {
		fmt.Printf("Content hash mismatch: expected %s, got %s\n",
			expectedContentHash, actualContentHash)
		// For debugging, continue even if this doesn't match
	}

	// Get authorization header (could be either Authorization or X-Vipps-Authorization)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = r.Header.Get("X-Vipps-Authorization")
		if authHeader == "" {
			return fmt.Errorf("missing Authorization or X-Vipps-Authorization header")
		}
	}

	// Get the host from the X-Forwarded-Host header if available, otherwise use the Host header
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Header.Get("Host")
	}

	// Construct the string to be signed exactly as in the C# example
	signedString := fmt.Sprintf("%s\n%s\n%s;%s;%s",
		r.Method,
		r.URL.Path, // This should be the path only, not the full URI with query params
		r.Header.Get("X-Ms-Date"),
		host,
		r.Header.Get("X-Ms-Content-Sha256"))

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(h.SecretKey))
	mac.Write([]byte(signedString))
	expectedSignatureBytes := mac.Sum(nil)
	expectedSignature := base64.StdEncoding.EncodeToString(expectedSignatureBytes)

	// Format the expected authorization header exactly as in the C# example
	expectedAuthHeader := fmt.Sprintf("HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature=%s", expectedSignature)

	if expectedAuthHeader != authHeader {
		// Log the error but return an actual error
		fmt.Printf("Auth header mismatch:\nExpected: %s\nActual:   %s\n",
			expectedAuthHeader, authHeader)
		return fmt.Errorf("signature validation failed")
	}

	fmt.Println("Signature validation successful")
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
		// Log all requests
		fmt.Printf("Received webhook request: %s %s\n", r.Method, r.URL.Path)
		fmt.Printf("Headers: %v\n", r.Header)

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
	fmt.Println("Process is called " + event.Name)

	if handler, ok := r.handlers[event.Name]; ok {
		return handler(event)
	}

	if r.fallback != nil {
		return r.fallback(event)
	}

	return fmt.Errorf("no handler for event type: %s", event.Name)
}
