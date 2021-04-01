module github.com/delicb/oso-go-tutorial

go 1.16

require (
	github.com/go-chi/chi/v5 v5.0.2
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/osohq/go-oso v0.11.3
	go.uber.org/multierr v1.6.0
)

replace github.com/osohq/go-oso => ../oso/languages/go
