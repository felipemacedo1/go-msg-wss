package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Testing .env file loading...")

	// Try loading the .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	} else {
		fmt.Println(".env file loaded successfully")
	}

	// Check if MSGWSS_JWT_SECRET is set
	jwtSecret := os.Getenv("MSGWSS_JWT_SECRET")
	fmt.Printf("MSGWSS_JWT_SECRET value: '%s'\n", jwtSecret)

	// Check other environment variables
	dbHost := os.Getenv("MSGWSS_DATABASE_HOST")
	fmt.Printf("MSGWSS_DATABASE_HOST value: '%s'\n", dbHost)
}
