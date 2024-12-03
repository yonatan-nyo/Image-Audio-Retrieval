package main

import (
	"bos/pablo/models"
	"bos/pablo/routes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	dsn := dsn()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	// Enable GORM logging
	db = db.Debug()

	// Auto migrate schema
	models.AutoMigrateAll(db)

	// Initialize gin router
	router := gin.Default()

	// Cors for development
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:4000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Setup Routes
	routes.SetupRouter(router, db)

	// Run server
	router.Run(":4001")
}

func dsn() string {
	dbname := os.Getenv("DATABASE_NAME")
	username := os.Getenv("DATABASE_USERNAME")
	password := os.Getenv("DATABASE_PASSWORD")
	hostname := os.Getenv("DATABASE_SERVER")
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=Asia/Shanghai", hostname, username, password, dbname)
}
