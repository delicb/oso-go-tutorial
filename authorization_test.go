package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func getManager(t *testing.T) *authManager {
	t.Helper()
	// load production policy, since that is the one we want to test here
	// this also checks that authorization.polar is in a good shape, since during
	// the build it is only embedded
	policy, err := os.ReadFile("./authorization.polar")
	if err != nil {
		t.Fatalf("failed to read authorization.polar policy file")
		return nil
	}
	manager, err := NewAuthorizer(string(policy))
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

type organizationsRequest struct {
	expectedAllow bool
	user          User
	action        string
	organization  Organization
}

func TestOrganizationsAuth(t *testing.T) {
	manager := getManager(t)

	data := []organizationsRequest{
		{
			true,
			User{ID: 1, OrganizationID: 1},
			"read",
			Organization{ID: 1, Name: "org"},
		},
		{
			false,
			User{ID: 1, OrganizationID: 1},
			"write",
			Organization{ID: 1, Name: "org"},
		},
		{
			false,
			User{ID: 1, OrganizationID: 2},
			"write",
			Organization{ID: 1, Name: "org"},
		},
	}

	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("user %d - %s - organiation %d", d.user.ID, d.action, d.organization.ID), func(t *testing.T) {
			allow := manager.Authorize(d.user, d.action, d.organization)
			if allow != d.expectedAllow {
				t.Errorf("got auth resolution %v, expected %v", allow, d.expectedAllow)
			}
		})
	}

}
