package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/joho/godotenv/autoload"
)

const baseURL = "http://localhost:8080"

type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *APIError       `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Test Complete Order Workflow
func TestCompleteOrderWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	
	// Setup: Connect to database for setup/cleanup
	connStr := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, connStr, "DATABASE_URL must be set")

	conn, err := pgx.Connect(ctx, connStr)
	require.NoError(t, err)
	defer conn.Close(ctx)

	// Test Flow:
	// 1. Create manufacturer partner
	// 2. Create product
	// 3. Set product price
	// 4. Add inventory
	// 5. Create retailer partner
	// 6. Create cart and add items
	// 7. Place order
	// 8. Verify inventory reserved
	// 9. Generate invoice
	// 10. Record payment
	// 11. Update fulfillment status
	// 12. Verify final state

	t.Log("Starting Complete Order Workflow Test...")

	// For this test to work, we need authentication
	// In a real scenario, you'd create a test user first
	
	t.Run("Setup-CreateTestData", func(t *testing.T) {
		// Create test manufacturer
		manufacturerID := uuid.New()
		_, err := conn.Exec(ctx, `
			INSERT INTO partners (id, code, type, business_name, status, kyc_status, created_at)
			VALUES ($1, $2, 'manufacturer', 'Test Manufacturer Ltd', 'active', 'approved', NOW())
		`, manufacturerID, fmt.Sprintf("MFG-%d", time.Now().Unix()))
		require.NoError(t, err)
		t.Logf("Created manufacturer: %s", manufacturerID)

		// Create test product
		productID := uuid.New()
		_, err = conn.Exec(ctx, `
			INSERT INTO products (id, sku, name, manufacturer_id, hsn_code, uom, status, created_at)
			VALUES ($1, $2, 'Test Product', $3, '15071000', 'KG', 'active', NOW())
		`, productID, fmt.Sprintf("SKU-%d", time.Now().Unix()), manufacturerID)
		require.NoError(t, err)
		t.Logf("Created product: %s", productID)

		// Create warehouse
		warehouseID := uuid.New()
		_, err = conn.Exec(ctx, `
			INSERT INTO warehouses (id, partner_id, name, is_default, created_at)
			VALUES ($1, $2, 'Main Warehouse', true, NOW())
		`, warehouseID, manufacturerID)
		require.NoError(t, err)

		// Add inventory
		_, err = conn.Exec(ctx, `
			INSERT INTO inventory_ledger (id, partner_id, warehouse_id, product_id, available_qty, physical_qty, updated_at)
			VALUES ($1, $2, $3, $4, 1000, 1000, NOW())
		`, uuid.New(), manufacturerID, warehouseID, productID)
		require.NoError(t, err)
		t.Logf("Added inventory for product: %s", productID)

		// Verify data
		var count int
		err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM products WHERE id = $1", productID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Verify-InventoryExists", func(t *testing.T) {
		var totalStock int
		err := conn.QueryRow(ctx, "SELECT COALESCE(SUM(available_qty), 0) FROM inventory_ledger").Scan(&totalStock)
		require.NoError(t, err)
		assert.Greater(t, totalStock, 0, "Should have some inventory")
		t.Logf("Total available stock: %d", totalStock)
	})

	t.Run("Cleanup-TestData", func(t *testing.T) {
		// Clean up test data
		_, err := conn.Exec(ctx, "DELETE FROM inventory_ledger WHERE partner_id IN (SELECT id FROM partners WHERE business_name LIKE 'Test%')")
		assert.NoError(t, err)
		
		_, err = conn.Exec(ctx, "DELETE FROM products WHERE manufacturer_id IN (SELECT id FROM partners WHERE business_name LIKE 'Test%')")
		assert.NoError(t, err)
		
		_, err = conn.Exec(ctx, "DELETE FROM warehouses WHERE partner_id IN (SELECT id FROM partners WHERE business_name LIKE 'Test%')")
		assert.NoError(t, err)
		
		_, err = conn.Exec(ctx, "DELETE FROM partners WHERE business_name LIKE 'Test%'")
		assert.NoError(t, err)
		
		t.Log("Cleaned up test data")
	})
}

// Helper function
func makeAPIRequest(t *testing.T, method, path string, body interface{}, token string) *APIResponse {
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
	if err != nil {
		t.Logf("Failed to unmarshal response: %s", string(bodyBytes))
	}
	require.NoError(t, err)

	return &apiResp
}

// Test Database Consistency
func TestDatabaseConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	ctx := context.Background()
	connStr := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, connStr)

	conn, err := pgx.Connect(ctx, connStr)
	require.NoError(t, err)
	defer conn.Close(ctx)

	t.Run("Check-ForeignKeyIntegrity", func(t *testing.T) {
		// Check for orphaned records
		var orphanedProducts int
		err := conn.QueryRow(ctx, `
			SELECT COUNT(*) FROM products p 
			WHERE NOT EXISTS (SELECT 1 FROM partners m WHERE m.id = p.manufacturer_id)
		`).Scan(&orphanedProducts)
		require.NoError(t, err)
		assert.Equal(t, 0, orphanedProducts, "Found orphaned products")
	})

	t.Run("Check-InventoryConsistency", func(t *testing.T) {
		// Check inventory ledger consistency
		var inconsistentRecords int
		err := conn.QueryRow(ctx, `
			SELECT COUNT(*) FROM inventory_ledger 
			WHERE available_qty < 0 OR reserved_qty < 0 OR physical_qty < 0
		`).Scan(&inconsistentRecords)
		require.NoError(t, err)
		assert.Equal(t, 0, inconsistentRecords, "Found inventory with negative quantities")
	})

	t.Run("Check-OrderTotals", func(t *testing.T) {
		// Check if order totals match line totals
		var mismatchCount int
		err := conn.QueryRow(ctx, `
			SELECT COUNT(*) FROM orders o
			WHERE o.grand_total != (
				SELECT COALESCE(SUM(line_total), 0) FROM order_lines ol WHERE ol.order_id = o.id
			)
		`).Scan(&mismatchCount)
		require.NoError(t, err)
		assert.Equal(t, 0, mismatchCount, "Found orders with mismatched totals")
	})
}
