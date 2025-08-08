package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"qr-linker/database"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	// Get default database path from environment variables (same logic as main app)
	defaultDBPath := getEnv("DB_PATH_DEV", "")
	if defaultDBPath == "" {
		defaultDBPath = getEnv("DB_PATH", "urls.db")
	}

	// Define command-line flags
	var (
		help   = flag.Bool("help", false, "Show help message")
		h      = flag.Bool("h", false, "Show help message (shorthand)")
		dbPath = flag.String("db", defaultDBPath, "Path to database file")
		list   = flag.Bool("list", false, "List all users and exit")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `QR Linker - User Management Tool

Usage:
  go run cmd/manageusers/main.go [options]

Options:
  -h, -help     Show this help message
  -db <path>    Path to database file (default: urls.db)
  -list         List all users and exit (non-interactive mode)

Interactive Menu Options:
  1. List all users       - Display all registered users with ID and creation date
  2. Add new user         - Create a new user with username and password
  3. Delete user          - Remove an existing user from the database (by ID)
  4. Change password      - Update password for an existing user (by username)
  5. Exit                 - Quit the application

Examples:
  # Interactive mode (menu-driven interface)
  go run cmd/manageusers/main.go

  # Quick user list (non-interactive)
  go run cmd/manageusers/main.go -list

  # Use different database file
  go run cmd/manageusers/main.go -db /path/to/database.db

Security Notes:
  - All passwords are hashed using bcrypt
  - Usernames must be unique (3-50 characters)
  - Passwords must be at least 6 characters
  - User deletion requires confirmation
  - Password changes require confirmation

For adding a single user quickly, consider using:
  go run cmd/adduser/main.go

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

	// Handle non-interactive list mode
	if *list {
		listUsers(db)
		return
	}

	// Interactive mode
	for {
		fmt.Println("\n=== QR Linker User Management ===")
		fmt.Println("1. List all users")
		fmt.Println("2. Add new user")
		fmt.Println("3. Delete user")
		fmt.Println("4. Change password")
		fmt.Println("5. Exit")
		fmt.Print("\nSelect option (1-5): ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			listUsers(db)
		case "2":
			addUser(db)
		case "3":
			deleteUser(db)
		case "4":
			changePassword(db)
		case "5":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func listUsers(db *database.DB) {
	fmt.Println("\n--- User List ---")
	
	users, err := db.GetAllUsers()
	if err != nil {
		fmt.Printf("Error fetching users: %v\n", err)
		return
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}

	fmt.Printf("\n%-5s %-20s %-20s\n", "ID", "Username", "Created")
	fmt.Println(strings.Repeat("-", 50))
	
	for _, user := range users {
		fmt.Printf("%-5d %-20s %-20s\n", 
			user.ID, 
			user.Username, 
			user.CreatedAt.Format("2006-01-02 15:04"))
	}
}

func addUser(db *database.DB) {
	fmt.Println("\n--- Add New User ---")
	
	reader := bufio.NewReader(os.Stdin)
	
	// Get username
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	
	if username == "" {
		fmt.Println("Username cannot be empty.")
		return
	}
	
	// Check if user exists
	existingUser, _ := db.GetUserByUsername(username)
	if existingUser != nil {
		fmt.Printf("User '%s' already exists.\n", username)
		return
	}
	
	// Get password
	fmt.Print("Password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		return
	}
	
	if len(password) < 6 {
		fmt.Println("Password must be at least 6 characters.")
		return
	}
	
	// Confirm password
	fmt.Print("Confirm Password: ")
	confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		return
	}
	
	if string(password) != string(confirmPassword) {
		fmt.Println("Passwords do not match.")
		return
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		return
	}
	
	// Create user
	user, err := db.CreateUser(username, string(hashedPassword))
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}
	
	fmt.Printf("✓ User '%s' created successfully (ID: %d)\n", user.Username, user.ID)
}

func deleteUser(db *database.DB) {
	fmt.Println("\n--- Delete User ---")
	
	// List users first
	listUsers(db)
	
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Print("\nEnter user ID to delete (or 'cancel' to abort): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "cancel" || input == "" {
		fmt.Println("Deletion cancelled.")
		return
	}
	
	// Parse user ID
	userID, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid user ID. Please enter a valid number.")
		return
	}
	
	// Check if user exists
	user, err := db.GetUserByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("User with ID %d not found.\n", userID)
		} else {
			fmt.Printf("Error finding user: %v\n", err)
		}
		return
	}
	
	// Confirm deletion
	fmt.Printf("Are you sure you want to delete user ID %d ('%s')? (yes/no): ", user.ID, user.Username)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	
	if confirm != "yes" && confirm != "y" {
		fmt.Println("Deletion cancelled.")
		return
	}
	
	// Delete user
	err = db.DeleteUser(user.ID)
	if err != nil {
		fmt.Printf("Error deleting user: %v\n", err)
		return
	}
	
	fmt.Printf("✓ User '%s' (ID: %d) deleted successfully.\n", user.Username, user.ID)
}

func changePassword(db *database.DB) {
	fmt.Println("\n--- Change Password ---")
	
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	
	// Check if user exists
	user, err := db.GetUserByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("User '%s' not found.\n", username)
		} else {
			fmt.Printf("Error finding user: %v\n", err)
		}
		return
	}
	
	// Get new password
	fmt.Print("New Password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		return
	}
	
	if len(password) < 6 {
		fmt.Println("Password must be at least 6 characters.")
		return
	}
	
	// Confirm password
	fmt.Print("Confirm New Password: ")
	confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		return
	}
	
	if string(password) != string(confirmPassword) {
		fmt.Println("Passwords do not match.")
		return
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		return
	}
	
	// Update password
	err = db.UpdateUserPassword(user.ID, string(hashedPassword))
	if err != nil {
		fmt.Printf("Error updating password: %v\n", err)
		return
	}
	
	fmt.Printf("✓ Password changed successfully for user '%s'.\n", username)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}