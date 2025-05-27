package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"github.com/junaidrashid-git/ecommerce-api/routes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("✅ Starting application...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️  No .env file found. Using environment variables set by Render or OS.")
	} else {
		log.Println("✅ .env file loaded successfully.")
	}

	// Log important environment variables
	requiredEnvVars := []string{"PORT", "DATABASE_URL", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
	for _, key := range requiredEnvVars {
		if val := os.Getenv(key); val != "" {
			log.Printf("🔑 %s=%s", key, val)
		} else {
			log.Printf("⚠️  %s is not set!", key)
		}
	}

	// Initialize database
	db := initDatabase()
	log.Println("✅ Database connected successfully.")

	// Auto-migrate tables
	if err := db.AutoMigrate(&models.User{}, &models.Product{}, &models.Category{}, &models.Admin{}); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}
	log.Println("✅ Database tables migrated successfully.")

	// Set up Gin router
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-KEY"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve uploaded images
	r.Static("/uploads", "./uploads")

	// Set up routes
	routes.SetupRoutes(r, db)
	log.Println("✅ Routes initialized.")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("⚠️  PORT not set. Using default port 8080.")
	}
	log.Printf("🚀 Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// initDatabase sets up the GORM DB connection, preferring DATABASE_URL if set
func initDatabase() *gorm.DB {
	// If Render provides DATABASE_URL, use it directly
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		log.Println("🔗 Using DATABASE_URL for DB connection")
		db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err != nil {
			log.Fatalf("❌ Failed to connect using DATABASE_URL: %v", err)
		}
		return db
	}

	// Fallback to individual DB_* variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		log.Fatal("❌ One or more DB environment variables are not set.")
	}

	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)

	log.Printf("🔗 Connecting to DB at host=%s, port=%s, dbname=%s, user=%s", host, port, dbname, user)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	return db
}
