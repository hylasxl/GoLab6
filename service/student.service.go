package service

import (
	"Lab6/models"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// CreateStudent creates a new student
// @Summary Create a student
// @Description Create a new student in a specific class
// @Tags students
// @Accept json
// @Produce json
// @Param student body models.Student true "Student Data"
// @Success 201 {object} models.Student
// @Router /students [post]
func CreateStudent(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	var student models.Student
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	var class models.Class
	if err := db.First(&class, student.ClassID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
		return
	}
	var currentStudentCount int64
	db.Model(&models.Student{}).Where("class_id = ?", student.ClassID).Count(&currentStudentCount)
	if currentStudentCount >= int64(class.Capacity) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add student: class is full"})
		return
	}
	if err := db.Create(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student"})
		return
	}
	if err := redisClient.Set(ctx, "student:"+strconv.Itoa(int(student.ID)), student, 10*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cache student"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"student": student})
}

// GetAllStudents retrieves all students
// @Summary Get all students
// @Tags students
// @Produce json
// @Success 200 {array} models.Student
// @Router /students [get]
func GetAllStudents(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	var students []models.Student

	// Check cache first
	val, err := redisClient.Get(ctx, "students").Result()
	if errors.Is(err, redis.Nil) {
		if err := db.Find(&students).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve students"})
			return
		}
		redisClient.Set(ctx, "students", students, 10*time.Minute) // Cache for 10 minutes
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving from cache"})
		return
	} else {
		if err := json.Unmarshal([]byte(val), &students); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal students from cache"})
			return
		}
	}

	c.JSON(http.StatusOK, students)
}

// GetStudentByID retrieves a student by ID
// @Summary Get a student by ID
// @Tags students
// @Produce json
// @Param id path int true "Student ID"
// @Success 200 {object} models.Student
// @Router /students/{id} [get]
func GetStudentByID(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	var student models.Student
	id := c.Param("id")
	val, err := redisClient.Get(ctx, "student:"+id).Result()
	if errors.Is(err, redis.Nil) {
		if err := db.First(&student, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
			return
		}
		redisClient.Set(ctx, "student:"+id, student, 10*time.Minute) // Cache for 10 minutes
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving from cache"})
		return
	} else {
		if err := json.Unmarshal([]byte(val), &student); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal student from cache"})
			return
		}
	}

	c.JSON(http.StatusOK, student)
}

// UpdateStudent updates a student by ID
// @Summary Update a student
// @Tags students
// @Accept json
// @Produce json
// @Param id path int true "Student ID"
// @Param student body models.Student true "Student Data"
// @Success 200 {object} models.Student
// @Router /students/{id} [put]
func UpdateStudent(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	var student models.Student
	id := c.Param("id")
	if err := db.First(&student, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	if err := db.Save(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student"})
		return
	}
	redisClient.Set(ctx, "student:"+strconv.Itoa(int(student.ID)), student, 10*time.Minute)
	c.JSON(http.StatusOK, gin.H{"student": student})
}

// DeleteStudent deletes a student by ID
// @Summary Delete a student
// @Tags students
// @Produce json
// @Param id path int true "Student ID"
// @Router /students/{id} [delete]
func DeleteStudent(c *gin.Context, db *gorm.DB, redisClient *redis.Client) {
	id := c.Param("id")
	var student models.Student
	if err := db.First(&student, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}
	if err := db.Delete(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete student"})
		return
	}
	redisClient.Del(ctx, "student:"+id)
	c.JSON(http.StatusOK, gin.H{"message": "Student deleted successfully"})
}
