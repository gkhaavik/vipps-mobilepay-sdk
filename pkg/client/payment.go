package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/zenfulcode/vipps-mobilepay-sdk/pkg/models"
)

// Payment handles all payment-related API calls
type Payment struct {
	client *Client
}

// NewPayment creates a new payment API handler
func NewPayment(client *Client) *Payment {
	return &Payment{
		client: client,
	}
}

// Create initiates a new payment
func (p *Payment) Create(req models.CreatePaymentRequest) (*models.CreatePaymentResponse, error) {
	endpoint := "/epayment/v1/payments"

	// Generate a new idempotency key for the request
	idempotencyKey := uuid.New().String()

	body, statusCode, err := p.client.DoRequest(http.MethodPost, endpoint, req, idempotencyKey)
	if err != nil {
		log.Printf("Error creating payment, status code: %d, response: %s", statusCode, string(body))
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	var response models.CreatePaymentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// Get retrieves information about a payment by its reference
func (p *Payment) Get(reference string) (*models.GetPaymentResponse, error) {
	endpoint := fmt.Sprintf("/epayment/v1/payments/%s", reference)

	body, _, err := p.client.DoRequest(http.MethodGet, endpoint, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	var response models.GetPaymentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetEvents retrieves the event log for a payment by its reference
func (p *Payment) GetEvents(reference string) ([]models.PaymentEvent, error) {
	endpoint := fmt.Sprintf("/epayment/v1/payments/%s/events", reference)

	body, _, err := p.client.DoRequest(http.MethodGet, endpoint, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get payment events: %w", err)
	}

	var events []models.PaymentEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return events, nil
}

// Capture captures funds from a previously authorized payment
func (p *Payment) Capture(reference string, req models.ModificationRequest) (*models.AdjustmentResponse, error) {
	endpoint := fmt.Sprintf("/epayment/v1/payments/%s/capture", reference)

	idempotencyKey := uuid.New().String()
	body, _, err := p.client.DoRequest(http.MethodPost, endpoint, req, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to capture payment: %w", err)
	}

	var response models.AdjustmentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// Refund returns funds from a previously captured payment
func (p *Payment) Refund(reference string, req models.ModificationRequest) (*models.AdjustmentResponse, error) {
	endpoint := fmt.Sprintf("/epayment/v1/payments/%s/refund", reference)

	idempotencyKey := uuid.New().String()
	body, _, err := p.client.DoRequest(http.MethodPost, endpoint, req, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to refund payment: %w", err)
	}

	var response models.AdjustmentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// Cancel cancels a payment
func (p *Payment) Cancel(reference string, req *models.CancelModificationRequest) (*models.AdjustmentResponse, error) {
	endpoint := fmt.Sprintf("/epayment/v1/payments/%s/cancel", reference)

	body, _, err := p.client.DoRequest(http.MethodPost, endpoint, req, "")
	if err != nil {
		return nil, fmt.Errorf("failed to cancel payment: %w", err)
	}

	var response models.AdjustmentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// ForceApprove force approves a payment (only available in test environment)
func (p *Payment) ForceApprove(reference string, customerPhoneNumber string) error {
	if !p.client.TestMode {
		return fmt.Errorf("force approve is only available in test environment")
	}

	endpoint := fmt.Sprintf("/epayment/v1/test/payments/%s/approve", reference)

	// Prepare the request body according to API specs
	reqBody := struct {
		Customer struct {
			PhoneNumber string `json:"phoneNumber"`
		} `json:"customer"`
	}{}
	reqBody.Customer.PhoneNumber = customerPhoneNumber

	idempotencyKey := uuid.New().String()
	_, _, err := p.client.DoRequest(http.MethodPost, endpoint, reqBody, idempotencyKey)
	if err != nil {
		return fmt.Errorf("failed to force approve payment: %w", err)
	}

	return nil
}
