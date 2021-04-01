package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// HTTPServer provides HTTP endpoints functionality
type HTTPServer struct {
	db   DBManager
	auth Authorizer
}

// NewHTTPHandler returns handler that serves all HTTP endpoints with
// authentication and authorization built-in.
func NewHTTPHandler(db DBManager, auth Authorizer) http.Handler {
	server := &HTTPServer{
		db:   db,
		auth: auth,
	}

	mux := chi.NewMux()
	mux.Use(Authenticate(db))
	mux.Use(Authorize(auth))

	mux.Put(`/expenses/submit`, server.createExpense)
	mux.Get(`/expenses/{id:[0-9]+}`, server.getExpense)
	mux.Get(`/organizations/{id:[0-9]+}`, server.getOrganization)
	mux.Get(`/whoami`, server.whoami)
	mux.Get("/", server.hello)

	return mux
}

func (h *HTTPServer) hello(w http.ResponseWriter, r *http.Request) {
	user := UserFromRequest(r)
	if !user.IsAuthenticated() {
		_, _ = fmt.Fprint(w, "hello guest user")
	} else {
		_, _ = fmt.Fprintf(w, "hello %s", user.Email)
	}
}

func (h *HTTPServer) whoami(w http.ResponseWriter, r *http.Request) {
	user := UserFromRequest(r)
	if !user.IsAuthenticated() {
		_, _ = fmt.Fprint(w, "guest user")
		return
	}
	organization, err := h.db.OrganizationByID(user.OrganizationID)
	if err != nil {
		http.Error(w, "failed to fetch organization", http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, "You are %s, the %s at %s", user.Email, user.Title, organization.Name)
}

func (h *HTTPServer) getExpense(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid expense ID", http.StatusBadRequest)
		return
	}

	expense, err := h.db.ExpenseByID(id)
	if err != nil {
		http.Error(w, "unable to find expense", http.StatusNotFound)
		return
	}

	if allowed := h.auth.Authorize(UserFromRequest(r), "read", expense); !allowed {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	payload, err := json.Marshal(expense)
	if err != nil {
		http.Error(w, "failed to marshal json", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(payload)
}

func (h *HTTPServer) createExpense(w http.ResponseWriter, r *http.Request) {
	// read body and parse json into a struct
	bodyReader := io.LimitReader(r.Body, 1024*1024)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		http.Error(w, "unable to read provided body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var expense Expense
	if err := json.Unmarshal(body, &expense); err != nil {
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		log.Println("json parse error", err)
		return
	}

	// verify if userID is provided in payload, since it should not be, we want to
	// set current user ID
	if expense.UserID != 0 {
		http.Error(w, "setting user ID for expense not allowed", http.StatusBadRequest)
		return
	}
	expense.UserID = UserFromRequest(r).ID

	if ex, err := h.db.CreateExpense(expense); err != nil {
		http.Error(w, "failed saving expense", http.StatusInternalServerError)
		return
	} else {
		// redirect to expense that was just created
		http.Redirect(w, r, fmt.Sprintf("/expenses/%d", ex.ID), http.StatusTemporaryRedirect)
		return
	}
}

func (h *HTTPServer) getOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid organization ID", http.StatusBadRequest)
		return
	}

	organization, err := h.db.OrganizationByID(id)
	if err != nil {
		http.Error(w, "unable to find organization", http.StatusNotFound)
		return
	}

	if allowed := h.auth.Authorize(UserFromRequest(r), "read", organization); !allowed {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	payload, err := json.Marshal(organization)
	if err != nil {
		http.Error(w, "failed to marshal json", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(payload)
}

// middlewares

// unique type to use for context keys for authnz purposes
type ctxKey string

// context key for user in context
const userKey ctxKey = "user"

// UserFromContext returns user that is attached to a context.
// Note that user instance is always returned, but it might be empty
// for non-authorized users. User should call IsAuthenticated method on
// user in order to check if user is authenticated.
func UserFromRequest(r *http.Request) User {
	val := r.Context().Value(userKey)
	if u, ok := val.(User); ok {
		return u
	}
	return User{}
}

// Authenticate checks if user provided in "User" header exists
// and attaches instances of a user to context for next handler
// in chain to use
func Authenticate(db DBManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userEmail := r.Header.Get("user")
			userFromDB, err := db.UserByEmail(userEmail)
			if err == nil {
				ctx := context.WithValue(r.Context(), userKey, userFromDB)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Authenticate checks if user provided in "User" header exists
// and attaches instances of a user to context for next handler
// in chain to use
func Authorize(auth Authorizer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromRequest(r)
			allowed := auth.Authorize(user, r.Method, r)
			if !allowed {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			// allow access
			next.ServeHTTP(w, r)
		})
	}
}
