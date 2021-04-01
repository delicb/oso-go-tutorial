package main

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DBManager interface {
	UserByID(int) (User, error)
	UserByEmail(string) (User, error)
	OrganizationByID(int) (Organization, error)
	ExpenseByID(int) (Expense, error)
	CreateExpense(Expense) (Expense, error)
}

type dBManager struct {
	db *sql.DB
}

// NewDBManager returns an instance of dBManager connected to a database
// defined with provided dsn
func NewDBManager(dsn string) (*dBManager, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	return &dBManager{db}, nil
}

// rawExec is used for initialization (e.g. poor man's migration management)
// or preparing data for tests
func (m *dBManager) rawExec(sql string) error {
	if _, err := m.db.Exec(sql); err != nil {
		return fmt.Errorf("failed to execute sql: %w", err)
	}
	return nil
}

func (m *dBManager) UserByEmail(forEmail string) (User, error) {
	row := m.db.QueryRow(`SELECT id, email, title, organization_id FROM users WHERE email = ?`, forEmail)
	return m.constructUser(row)
}

func (m *dBManager) UserByID(id int) (User, error) {
	row := m.db.QueryRow(`SELECT id, email, title, organization_id FROM users WHERE id = ?`, id)
	return m.constructUser(row)
}

func (m *dBManager) constructUser(row *sql.Row) (User, error) {
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

func (m *dBManager) OrganizationByID(forID int) (Organization, error) {
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

func (m *dBManager) ExpenseByID(forID int) (Expense, error) {
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
			ID:          id,
			UserID:      userID,
			Amount:      amount,
			Description: description,
		}, nil
	default:
		return Expense{}, err // unknown error, just propagate
	}
}

func (m *dBManager) CreateExpense(in Expense) (Expense, error) {
	// TODO: error handling
	tx, err := m.db.Begin()
	if err != nil {
		tx.Rollback()
		return Expense{}, err
	}
	res, err := tx.Exec(`INSERT INTO expenses (amount, description, user_id) VALUES (?, ?, ?)`, in.Amount, in.Description, in.UserID)
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
	in.ID = int(expenseID)
	return in, nil
}

// build time guarantee that dbManager implement DBManager
var _ DBManager = &dBManager{}
