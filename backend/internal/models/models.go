package models

type Color struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Name   string `json:"name" gorm:"unique;not null"`
	Hex    string `json:"hex"`
	Shapes []Shape `json:"shapes,omitempty"`
}

type Shape struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Name    string `json:"name" gorm:"not null"`
	Type    string `json:"type"`
	ColorID uint   `json:"color_id"`
	Color   Color  `json:"color,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Admin struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique;not null"`
	Password string `json:"-" gorm:"not null"`
}