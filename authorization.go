package main

import (
	"net/http"
	"reflect"

	"github.com/osohq/go-oso"
	"go.uber.org/multierr"
)

func prepareOso() (oso.Oso, error) {
	engine, err := oso.NewOso()
	if err != nil {
		return oso.Oso{}, err
	}

	// register types
	err = multierr.Combine(
		// http types
		engine.RegisterClass(reflect.TypeOf(http.Request{}), nil),

		// domain types
		engine.RegisterClass(reflect.TypeOf(User{}), nil),
		engine.RegisterClass(reflect.TypeOf(Guest{}), nil),
		engine.RegisterClass(reflect.TypeOf(Organization{}), nil),
		engine.RegisterClass(reflect.TypeOf(Expense{}), nil),
	)
	if err != nil {
		return oso.Oso{}, err
	}

	// load policy
	if err := engine.LoadString(osoPolicy); err != nil {
		return oso.Oso{}, err
	}

	return engine, nil
}
