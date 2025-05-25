package main

import (
	
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"github.com/junaidrashid-git/ecommerce-api/routes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("✅ Starting application...")

	// ✅ Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️  No .env file found. Using environment variables set by Render or OS.")
	} else {
		log.Println("✅ .env file loaded successfully.")
	}

	// ✅ Log important environment variables
	requiredEnvVars := []string{"PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
	for _, key := range requiredEnvVars {
		val := os.Getenv(key)
		if val == "" {
			log.Printf("⚠️  WARNING: %s is not set!", key)
		} else {
			log.Printf("🔑 %s=%s", key, val)
		}
	}

	// ✅ Initialize database
	db := initDatabase()
	log.Println("✅ Database connected successfully.")

	// ✅ Auto-migrate tables
	if err := db.AutoMigrate(&models.User{}, &models.Product{}, &models.Category{}); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}
	log.Println("✅ Database tables migrated successfully.")

	// ✅ Set up Gin router
	r := gin.Default()

	// ✅ Serve uploaded images
	r.Static("/uploads", "./uploads")

	// ✅ Set up routes
	routes.SetupRoutes(r, db)
	log.Println("✅ Routes initialized.")

	// ✅ Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback default port
		log.Println("⚠️  PORT not set. Using default port 8080.")
	}
	log.Printf("🚀 Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// func initDatabase() *gorm.DB {
// 	host := os.Getenv("DB_HOST")
// 	port := os.Getenv("DB_PORT")
// 	user := os.Getenv("DB_USER")
// 	password := os.Getenv("DB_PASSWORD")
// 	dbname := os.Getenv("DB_NAME")

// 	dsn := fmt.Sprintf(
// 		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		host, port, user, password, dbname,
// 	)

// 	log.Printf("🔗 Connecting to database: %s@%s:%s/%s", user, host, port, dbname)
// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("❌ Failed to connect to database: %v", err)
// 	}

// 	return db
// }

func initDatabase() *gorm.DB {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL is not set")
    }
    log.Printf("🔗 Connecting to database via DATABASE_URL")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("❌ Failed to connect to database: %v", err)
    }
    return db
}
