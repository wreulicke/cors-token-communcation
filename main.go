package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
)

func NewAuthorizationServer() *http.Server {
	b := []byte("session-key")
	store := sessions.NewCookieStore(b)
	store.Options.SameSite = http.SameSiteLaxMode
	authorizationRouter := chi.NewRouter()
	authorizationRouter.Get("/user-profile", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Access-Control-Allow-Origin", "http://localhost:8081")
		res.Header().Add("Access-Control-Allow-Credentials", "true")
		session, err := store.Get(req, "test-session")
		if err != nil {
			res.WriteHeader(500)
			return
		}
		if session.IsNew {
			res.WriteHeader(401)
			return
		}
		err = json.NewEncoder(res).Encode(map[string]string{
			"name": "test",
		})
		if err != nil {
			log.Println(err)
			res.WriteHeader(500)
		}
	})
	authorizationRouter.Post("/login", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Access-Control-Allow-Origin", "http://localhost:8081")
		res.Header().Add("Access-Control-Allow-Credentials", "true")
		name := req.FormValue("name")
		password := req.FormValue("password")
		if name == "admin" && password == "admin" {
			sess := sessions.NewSession(store, "test-session")
			sess.Values["authorized"] = true
			sess.Values["token"] = "mytoken"
			sess.Options.SameSite = http.SameSiteLaxMode
			err := store.Save(req, res, sess)
			if err != nil {
				log.Println(err)
				res.WriteHeader(500)
				return
			}
			res.WriteHeader(200)
			err = json.NewEncoder(res).Encode(map[string]string{
				"name": "test",
			})
			if err != nil {
				log.Println(err)
				res.WriteHeader(500)
				return
			}
		} else {
			res.WriteHeader(401)
		}
	})
	authorizationRouter.Handle("/*", http.FileServer(http.Dir("authorization")))
	return &http.Server{
		Handler: authorizationRouter,
		Addr:    ":8080",
	}
}

func NewFrontendServer() *http.Server {
	frontendRouter := http.NewServeMux()
	frontendRouter.Handle("/", http.FileServer(http.Dir("frontend")))
	return &http.Server{
		Handler: frontendRouter,
		Addr:    ":8081",
	}
}

func main() {
	authorizationServer := NewAuthorizationServer()
	frontendServer := NewFrontendServer()

	wg := &sync.WaitGroup{}
	startServer(wg, authorizationServer)
	startServer(wg, frontendServer)
	log.Println("start authorization server http://localhost:8080")
	log.Println("start frontend server http://localhost:8081")
	wg.Wait()
}

func startServer(wg *sync.WaitGroup, s *http.Server) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.ListenAndServe()
	}()
}
