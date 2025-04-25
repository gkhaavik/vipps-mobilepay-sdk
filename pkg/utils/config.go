package utils

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gkhaavik/vipps-mobilepay-sdk/pkg/client"
)

// DefaultEnvFile is the default path to the .env file
const DefaultEnvFile = ".env"

var PhoneNumber string
var WebhookURL string

// LoadEnvFromRoot attempts to load the .env file from the project root
func LoadEnvFromRoot() error {
	// Try relative paths starting from the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Try to find .env file walking up from current directory
	dir := currentDir
	for {
		envPath := filepath.Join(dir, DefaultEnvFile)
		if _, err := os.Stat(envPath); err == nil {
			return LoadEnv(envPath)
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached the root directory
			break
		}
		dir = parent
	}

	// As a fallback, try with the current directory
	return LoadEnv(DefaultEnvFile)
}

// NewClientFromEnv creates a new Vipps MobilePay client using environment variables
func NewClientFromEnv() (*client.Client, error) {
	// Try to load environment variables from .env file, but don't fail if not found
	_ = LoadEnvFromRoot()

	// Get configuration from environment
	clientID := GetEnv("VIPPS_CLIENT_ID", "")
	clientSecret := GetEnv("VIPPS_CLIENT_SECRET", "")
	subscriptionKey := GetEnv("VIPPS_SUBSCRIPTION_KEY", "")
	msn := GetEnv("VIPPS_MSN", "")
	testMode := GetEnvBool("VIPPS_TEST_MODE", true)
	PhoneNumber = GetEnv("VIPPS_PHONE_NUMBER", "")
	WebhookURL = GetEnv("VIPPS_WEBHOOK_URL", "")

	// Create client
	vippsClient := client.NewClient(
		clientID,
		clientSecret,
		subscriptionKey,
		msn,
		testMode,
	)

	// Set optional system information
	vippsClient.SetSystemInfo(
		GetEnv("VIPPS_SYSTEM_NAME", "go-vipps-mobilepay-sdk"),
		GetEnv("VIPPS_SYSTEM_VERSION", "1.0.0"),
		GetEnv("VIPPS_SYSTEM_PLUGIN_NAME", "Mobilepay SDK"),
		GetEnv("VIPPS_SYSTEM_PLUGIN_VERSION", "0.0.1"),
	)

	// Set timeout if specified
	if timeoutStr := GetEnv("VIPPS_TIMEOUT", ""); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			vippsClient.SetTimeout(timeout)
		}
	}

	// Get access token
	err := vippsClient.GetAccessToken()

	return vippsClient, err
}
