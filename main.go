package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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
	_ = godotenv.Load()

	// Init DB
	db := initDatabase()

	// Auto-migrate all tables
	if err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Category{},
		&models.Admin{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Banner{},
		&models.QRFile{},
	); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	// Gin setup
	r := gin.Default()

	// Allow large file uploads (1 GB)
	r.MaxMultipartMemory = 1 << 30 // 1GB

	// CORS settings
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-KEY"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Directories
	uploadsDir := "/var/www/trendybacked/uploads"
	backupDir := "/var/www/trendybacked/backup/uploads"

	// Serve uploaded images
	r.Static("/uploads", uploadsDir)

	// Setup routes
	routes.SetupRoutes(r, db)

	// Start backup routine at 2 AM daily, keep 4 days of backups
	go startDailyBackupAtFixedTime(uploadsDir, backupDir, 4*24*time.Hour, 2, 0)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// initDatabase sets up the GORM DB connection
func initDatabase() *gorm.DB {
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err != nil {
			log.Fatalf("❌ DB connection failed: %v", err)
		}
		return db
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect DB: %v", err)
	}
	return db
}

// startDailyBackupAtFixedTime backs up images daily at a fixed hour and removes old backups
func startDailyBackupAtFixedTime(srcDir, backupDir string, retention time.Duration, hour, min int) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		sleepDuration := next.Sub(now)
		log.Printf("⏳ Next image backup scheduled at: %s", next.Format("2006-01-02 15:04:05"))
		time.Sleep(sleepDuration)

		timestamp := time.Now().Format("2006-01-02_15-04-05")
		destDir := filepath.Join(backupDir, timestamp)

		if err := copyDir(srcDir, destDir); err != nil {
			log.Printf("❌ Failed to back up images: %v", err)
		} else {
			log.Printf("✅ Images backed up to %s", destDir)
		}

		cleanupOldBackups(backupDir, retention)
	}
}

// copyDir recursively copies a folder
func copyDir(src, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a single file
func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// cleanupOldBackups removes backup folders older than retention duration
func cleanupOldBackups(backupDir string, retention time.Duration) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		log.Printf("❌ Failed to read backup directory: %v", err)
		return
	}

	cutoff := time.Now().Add(-retention)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		folderPath := filepath.Join(backupDir, entry.Name())
		info, err := os.Stat(folderPath)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.RemoveAll(folderPath); err != nil {
				log.Printf("❌ Failed to remove old backup %s: %v", folderPath, err)
			} else {
				log.Printf("🗑️ Removed old backup: %s", folderPath)
			}
		}
	}
}
