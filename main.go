package main

import (
	"fmt"
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
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db := initDatabase()
	db.AutoMigrate(&models.User{}, &models.Product{})

	r := gin.Default()

	// ✅ Serve uploaded images from /uploads path
	r.Static("/uploads", "./uploads")

	// Set up routes
	routes.SetupRoutes(r, db)

	// Start server
	r.Run(":" + os.Getenv("PORT"))
}


func initDatabase() *gorm.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	return db
}
