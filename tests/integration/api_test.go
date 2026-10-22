package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"

// Helper struct for API responses
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *APIError       `json:"error,omitempty"`
	Meta    json.RawMessage `json:"meta"`
}

type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Test Helper Functions
func makeRequest(t *testing.T, method, path string, body interface{}, token string) *APIResponse {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var apiResp APIResponse
	err = json.Unmarshal(bodyBytes, &apiResp)
	require.NoError(t, err, "Response body: %s", string(bodyBytes))

	return &apiResp
}

// Test 1: Health Check
func TestHealthCheck(t *testing.T) {
	resp := makeRequest(t, "GET", "/health", nil, "")
	assert.True(t, resp.Success)

	var data map[string]string
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)
	assert.Equal(t, "ok", data["status"])
}

// Test 2: Readiness Check
func TestReadinessCheck(t *testing.T) {
	resp := makeRequest(t, "GET", "/ready", nil, "")
	assert.True(t, resp.Success)

	var data map[string]string
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)
	assert.Equal(t, "ready", data["status"])
}

// Test 3: OTP Flow
func TestOTPFlow(t *testing.T) {
	phone := fmt.Sprintf("+9198765%05d", time.Now().Unix()%100000)

	// Step 1: Send OTP
	t.Run("SendOTP", func(t *testing.T) {
		body := map[string]string{"phone": phone}
		resp := makeRequest(t, "POST", "/api/v1/auth/otp/send", body, "")
		assert.True(t, resp.Success, "Failed to send OTP")
	})

	// Note: In real tests, you'd need to retrieve the OTP from database or mock
	// For now, this demonstrates the test structure
}

// Test 4: Product Listing (Unauthorized)
func TestProductListingUnauthorized(t *testing.T) {
	resp := makeRequest(t, "GET", "/api/v1/products", nil, "")
	// This might return 401 or allow public listing depending on your implementation
	// Adjust assertion based on your auth requirements
	t.Logf("Product listing response: success=%v", resp.Success)
}

// Test 5: Cart Operations
func TestCartOperations(t *testing.T) {
	// This test requires authentication
	// In a real scenario, you'd first authenticate and get a token
	
	t.Skip("Skipping: Requires authentication token")

	token := "your-jwt-token-here"

	// Create cart
	t.Run("CreateCart", func(t *testing.T) {
		body := map[string]string{
			"seller_id": "some-uuid",
		}
		resp := makeRequest(t, "POST", "/api/v1/cart", body, token)
		assert.True(t, resp.Success)
	})

	// Get cart
	t.Run("GetCart", func(t *testing.T) {
		resp := makeRequest(t, "GET", "/api/v1/cart", nil, token)
		assert.True(t, resp.Success)
	})
}

// Test 6: Admin Health Check
func TestAdminHealthCheck(t *testing.T) {
	resp := makeRequest(t, "GET", "/api/v1/admin/health", nil, "")
	assert.True(t, resp.Success)

	var data map[string]string
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)
	assert.Equal(t, "admin_service_ok", data["status"])
}

// Test 7: Inventory Listing (Unauthorized)
func TestInventoryListingUnauthorized(t *testing.T) {
	resp := makeRequest(t, "GET", "/api/v1/inventory", nil, "")
	
	// Should return 401 Unauthorized
	if resp.Success {
		t.Log("Inventory listing allowed without auth - check if this is intentional")
	} else {
		assert.NotNil(t, resp.Error)
		t.Logf("Error code: %s", resp.Error.Code)
	}
}

// Test 8: Dashboard Without Auth
func TestDashboardWithoutAuth(t *testing.T) {
	resp := makeRequest(t, "GET", "/api/v1/dashboard", nil, "")
	
	// Should require authentication
	if resp.Success {
		t.Log("Dashboard accessible without auth - check if this is intentional")
	} else {
		assert.NotNil(t, resp.Error)
	}
}

// Test 9: Invalid Endpoint
func TestInvalidEndpoint(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/nonexistent", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return 404
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Test 10: CORS Headers (if implemented)
func TestCORSHeaders(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", baseURL+"/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	t.Logf("CORS headers present: %v", resp.Header.Get("Access-Control-Allow-Origin") != "")
}
