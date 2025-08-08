package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"qr-linker/database"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	// Define command-line flags
	var (
		help     = flag.Bool("help", false, "Show help message")
		h        = flag.Bool("h", false, "Show help message (shorthand)")
		dbPath   = flag.String("db", "urls.db", "Path to database file")
		username = flag.String("username", "", "Username for the new user (non-interactive mode)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `QR Linker - Add User Tool

Usage:
  go run cmd/adduser/main.go [options]

Options:
  -h, -help        Show this help message
  -db <path>       Path to database file (default: urls.db)
  -username <name> Specify username directly (will still prompt for password)

Examples:
  # Interactive mode (prompts for username and password)
  go run cmd/adduser/main.go

  # Specify username, prompt for password only
  go run cmd/adduser/main.go -username john

  # Use a different database file
  go run cmd/adduser/main.go -db /path/to/database.db

Description:
  This tool creates new users for the QR Linker application.
  Passwords are securely hashed using bcrypt before storage.
  Usernames must be unique and between 3-50 characters.
  Passwords must be at least 6 characters long.

`)
	}

	flag.Parse()

	if *help || *h {
		flag.Usage()
		os.Exit(0)
	}

	// Initialize database connection
	db, err := database.NewDB(*dbPath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== QR Linker Add User ===")
	fmt.Println()

	// Get username
	var user string
	if *username != "" {
		user = *username
		// Validate provided username
		if err := validateUsername(user, db); err != nil {
			log.Fatal(err)
		}
	} else {
		user = promptUsername(db)
	}

	// Get password
	password := promptPassword()

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create the user
	newUser, err := db.CreateUser(user, hashedPassword)
	if err != nil {
		log.Fatal("Failed to create user:", err)
	}

	fmt.Println()
	fmt.Printf("✓ User '%s' created successfully!\n", newUser.Username)
	fmt.Printf("  ID: %d\n", newUser.ID)
	fmt.Printf("  Created: %s\n", newUser.CreatedAt.Format("2006-01-02 15:04:05"))
}

func validateUsername(username string, db *database.DB) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 50 {
		return fmt.Errorf("username must be less than 50 characters")
	}

	// Check if username already exists
	existingUser, _ := db.GetUserByUsername(username)
	if existingUser != nil {
		return fmt.Errorf("username '%s' already exists", username)
	}

	return nil
}

func promptUsername(db *database.DB) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Username: ")
		username, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Failed to read username:", err)
		}

		username = strings.TrimSpace(username)

		// Validate username
		if username == "" {
			fmt.Println("✗ Username cannot be empty. Please try again.")
			continue
		}

		if len(username) < 3 {
			fmt.Println("✗ Username must be at least 3 characters long. Please try again.")
			continue
		}

		if len(username) > 50 {
			fmt.Println("✗ Username must be less than 50 characters. Please try again.")
			continue
		}

		// Check if username already exists
		existingUser, _ := db.GetUserByUsername(username)
		if existingUser != nil {
			fmt.Printf("✗ Username '%s' already exists. Please choose a different username.\n", username)
			continue
		}

		return username
	}
}

func promptPassword() string {
	for {
		fmt.Print("Password: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // New line after password input
		if err != nil {
			log.Fatal("Failed to read password:", err)
		}

		passwordStr := string(password)

		// Validate password
		if passwordStr == "" {
			fmt.Println("✗ Password cannot be empty. Please try again.")
			continue
		}

		if len(passwordStr) < 6 {
			fmt.Println("✗ Password must be at least 6 characters long. Please try again.")
			continue
		}

		// Confirm password
		fmt.Print("Confirm Password: ")
		confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // New line after password input
		if err != nil {
			log.Fatal("Failed to read password confirmation:", err)
		}

		if passwordStr != string(confirmPassword) {
			fmt.Println("✗ Passwords do not match. Please try again.")
			continue
		}

		return passwordStr
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}