package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Connect to the default 'postgres' database
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:9v`z7NmX0<<fV*~u@localhost:5432/postgres?sslmode=disable"
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Check if database exists
	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'agribid_shudh')").Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if database exists: %v", err)
	}

	if exists {
		fmt.Println("✅ Database 'agribid_shudh' already exists")
		return
	}

	// Create the database
	_, err = conn.Exec(ctx, "CREATE DATABASE agribid_shudh")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	fmt.Println("✅ Successfully created database 'agribid_shudh'")
}
