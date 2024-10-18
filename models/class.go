package models

import "gorm.io/gorm"

// Class represents the class entity
// @Description Class model representing a class in the system
type Class struct {
	gorm.Model
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	// Name is the class name
	// @Description The name of the class
	// @Example Math 101
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	// Teacher is the name of the teacher for the class
	// @Description The name of the teacher for the class
	// @Example Jane Smith
	Teacher string `gorm:"type:varchar(255);not null" json:"teacher"`
	// Capacity is the maximum number of students the class can hold
	// @Description The capacity of the class
	// @Example 30
	Capacity int `gorm:"not null" json:"capacity"`
	// Students is the list of students enrolled in the class
	// @Description A list of students enrolled in the class
	Students []Student `gorm:"foreignKey:ClassID" json:"students"`
}
