package handlers

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
	"regexp"
	"sort"
	_ "sort"
	"strconv"
	_ "strconv"
	"strings"
	_ "strings"
	"sync"
	"time"
	_ "twitterLikeHW/generator"

	//"twitterLikeHW/generator"
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
	//storageType := "inmemory"

	//var newPostOld storage.PostOld
	//var newPostMongo storage.Post
	var newPost storage.Post
	var rawResponse []byte

	if storageType == "inmemory" {
		newPost = storage.Post{
			Id:             storage.PostId(primitive.NewObjectID().Hex() + "G"),
			Text:           post.Text,
			AuthorId:       storage.UserId(tokenHeader),
			CreatedAt:      storage.ISOTimestamp(time.Now().UTC().Format("2006-01-02T15:04:05.000Z")),
			LastModifiedAt: storage.ISOTimestamp(time.Now().UTC().Format("2006-01-02T15:04:05.000Z")),
		}
		h.StorageMu.Lock()
		h.StorageOld[newPost.Id] = newPost
		h.StorageMu.Unlock()
		rawResponse, _ = json.Marshal(newPost)
	} else {
		newPost, err = h.Storage.PutPost(r.Context(), post.Text, storage.UserId(tokenHeader))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		rawResponse, _ = json.Marshal(newPost)
	}

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
	//storageType := "inmemory"

	var currentPost storage.Post
	var err error
	var found bool
	var rawResponse []byte

	if storageType == "inmemory" {
		h.StorageMu.RLock()
		//valueId, _ := primitive.ObjectIDFromHex(postId)
		currentPost, found = h.StorageOld[storage.PostId(postId)]
		h.StorageMu.RUnlock()
		if !found {
			http.NotFound(rw, r)
			return
		}
	} else {
		currentPost, err = h.Storage.GetPostById(r.Context(), Id)
		if err != nil {
			http.NotFound(rw, r)
			return
		}
	}
	rawResponse, _ = json.Marshal(currentPost)
	_, err = rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, "Поста с указанным идентификатором не существует", http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandlePatchPosts(rw http.ResponseWriter, r *http.Request) {
	postId := strings.Trim(r.URL.Path, "/api/v1/posts/")
	Id := storage.PostId(postId)
	tokenHeader := r.Header.Get("System-Design-User-Id")
	if tokenHeader == "" || !authorIdPattern.MatchString(tokenHeader) {
		http.Error(rw, "problem with token", http.StatusUnauthorized)
		return
	}
	var updatePostText PutRequestData
	err := json.NewDecoder(r.Body).Decode(&updatePostText)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.Header().Set("Content-Type", "application/json")

	storageType := os.Getenv("STORAGE_MODE")
	//storageType := "inmemory"

	var found bool
	//var updatePostOld storage.PostOld
	var updatePost storage.Post

	if storageType == "inmemory" {
		h.StorageMu.RLock()
		//valueId, _ := primitive.ObjectIDFromHex(string(Id))
		updatePost, found = h.StorageOld[Id]
		h.StorageMu.RUnlock()
		if !found {
			http.NotFound(rw, r)
			return
		}
		if updatePost.AuthorId != storage.UserId(tokenHeader) {
			http.Error(rw, "wrong user for this post", http.StatusForbidden)
			return
		}
		updatePost.LastModifiedAt = storage.ISOTimestamp(time.Now().UTC().Format("2006-01-02T15:04:05.000Z"))
		updatePost.Text = updatePostText.Text
		h.StorageMu.Lock()
		h.StorageOld[Id] = updatePost
		h.StorageMu.Unlock()

	} else {
		updatePost, err = h.Storage.GetPostById(r.Context(), Id)
		if err != nil {
			http.NotFound(rw, r)
			return
		}
		if updatePost.AuthorId != storage.UserId(tokenHeader) {
			http.Error(rw, "wrong user for this post", http.StatusForbidden)
			return
		}

		updatePost, err = h.Storage.PatchPostById(r.Context(), Id, updatePostText.Text, storage.UserId(tokenHeader))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
	}

	rawResponse, _ := json.Marshal(updatePost)
	_, err = rw.Write(rawResponse)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleGetUserPosts(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	pagetoken := r.URL.Query().Get("page")
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
	//storageType := "inmemory"

	var finalResponse []storage.Post
	rawResponse := PutAllPostsResponseData{}
	var err error

	if storageType == "inmemory" {
		h.StorageMu.RLock()
		for _, value := range h.StorageOld { //итерируемся по мапу постов и выводим пост если совпал айдишник автора и юзера в запросе
			if value.AuthorId == storage.UserId(Id) {
				finalResponse = append(finalResponse, value)
			}
		}
		h.StorageMu.RUnlock()

		sort.Slice(finalResponse, func(i, j int) bool {
			layout := "2006-01-02T15:04:05.000Z"
			first, _ := time.Parse(layout, string(finalResponse[i].CreatedAt))
			second, _ := time.Parse(layout, string(finalResponse[j].CreatedAt))
			return first.After(second)
		})
	} else {

		finalResponse, err = h.Storage.GetPostsByUser(r.Context(), storage.UserId(Id))
		if err != nil {
			http.Error(rw, "YOU SUCK AT DB LOSER", http.StatusBadRequest)
			return
		}
	}

	startPage := 0
	for i, value := range finalResponse {
		//valueId, _ := primitive.ObjectIDFromHex(pagetoken)
		if pagetoken != "" && value.Id == storage.PostId(pagetoken) {
			startPage = i
		}
	}
	if pagetoken != "" && startPage == 0 {
		http.Error(rw, "InvalidPageToken", http.StatusBadRequest)
		return
	}
	finalResponse = finalResponse[startPage:]
	returnResponseOld, _ := json.Marshal("")
	if len(finalResponse) >= sizepage+1 {
		rawResponse.Posts = finalResponse[0:sizepage]
		rawResponse.NextPage = finalResponse[sizepage].Id
		returnResponseOld, _ = json.Marshal(rawResponse)
	} else {
		returnResponseOld, _ = json.Marshal(PutAllPostsResponseNoNext{Posts: finalResponse})
	}
	if len(finalResponse) == 0 {
		returnResponseOld, _ = json.Marshal(PutAllPostsResponseNoNext{Posts: make([]storage.Post, 0)})
	}
	_, err = rw.Write(returnResponseOld)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}
