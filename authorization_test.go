package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func getManager(t *testing.T) *AuthManager {
	t.Helper()
	manager, err := NewAuthManager()
	if err != nil {
		t.Fatalf("failed to create auth manager: %v", err)
		return nil
	}
	return manager
}

type httpAuthRequest struct {
	expectedAllow bool
	user          User
	action        string
	path          string
	// request       *http.Request
}

func TestAuthByHTTPPath(t *testing.T) {
	manager := getManager(t)

	data := []httpAuthRequest{
		{
			true,
			User{},
			"GET",
			"/",
		},
		{
			false,
			User{},
			"POST",
			"/",
		},
		{
			true,
			User{},
			"GET",
			"/whoami",
		},
		{
			false,
			User{},
			"PUT",
			"/whoami",
		},
		{
			false,
			User{},
			"GET",
			"/random",
		},
		{
			true,
			User{},
			"GET",
			"/expenses/1",
		},
		{
			true,
			User{},
			"GET",
			"/expenses",
		},
		{
			false,
			User{},
			"PUT",
			"/expenses/1",
		},
		{
			false,
			User{},
			"PUT",
			"/expenses/submit",
		},
		{
			true,
			User{Email: "test@example.com"}, // make user authenticated
			"PUT",
			"/expenses/submit",
		},
		{
			false,
			User{Email: "test@example.com"}, // make user authenticated
			"POST",
			"/expenses/submit",
		},
	}

	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("%s on %s", d.action, d.path), func(t *testing.T) {
			r := &http.Request{URL: &url.URL{Path: d.path}}
			allow := manager.Authorize(d.user, d.action, r)
			if allow != d.expectedAllow {
				t.Errorf("got auth resolution %v, expected %v", allow, d.expectedAllow)
			}
		})
	}
}

type expenseRequest struct {
	expectedAllow bool
	user          User
	action        string
	expense       Expense
}

func TestExpenseAuth(t *testing.T) {
	manager := getManager(t)

	data := []expenseRequest{
		{
			true,
			User{ID: 1},
			"read",
			Expense{ID: 1, UserID: 1},
		},
		{
			false,
			User{ID: 1},
			"write",
			Expense{ID: 1, UserID: 1},
		},
		{
			false,
			User{ID: 1},
			"read",
			Expense{ID: 1, UserID: 2},
		},
	}

	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("user %d - %s - expense %d", d.user.ID, d.action, d.expense.ID), func(t *testing.T) {
			allow := manager.Authorize(d.user, d.action, d.expense)
			if allow != d.expectedAllow {
				t.Errorf("got auth resolution %v, expected %v", allow, d.expectedAllow)
			}
		})
	}

}
