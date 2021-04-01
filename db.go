package main

import (
	"database/sql"
	"fmt"
)

type DBManager struct {
	db *sql.DB
}

func NewDBManager(file string) (*DBManager, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	// prepare schema, all create table statements should have "if not exists"
	// part, so this should not fail even on multiple runs
	// TODO: maybe proper migrations support?
	if _, err := db.Exec(dbSchema); err != nil {

		return nil, fmt.Errorf("failed to create DB schema: %w", err)
	}

	return &DBManager{db}, nil
}

func (m *DBManager) UserByEmail(forEmail string) (User, error) {
	row := m.db.QueryRow(`SELECT id, email, title, organization_id FROM users WHERE email = ?`, forEmail)
	return m.constructUser(row)
}

func (m *DBManager) UserByID(id int) (User, error) {
	row := m.db.QueryRow(`SELECT id, email, title, organization_id FROM users WHERE id = ?`, id)
	return m.constructUser(row)
}

func (m *DBManager) constructUser(row *sql.Row) (User, error) {
	var id int
	var email string
	var title string
	var organizationID int

	switch err := row.Scan(&id, &email, &title, &organizationID); err {
	case sql.ErrNoRows:
		return User{}, fmt.Errorf("no user found for selected criteria")
	case nil:
		return User{
			ID:             id,
			Email:          email,
			Title:          title,
			OrganizationID: organizationID,
		}, nil
	default:
		return User{}, err // unknown error, just propagate
	}
}

func (m *DBManager) OrganizationByID(forID int) (Organization, error) {
	var id int
	var name string

	row := m.db.QueryRow(`SELECT id, name FROM organizations WHERE id = ?`, forID)

	switch err := row.Scan(&id, &name); err {
	case sql.ErrNoRows:
		return Organization{}, fmt.Errorf("no organization for ID %d", forID)
	case nil:
		return Organization{
			ID:   id,
			Name: name,
		}, nil
	default:
		return Organization{}, err // unknown error, just propagate
	}
}

func (m *DBManager) ExpenseByID(forID int) (Expense, error) {
	var id int
	var userID int
	var amount int
	var description string

	row := m.db.QueryRow(`SELECT id, user_id, amount, description FROM expenses WHERE id = ?`, forID)

	switch err := row.Scan(&id, &userID, &amount, &description); err {
	case sql.ErrNoRows:
		return Expense{}, fmt.Errorf("no organization for ID %d", forID)
	case nil:
		return Expense{
			ID:   id,
			UserID: userID,
			Amount: amount,
			Description: description,
		}, nil
	default:
		return Expense{}, err // unknown error, just propagate
	}
}

func (m *DBManager) CreateExpense(userID, amount int, description string) (Expense, error) {
	tx, err := m.db.Begin()
	if err != nil {
		tx.Rollback()
		return Expense{}, err
	}
	res, err := tx.Exec(`INSERT INTO expenses (amount, description, user_id) VALUES (?, ?, ?)`, amount, description, userID)
	if err != nil {
		tx.Rollback()
		return Expense{}, err
	}
	expenseID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return Expense{}, err
	}
	tx.Commit()
	return Expense{
		ID:          int(expenseID),
		UserID:      userID,
		Amount:      amount,
		Description: description,
	}, nil
}
