package main

import (
	"fmt"
)

// Logged in user. Has an email address.
type User struct {
	ID             int
	Email          string
	Title          string
	OrganizationID int
}

func (u User) String() string {
	return fmt.Sprintf("<User: %s (id: %d)>", u.Email, u.ID)
}

// Anonymous user
type Guest struct{}

func (g Guest) String() string {
	return "<User: Guest>"
}

// Organization model
type Organization struct {
	ID   int
	Name string
}

func (o Organization) String() string {
	return fmt.Sprintf("<Organization: %s (id: %d)>", o.Name, o.ID)
}

// Expense model
type Expense struct {
	ID          int
	UserID      int
	Amount      int
	Description string
}

func (e Expense) String() string {
	return fmt.Sprintf("<Expense: %d (amount: %d, user: %d)>", e.ID, e.Amount, e.UserID)
}

