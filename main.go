package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/osohq/go-oso"
)

type WebApp struct {
	db *DBManager
}

func (a *WebApp) helloWorld(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	fmt.Fprintf(w, "got user %s", user.Email)
}

func (a *WebApp) createExpense(w http.ResponseWriter, r *http.Request) {
	if ex, err := a.db.CreateExpense(2, 100, "some description"); err != nil {
		fmt.Fprintf(w, "failed to create expense: %v", err)
	} else {
		fmt.Fprintf(w, "created expense: %v", ex)
	}
}

type Middleware func(handler http.Handler) http.Handler

// func Authorize(db *sql.DB) Middleware {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			userEmail := r.Header.Get("user")
// 			authUser, err := getUser(db, userEmail)
// 			fmt.Println("user from database: ", authUser)
// 			fmt.Println("error from database: ", err)
// 			if err == nil {
// 				ctx := context.WithValue(r.Context(), "user", authUser)
// 				r = r.WithContext(ctx)
// 			}
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

type ctxKey string

const userKey ctxKey = "user"

func UserFromContext(ctx context.Context) User {
	val := ctx.Value(userKey)
	if u, ok := val.(User); ok {
		return u
	}
	return User{} // maybe return Guest user?
}

// Authenticate checks if user provided in "User" header exists
// and attaches instances of a user to context for next handler
// in chain to use
func Authenticate(db *DBManager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("in authenticate")
			userEmail := r.Header.Get("user")
			log.Println("user email:", userEmail)
			userFromDB, err := db.UserByEmail(userEmail)
			log.Printf("user from db: %v", userFromDB)
			log.Println("db error:", err)
			if err == nil {
				ctx := context.WithValue(r.Context(), userKey, userFromDB)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Authorize(oso oso.Oso) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("in authorize")
			user := UserFromContext(r.Context())
			log.Printf("got user: %v", user)
			allowed, err := oso.IsAllowed(user, r.Method, r)
			log.Println("authorization context:", user, r.Method, r.URL.Path)
			log.Println("result from authorization:", allowed, err)
			if err != nil {
				// this should not fail, be careful and do not allow action, but log error
				log.Printf("authorization resolution failed: %v", err)
				httpForbid(w)
				return
			}
			if !allowed {
				httpForbid(w)
				return
			}

			// allow access
			next.ServeHTTP(w, r)
		})
	}
}

func httpForbid(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprintf(w, "forbidden")
}

//go:embed authorization.polar
var osoPolicy string

//go:embed schema.sql
var dbSchema string

func main() {
	osoEngine, err := prepareOso()
	if err != nil {
		panic(err)
	}
	fmt.Println(osoEngine)

	db, err := NewDBManager("expenses.sqlite")
	if err != nil {
		panic(err)
	}
	fmt.Println(db)

	// prepare HTTP server
	mux := http.NewServeMux()
	app := &WebApp{db}

	mux.Handle("/", chain(
		http.HandlerFunc(app.helloWorld),
		Authenticate(db),
		Authorize(osoEngine),
	))
	mux.Handle("/expense", chain(
		http.HandlerFunc(app.createExpense),
		Authenticate(db),
		Authorize(osoEngine),
	))

	// run server
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", mux))

}

func chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	h := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
