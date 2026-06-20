package models

import (
	"testing"
)

func TestColorStruct(t *testing.T) {
	c := Color{
		ID:   1,
		Name: "Red",
		Hex:  "#FF0000",
	}
	if c.Name != "Red" {
		t.Error("Expected name Red")
	}
	if c.Hex != "#FF0000" {
		t.Error("Expected hex #FF0000")
	}
}

func TestShapeStruct(t *testing.T) {
	s := Shape{
		ID:      1,
		Name:    "Circle",
		Type:    "circle",
		ColorID: 1,
	}
	if s.Type != "circle" {
		t.Error("Expected type circle")
	}
	if s.ColorID != 1 {
		t.Error("Expected ColorID 1")
	}
}

func TestAdminStruct(t *testing.T) {
	a := Admin{
		ID:       1,
		Email:    "admin@admin.ru",
		Password: "111111",
	}
	if a.Email != "admin@admin.ru" {
		t.Error("Expected email admin@admin.ru")
	}
}