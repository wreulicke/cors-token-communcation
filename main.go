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
	authorizationRouter := chi.NewRouter()
	authorizationRouter.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			v := req.URL.Query()
			r := v.Get("r")
			if r == "" || req.URL.Path != "/" {
				h.ServeHTTP(w, req)
				return
			}
			sess, err := store.New(req, "test-session")
			if err != nil {
				log.Println("fuck", err)
				w.WriteHeader(500)
				return
			}
			log.Println(r)
			sess.Values["r"] = r
			sess.Save(req, w)
			h.ServeHTTP(w, req)
		})
	})
	authorizationRouter.Post("/token", func(res http.ResponseWriter, req *http.Request) {
		sess, err := store.Get(req, "test-session")
		if err != nil {
			res.WriteHeader(400)
			return
		}
		if sess.IsNew {
			res.WriteHeader(400)
			return
		}
		token := sess.Values["token"].(string)
		if err := json.NewEncoder(res).Encode(map[string]string{
			"token": token,
		}); err != nil {
			log.Println(err)
		}
	})

	authorizationRouter.Options("/user-profile", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Access-Control-Allow-Origin", "*")
		res.Header().Add("Access-Control-Allow-Headers", "Authorization")
		res.WriteHeader(200)
	})
	authorizationRouter.Get("/user-profile", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Access-Control-Allow-Origin", "*")
		t := req.Header.Get("Authorization")
		log.Println(t)
		if t != "mytoken" {
			res.WriteHeader(401)
			return
		}
		err := json.NewEncoder(res).Encode(map[string]string{
			"name": "test",
		})
		if err != nil {
			log.Println(err)
			res.WriteHeader(500)
		}
	})
	authorizationRouter.Post("/login", func(res http.ResponseWriter, req *http.Request) {
		name := req.FormValue("name")
		password := req.FormValue("password")
		if name == "admin" && password == "admin" {
			oldSession, err := store.Get(req, "test-session")
			if err != nil {
				res.WriteHeader(500)
				return
			}
			sess := sessions.NewSession(store, "test-session")
			sess.Values["authorized"] = true
			sess.Values["token"] = "mytoken"
			err = store.Save(req, res, sess)
			if err != nil {
				log.Println(err)
				res.WriteHeader(500)
				return
			}
			v, ok := oldSession.Values["r"].(string)
			if ok {
				http.Redirect(res, req, v, 301)
			} else {
				http.Redirect(res, req, "http://localhost:8080/", 301)
			}
		} else {
			http.Redirect(res, req, "http://localhost:8080/", 401)
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
