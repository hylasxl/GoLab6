package models

import "gorm.io/gorm"

// Student represents the student entity
// @Description Student model representing a student in the system
type Student struct {
	gorm.Model
	// Name is the student's name
	// @Description The name of the student
	// @Example John Doe
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	// Age is the student's age
	// @Description The age of the student
	// @Example 20
	Age int `gorm:"default:1" json:"age"`
	// Email is the student's email
	// @Description The email of the student
	// @Example john.doe@example.com
	Email string `gorm:"type:varchar(255);not null" json:"email"`
	// ClassID is the ID of the class the student belongs to
	// @Description The ID of the class the student belongs to
	// @Example 101
	ClassID uint `gorm:"default:0" json:"class_id"`
	// Class is the class that the student is enrolled in
	// @Description The class that the student is enrolled in
	Class Class `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"class"`
}
