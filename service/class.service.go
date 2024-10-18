package service

import (
	"Lab6/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

// CreateClass creates a new class
// @Summary Create a class
// @Description Create a new class
// @Tags classes
// @Accept json
// @Produce json
// @Param class body models.Class true "Class Data"
// @Success 201 {object} models.Class
// @Router /classes [post]
func CreateClass(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	var class models.Class
	if err := c.ShouldBindJSON(&class); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&class).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	redisClient.Del(ctx, "classes")

	c.JSON(http.StatusCreated, gin.H{"class": class})
}

// GetClasses retrieves all classes
// @Summary Get all classes
// @Tags classes
// @Produce json
// @Success 200 {array} models.Class
// @Router /classes [get]
func GetClasses(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	var classes []models.Class
	val, err := redisClient.Get(ctx, "classes").Result()
	if errors.Is(err, redis.Nil) {
		if err := db.Preload("Students").Find(&classes).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		jsonData, _ := json.Marshal(classes)
		redisClient.Set(ctx, "classes", jsonData, 10*time.Minute) // Cache for 10 minutes
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving from cache"})
		return
	} else {
		if err := json.Unmarshal([]byte(val), &classes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal classes from cache"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"classes": classes})
}

// GetClassByID retrieves a class by ID
// @Summary Get a class by ID
// @Tags classes
// @Produce json
// @Param id path int true "Class ID"
// @Success 200 {object} models.Class
// @Router /classes/{id} [get]
func GetClassByID(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	var class models.Class
	id := c.Param("id")
	val, err := redisClient.Get(ctx, "class:"+id).Result()
	if errors.Is(err, redis.Nil) {
		// Cache miss, fetch from database
		if err := db.Preload("Students").First(&class, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
			return
		}
		jsonData, _ := json.Marshal(class)
		redisClient.Set(ctx, "class:"+id, jsonData, 10*time.Minute) // Cache for 10 minutes
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving from cache"})
		return
	} else {
		if err := json.Unmarshal([]byte(val), &class); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal class from cache"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"class": class})
}

// UpdateClass updates a class by ID
// @Summary Update a class
// @Tags classes
// @Accept json
// @Produce json
// @Param id path int true "Class ID"
// @Param class body models.Class true "Class Data"
// @Success 200 {object} models.Class
// @Router /classes/{id} [put]
func UpdateClass(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	var class models.Class
	var requestData struct {
		Name     string `json:"name"`
		Teacher  string `json:"teacher"`
		Capacity int    `json:"capacity"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	id := c.Param("id")
	if err := db.Preload("Students").First(&class, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
		return
	}
	currentStudentCount := len(class.Students)
	if requestData.Capacity < currentStudentCount {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cannot reduce capacity below current number of students (%d)", currentStudentCount)})
		return
	}
	class.Name = requestData.Name
	class.Teacher = requestData.Teacher
	class.Capacity = requestData.Capacity
	if err := db.Save(&class).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update class"})
		return
	}
	redisClient.Del(ctx, "class:"+id, "classes")
	c.JSON(http.StatusOK, gin.H{"class": class})
}

// DeleteClass deletes a class by ID
// @Summary Delete a class
// @Tags classes
// @Produce json
// @Param id path int true "Class ID"
// @Success 200 {object} gin.H
// @Router /classes/{id} [delete]
func DeleteClass(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	id := c.Param("id")
	var class models.Class
	if err := db.Preload("Students").First(&class, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
		return
	}
	if err := db.Delete(&class).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete class"})
		return
	}
	redisClient.Del(ctx, "class:"+id, "classes")
	c.JSON(http.StatusOK, gin.H{"message": "Class deleted successfully"})
}
