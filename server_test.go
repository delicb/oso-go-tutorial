package main

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mock authorization manager
type authMock struct {
	mock bool
}

func (m *authMock) Authorize(actor, action, resource interface{}) bool {
	return m.mock
}

// mock db manager
type dbMock struct {
	user         User
	organization Organization
	expense      Expense
	err          error
}

func (d dbMock) UserByID(i int) (User, error) {
	return d.user, d.err
}

func (d dbMock) UserByEmail(s string) (User, error) {
	return d.user, d.err
}

func (d dbMock) OrganizationByID(i int) (Organization, error) {
	return d.organization, d.err
}

func (d dbMock) ExpenseByID(i int) (Expense, error) {
	return d.expense, d.err
}

func (d dbMock) CreateExpense(expense Expense) (Expense, error) {
	panic("implement me")
}

func TestServer(t *testing.T) {
	handler := NewHTTPHandler(&dbMock{err: sql.ErrNoRows}, &authMock{true})

	server := httptest.NewServer(handler)

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to get response from server: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected status code 200, got %d", resp.StatusCode)
	}
	content, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if !strings.Contains(string(content), "hello guest") {
		t.Fatalf("got unexpected response from server: %v", string(content))
	}
}

func userRecorderHandler(out *User) http.Handler {
	return http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		*out = UserFromRequest(r)
	})
}

func TestAuthenticate_HasUser(t *testing.T) {
	data := []struct {
		name          string
		db            DBManager
		expectedEmail string
	}{
		{
			"regular user",
			dbMock{user: User{Email: "test@example.com"}},
			"test@example.com",
		},
		{
			"db error",
			dbMock{err: sql.ErrNoRows},
			"",
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			authenticator := Authenticate(d.db)
			var recordedUser User
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			authenticator(userRecorderHandler(&recordedUser)).ServeHTTP(nil, req)

			if recordedUser.Email != d.expectedEmail {
				t.Fatalf("unexpected email, expected %q, got %q", d.expectedEmail, recordedUser.Email)
			}

		})
	}
}

func statusCodeHandler(statusCode int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(statusCode)
	})
}

func TestAuthorize(t *testing.T) {
	data := []struct {
		name               string
		allow              bool
		expectedStatusCode int
	}{
		{
			"allowed",
			true,
			http.StatusOK,
		},
		{
			"not allowed",
			false,
			http.StatusForbidden,
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			authorizer := Authorize(&authMock{d.allow})
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			authorizer(statusCodeHandler(http.StatusOK)).ServeHTTP(rec, req)

			if rec.Code != d.expectedStatusCode {
				t.Fatalf("wront status code, expected %d, got %d", d.expectedStatusCode, rec.Code)
			}
		})
	}
}
