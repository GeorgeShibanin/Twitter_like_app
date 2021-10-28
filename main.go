package main

import (
	"Design_System/twitterLikeHW/handlers"
	_ "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	_ "strconv"
	_ "strings"
	"time"
)

func NewServer() *http.Server {
	r := mux.NewRouter()

	handler := &handlers.HTTPHandler{
		Storage: make(map[handlers.PostId]handlers.Post),
	}

	r.HandleFunc("/", handlers.HandleRoot).Methods("GET", "POST")
	r.HandleFunc("/api/v1/posts", handler.HandleCreatePost).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/posts/{postId:\\w}", handler.HandleGetPosts).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/users/{userId:\\w}/posts", handler.HandleGetUserPosts).Methods(http.MethodGet)

	return &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func main() {
	srv := NewServer()
	log.Printf("Start serving on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
