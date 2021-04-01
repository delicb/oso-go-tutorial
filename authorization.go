package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/osohq/go-oso"
	"go.uber.org/multierr"
)

//go:embed authorization.polar
var osoPolicy string

type AuthManager struct {
	engine oso.Oso
}

func NewAuthorizationEngine() (*AuthManager, error) {
	engine, err := oso.NewOso()
	if err != nil {
		return nil, fmt.Errorf("creating OSO engine: %w", err)
	}

	// register types used in policies
	err = multierr.Combine(
		// http types
		engine.RegisterClass(reflect.TypeOf(http.Request{}), nil),

		// domain types
		engine.RegisterClass(reflect.TypeOf(User{}), nil),
		engine.RegisterClass(reflect.TypeOf(Organization{}), nil),
		engine.RegisterClass(reflect.TypeOf(Expense{}), nil),

		// library
		engine.RegisterClass(reflect.TypeOf(Lib{}), nil),
	)

	if err != nil {
		return nil, fmt.Errorf("registering classes failed: %w", err)
	}

	// load policy
	if err := engine.LoadString(osoPolicy); err != nil {
		return nil, fmt.Errorf("loading policies: %w", err)
	}

	return &AuthManager{engine}, nil
}

func (e *AuthManager) Authorize(actor, action, resource interface{}) bool {
	allowed, err := e.engine.IsAllowed(actor, action, resource)
	// if we got any error, we interpret that as not-authorized, but we log an error for debugging
	// since in normal operation we should get no-error and true/false
	if err != nil {
		log.Printf("authorization resolution error: %v", err)
		return false
	}
	return allowed
}


// Lib holds utility functions that might be useful for evaluating policies
type Lib struct{}

func (_ Lib) Split(what string, separator string) []string {
	return strings.Split(what, separator)
}
