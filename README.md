# oso: authorization as code [Demo Application]

oso allows you to write policy as code, tightly integrated
with application code, logic, and data, and provides a simple
way to enforce authorization on all requests.

## Example

This repository contains an example application using oso
to secure requests.

The application itself is a simple [Go](https://golang.org/)
web server, implementing an expenses application.

We have already partially implemented the core functionality;
some is left for you to complete.

## Libs
On purpose, only small number of external dependencies is used, to
keep things simple. 

For database, only sqlite3 drive is used, no ORM.

Stdlib HTTP functionalities are used, but with addition of [Chi router](https://github.com/go-chi/chi)
to make middlewares and URL parameters easier to implement. 

Of course, [OSO](https://github.com/osohq/go-oso) library is used to perform
authorization. 

Utility wise, [Uber multierr](https://github.com/uber-go/multierr) package is used. 