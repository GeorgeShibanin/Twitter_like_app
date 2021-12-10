package main

import (
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	_ "go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"os"
	"time"
	"twitterLikeHW/handlers"
	"twitterLikeHW/storage"
	"twitterLikeHW/storage/mongostorage"
	"twitterLikeHW/storage/rediscachedstorage"
)

func NewServer() *http.Server {
	r := mux.NewRouter()

	handler := &handlers.HTTPHandler{}

	storageType := os.Getenv("STORAGE_MODE")
	//storageType := "inmemory"

	if storageType == "mongo" {
		mongoUrl := os.Getenv("MONGO_URL")
		//mongoUrl := "mongodb://localhost:27017"
		mongoStorage := mongostorage.NewStorage(mongoUrl)
		handler = &handlers.HTTPHandler{
			Storage: mongoStorage,
		}
	} else if storageType == "inmemory" {
		handler = &handlers.HTTPHandler{
			StorageOld: make(map[primitive.ObjectID]*storage.PostOld),
		}
	} else if storageType == "cached" {
		mongoUrl := os.Getenv("MONGO_URL")
		mongoStorage := mongostorage.NewStorage(mongoUrl)
		redisUrl := os.Getenv("REDIS_URL")
		cachedStorage := rediscachedstorage.NewStorage(redisUrl, mongoStorage)
		handler = &handlers.HTTPHandler{
			Storage: cachedStorage,
		}
	}

	r.HandleFunc("/", handlers.HandleRoot).Methods("GET", "POST")
	r.HandleFunc("/api/v1/posts", handler.HandleCreatePost).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/posts/{postId}", handler.HandleGetPosts).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/posts/{postId}", handler.HandlePatchPosts).Methods(http.MethodPatch)
	r.HandleFunc("/api/v1/users/{userId}/posts", handler.HandleGetUserPosts).Methods(http.MethodGet)
	r.HandleFunc("/maintenance/ping", handlers.HandlePing).Methods(http.MethodGet)

	port := os.Getenv("SERVER_PORT")
	//port := "8080"
	return &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func main() {
	srv := NewServer()
	log.Printf("Start serving on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
