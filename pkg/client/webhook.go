package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gkhaavik/vipps-mobilepay-sdk/pkg/models"
)

// Webhook handles all webhook-related API calls
type Webhook struct {
	client *Client
}

// NewWebhook creates a new webhook API handler
func NewWebhook(client *Client) *Webhook {
	return &Webhook{
		client: client,
	}
}

// Register registers a new webhook
func (w *Webhook) Register(req models.WebhookRegistrationRequest) (*models.WebhookRegistration, error) {
	endpoint := "/webhooks/v1/webhooks"

	body, _, err := w.client.DoRequest(http.MethodPost, endpoint, req, "")
	if err != nil {
		return nil, fmt.Errorf("failed to register webhook: %w", err)
	}

	var response models.WebhookRegistration
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// webhooksResponse is a wrapper for the API response which contains a webhooks array
type webhooksResponse struct {
	Webhooks []models.WebhookRegistration `json:"webhooks"`
}

// GetAll retrieves all registered webhooks
func (w *Webhook) GetAll() ([]models.WebhookRegistration, error) {
	endpoint := "/webhooks/v1/webhooks"

	body, _, err := w.client.DoRequest(http.MethodGet, endpoint, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get webhooks: %w", err)
	}

	// Try parsing with the correct wrapper structure first
	var wrappedResponse webhooksResponse
	if err := json.Unmarshal(body, &wrappedResponse); err != nil {
		// Fall back to the old format in case API changes again
		var directResponse []models.WebhookRegistration
		if err2 := json.Unmarshal(body, &directResponse); err2 != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return directResponse, nil
	}

	return wrappedResponse.Webhooks, nil
}

// Get retrieves a specific webhook by ID
func (w *Webhook) Get(id string) (*models.WebhookRegistration, error) {
	endpoint := fmt.Sprintf("/webhooks/v1/webhooks/%s", id)

	body, _, err := w.client.DoRequest(http.MethodGet, endpoint, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	var response models.WebhookRegistration
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// Delete removes a webhook registration
func (w *Webhook) Delete(id string) error {
	endpoint := fmt.Sprintf("/webhooks/v1/webhooks/%s", id)

	_, _, err := w.client.DoRequest(http.MethodDelete, endpoint, nil, "")
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	return nil
}
