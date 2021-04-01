module github.com/delicb/oso-go-tutorial

go 1.16

require (
	github.com/mattn/go-sqlite3 v1.14.6 // indirect
	github.com/osohq/go-oso v0.11.3 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	gorm.io/driver/sqlite v1.1.4 // indirect
	gorm.io/gorm v1.21.6 // indirect
)

replace github.com/osohq/go-oso => ../oso/languages/go
