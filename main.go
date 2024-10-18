package main

import (
	_ "Lab6/docs"
	"Lab6/models"
	"Lab6/service"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var ctx = context.Background()

func connectRedis() *redis.Client {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Could not connect to Redis:", err)
	}

	fmt.Println("Connected to Redis successfully!")
	return client
}

func connectDB() (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	dbTimezone := os.Getenv("DB_TIMEZONE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode, dbTimezone)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to the database successfully!")
	return db, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := connectDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	redisClient := connectRedis()
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Println("Error closing Redis connection:", err)
		}
	}()

	if err = db.AutoMigrate(&models.Class{}, &models.Student{}); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	r.POST("/students", func(c *gin.Context) { service.CreateStudent(c, db, redisClient) })
	r.GET("/students", func(c *gin.Context) { service.GetAllStudents(c, db, redisClient) })
	r.GET("/students/:id", func(c *gin.Context) { service.GetStudentByID(c, db, redisClient) })
	r.PUT("/students/:id", func(c *gin.Context) { service.UpdateStudent(c, db, redisClient) })
	r.DELETE("/students/:id", func(c *gin.Context) { service.DeleteStudent(c, db, redisClient) })

	r.POST("/classes", func(c *gin.Context) { service.CreateClass(c, db, redisClient) })
	r.GET("/classes", func(c *gin.Context) { service.GetClasses(c, db, redisClient) })
	r.GET("/classes/:id", func(c *gin.Context) { service.GetClassByID(c, db, redisClient) })
	r.PUT("/classes/:id", func(c *gin.Context) { service.UpdateClass(c, db, redisClient) })
	r.DELETE("/classes/:id", func(c *gin.Context) { service.DeleteClass(c, db, redisClient) })

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
