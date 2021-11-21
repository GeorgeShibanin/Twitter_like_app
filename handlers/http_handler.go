package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"twitterLikeHW/generator"
	"twitterLikeHW/storage"
)

var authorIdPattern = regexp.MustCompile(`[0-9a-f]+`)

type HTTPHandler struct {
	StorageMu  sync.RWMutex
	Storage    storage.Storage
	StorageOld map[storage.PostId]storage.Post
}

type PutAllPostsResponseData struct {
	Posts    []storage.Post `json:"posts"`
	NextPage storage.PostId `json:"nextPage"`
}
type PutAllPostsResponseNoNext struct {
	Posts []storage.Post `json:"posts"`
}
type PutRequestData struct {
	Text storage.Text `json:"text"`
}

func HandleRoot(rw http.ResponseWriter, r *http.Request) {
	_, err := rw.Write([]byte("Hello from server"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rw.Header().Set("Content-Type", "plain/text")
}

func HandlePing(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

func (h *HTTPHandler) HandleCreatePost(rw http.ResponseWriter, r *http.Request) {
	tokenHeader := r.Header.Get("System-Design-User-Id")
	if tokenHeader == "" || !authorIdPattern.MatchString(tokenHeader) {
		http.Error(rw, "problem with token", http.StatusUnauthorized)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	var post PutRequestData
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	storageType := os.Getenv("STORAGE_MODE")
	var newPost storage.Post

	if storageType == "inmemory" {
		newId, _ := generator.GenerateBase64ID(6)
		//newId = newId + "G"
		newPost = storage.Post{
			Id:        storage.PostId(newId + "G"),
			Text:      storage.Text(post.Text),
			AuthorId:  storage.UserId(tokenHeader),
			CreatedAt: storage.Time(time.Now().UTC().Format("2006-01-02T15:04:05.000Z")),
		}
		h.StorageMu.Lock()
		h.StorageOld[newPost.Id] = newPost
		h.StorageMu.Unlock()
	} else {
		newPost, err = h.Storage.PutPost(r.Context(), storage.Text(post.Text), storage.UserId(tokenHeader))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
	}

	rawResponse, _ := json.Marshal(newPost)
	_, err = rw.Write(rawResponse)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleGetPosts(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	postId := strings.Trim(r.URL.Path, "/api/v1/posts/")
	Id := storage.PostId(postId)

	storageType := os.Getenv("STORAGE_MODE")
	var postText storage.Post
	var err error
	var found bool

	if storageType == "inmemory" {
		h.StorageMu.RLock()
		postText, found = h.StorageOld[Id]
		h.StorageMu.RUnlock()
		if !found {
			http.NotFound(rw, r)
			return
		}
	} else {
		postText, err = h.Storage.GetPostById(r.Context(), storage.PostId(postId))
		if err != nil {
			http.NotFound(rw, r)
			return
		}
	}

	rawResponse, _ := json.Marshal(postText)
	_, err = rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, "Поста с указанным идентификатором не существует", http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleGetUserPosts(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	pagetoken := storage.PageToken{Token: r.URL.Query().Get("page")}
	size := r.URL.Query().Get("size")
	sizepage, _ := strconv.Atoi(size)
	if size == "" {
		sizepage = 10
	} else {
		if sizepage < 0 {
			http.Error(rw, "Invalid size", http.StatusBadRequest)
			return
		}
		if sizepage > 100 {
			http.Error(rw, "Invalid size", http.StatusBadRequest)
			return
		}
	}

	userId := strings.TrimSuffix(r.URL.Path, "/posts")
	Id := strings.TrimPrefix(userId, "/api/v1/users/")

	storageType := os.Getenv("STORAGE_MODE")
	var finalResponse []storage.Post
	var err error

	if storageType == "inmemory" {
		h.StorageMu.RLock()
		for _, value := range h.StorageOld { //итерируемся по мапу постов и выводим пост если совпал айдишник автора и юзера в запросе
			if value.AuthorId == storage.UserId(Id) {
				finalResponse = append(finalResponse, value)
			}
		}
		h.StorageMu.RUnlock()
	} else {
		finalResponse, err = h.Storage.GetPostsByUser(r.Context(), storage.UserId(Id))
		if err != nil {
			http.Error(rw, "YOU SUCK AT DB LOSER", http.StatusBadRequest)
			return
		}

	}

	sort.Slice(finalResponse, func(i, j int) bool {
		layout := "2006-01-02T15:04:05.000Z"
		first, _ := time.Parse(layout, string(finalResponse[i].CreatedAt))
		second, _ := time.Parse(layout, string(finalResponse[j].CreatedAt))
		return first.After(second)
	})

	rawResponse := PutAllPostsResponseData{}
	startPage := 0
	for i, value := range finalResponse {
		if pagetoken.Token != "" && value.Id == storage.PostId(pagetoken.Token) {
			startPage = i
		}
	}
	if pagetoken.Token != "" && startPage == 0 {
		http.Error(rw, "InvalidPageToken", http.StatusBadRequest)
		return
	}
	finalResponse = finalResponse[startPage:]
	returnResponse, _ := json.Marshal("")
	if len(finalResponse) >= sizepage+1 {
		rawResponse.Posts = finalResponse[0:sizepage]
		rawResponse.NextPage = finalResponse[sizepage].Id
		returnResponse, _ = json.Marshal(rawResponse)
	} else {
		returnResponse, _ = json.Marshal(PutAllPostsResponseNoNext{Posts: finalResponse})
	}
	if len(finalResponse) == 0 {
		returnResponse, _ = json.Marshal(PutAllPostsResponseNoNext{Posts: make([]storage.Post, 0)})
	}

	_, err = rw.Write(returnResponse)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}
