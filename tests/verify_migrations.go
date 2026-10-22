package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("🔍 Verifying Database Migrations...")
	fmt.Println("=" + "=====================================")

	// Expected tables
	expectedTables := []string{
		"users", "roles", "permissions", "user_roles",
		"otp_sessions", "refresh_tokens",
		"partners", "kyc_documents", "warehouses",
		"categories", "products", "product_images", "hsn_tax_rates",
		"product_prices", "schemes", "scheme_applicability",
		"carts", "cart_items", "orders", "order_lines", "order_status_history",
		"invoices", "invoice_lines",
		"inventory_ledger", "inventory_movements",
		"credit_accounts", "payments", "ledger_entries",
		"shipments", "shipment_events", "return_requests", "return_lines",
		"notifications", "notification_preferences", "audit_logs", "system_config",
	}

	fmt.Printf("\n✓ Checking %d expected tables...\n", len(expectedTables))

	var totalTables int
	var missingTables []string

	for _, table := range expectedTables {
		var exists bool
		err := conn.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)",
			table).Scan(&exists)

		if err != nil {
			log.Fatalf("Error checking table %s: %v", table, err)
		}

		if exists {
			totalTables++
			fmt.Printf("  ✓ %s\n", table)
		} else {
			missingTables = append(missingTables, table)
			fmt.Printf("  ✗ %s (MISSING)\n", table)
		}
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  Total Expected: %d\n", len(expectedTables))
	fmt.Printf("  Found: %d\n", totalTables)
	fmt.Printf("  Missing: %d\n", len(missingTables))

	if len(missingTables) > 0 {
		fmt.Printf("\n❌ Missing tables:\n")
		for _, table := range missingTables {
			fmt.Printf("  - %s\n", table)
		}
		os.Exit(1)
	}

	// Check for seed data
	fmt.Printf("\n🌱 Checking seed data...\n")

	var roleCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM roles").Scan(&roleCount)
	if err == nil {
		fmt.Printf("  ✓ Roles: %d\n", roleCount)
	}

	var permissionCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM permissions").Scan(&permissionCount)
	if err == nil {
		fmt.Printf("  ✓ Permissions: %d\n", permissionCount)
	}

	var hsnCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM hsn_tax_rates").Scan(&hsnCount)
	if err == nil {
		fmt.Printf("  ✓ HSN Tax Rates: %d\n", hsnCount)
	}

	// Check indexes
	fmt.Printf("\n📇 Checking indexes...\n")
	var indexCount int
	err = conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public'
	`).Scan(&indexCount)
	if err == nil {
		fmt.Printf("  ✓ Total indexes: %d\n", indexCount)
	}

	// Check foreign keys
	fmt.Printf("\n🔗 Checking foreign keys...\n")
	var fkCount int
	err = conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM information_schema.table_constraints 
		WHERE constraint_type = 'FOREIGN KEY' AND table_schema = 'public'
	`).Scan(&fkCount)
	if err == nil {
		fmt.Printf("  ✓ Total foreign keys: %d\n", fkCount)
	}

	fmt.Printf("\n✅ All migrations verified successfully!\n")
}
