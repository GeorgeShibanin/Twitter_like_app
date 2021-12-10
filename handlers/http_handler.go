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
	_ "sync"
	"time"
	"twitterLikeHW/generator"
	_ "twitterLikeHW/generator"
	//"twitterLikeHW/generator"
	"twitterLikeHW/storage"
)

var authorIdPattern = regexp.MustCompile(`[0-9a-f]+`)

type HTTPHandler struct {
	StorageOld map[storage.PostId]storage.PostOld
	Storage    storage.Storage
}

type PutAllPostsResponseData struct {
	Posts    []storage.Post     `json:"posts"`
	NextPage primitive.ObjectID `json:"nextPage"`
}
type PutAllPostsResponseDataOld struct {
	Posts    []storage.PostOld `json:"posts"`
	NextPage storage.PostId    `json:"nextPage"`
}
type PutAllPostsResponseNoNext struct {
	Posts []storage.Post `json:"posts"`
}
type PutAllPostsResponseNoNextOld struct {
	Posts []storage.PostOld `json:"posts"`
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

	var newPostOld storage.PostOld
	var newPost storage.Post
	var rawResponse []byte

	if storageType == "inmemory" {
		newId, _ := generator.GenerateBase64ID(6)
		currentTime := storage.ISOTimestamp(time.Now().UTC().Format(time.RFC3339))
		//time.Now().UTC().
		newPostOld = storage.PostOld{
			Id:             storage.PostId(newId),
			Text:           post.Text,
			AuthorId:       storage.UserId(tokenHeader),
			CreatedAt:      currentTime,
			LastModifiedAt: currentTime,
		}
		h.StorageOld[newPostOld.Id] = newPostOld
		rawResponse, _ = json.Marshal(newPostOld)
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
	postId := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	Id := storage.PostId(postId)

	storageType := os.Getenv("STORAGE_MODE")
	//storageType := "inmemory"

	var postTextOld storage.PostOld
	var postText storage.Post
	var err error
	var found bool
	var rawResponse []byte

	if storageType == "inmemory" {
		postTextOld, found = h.StorageOld[Id]
		if !found {
			http.NotFound(rw, r)
			return
		}
		rawResponse, _ = json.Marshal(postTextOld)
	} else {
		postText, err = h.Storage.GetPostById(r.Context(), Id)
		if err != nil {
			http.NotFound(rw, r)
			return
		}
		rawResponse, _ = json.Marshal(postText)
	}
	_, err = rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, "Поста с указанным идентификатором не существует", http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandlePatchPosts(rw http.ResponseWriter, r *http.Request) {
	postId := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
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
	var updatePostOld storage.PostOld
	var updatePost storage.Post
	var rawResponse []byte

	if storageType == "inmemory" {
		updatePostOld, found = h.StorageOld[Id]
		if !found {
			http.NotFound(rw, r)
			return
		}
		if updatePostOld.AuthorId != storage.UserId(tokenHeader) {
			http.Error(rw, "wrong user for this post", http.StatusForbidden)
			return
		}
		currentTime := storage.ISOTimestamp(time.Now().UTC().Format(time.RFC3339))
		updatePostOld.LastModifiedAt = currentTime
		updatePostOld.Text = updatePostText.Text
		newPost := storage.PostOld{
			Id:             updatePostOld.Id,
			Text:           updatePostText.Text,
			AuthorId:       updatePostOld.AuthorId,
			CreatedAt:      updatePostOld.CreatedAt,
			LastModifiedAt: currentTime,
		}
		h.StorageOld[Id] = newPost
		rawResponse, _ = json.Marshal(newPost)
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
		rawResponse, _ = json.Marshal(updatePost)
	}

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

	userId := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	Id := strings.TrimSuffix(userId, "/posts")

	storageType := os.Getenv("STORAGE_MODE")
	//storageType := "inmemory"

	var finalResponse []storage.Post
	var finalResponseOld []storage.PostOld
	rawResponse := PutAllPostsResponseData{}
	rawResponseOld := PutAllPostsResponseDataOld{}
	var err error

	if storageType == "inmemory" {
		for _, value := range h.StorageOld { //итерируемся по мапу постов и выводим пост если совпал айдишник автора и юзера в запросе
			if value.AuthorId == storage.UserId(Id) {
				finalResponseOld = append(finalResponseOld, value)
			}
		}
		sort.Slice(finalResponseOld, func(i, j int) bool {
			//layout := "2006-01-02T15:04:05.000Z"
			first, _ := time.Parse(time.RFC3339, string(finalResponseOld[i].CreatedAt))
			second, _ := time.Parse(time.RFC3339, string(finalResponseOld[j].CreatedAt))
			return first.After(second)
		})

		startPage := 0
		for i, value := range finalResponseOld {
			if pagetoken != "" && value.Id == storage.PostId(pagetoken) {
				startPage = i
			}
		}
		if pagetoken != "" && startPage == 0 {
			http.Error(rw, "InvalidPageToken", http.StatusBadRequest)
			return
		}
		finalResponseOld = finalResponseOld[startPage:]
		returnResponseOld, _ := json.Marshal("")
		if len(finalResponseOld) >= sizepage+1 {
			rawResponseOld.Posts = finalResponseOld[0:sizepage]
			rawResponseOld.NextPage = finalResponseOld[sizepage].Id
			returnResponseOld, _ = json.Marshal(rawResponseOld)
		} else {
			returnResponseOld, _ = json.Marshal(PutAllPostsResponseNoNextOld{Posts: finalResponseOld})
		}
		if len(finalResponseOld) == 0 {
			returnResponseOld, _ = json.Marshal(PutAllPostsResponseNoNextOld{Posts: make([]storage.PostOld, 0)})
		}
		_, err = rw.Write(returnResponseOld)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
	} else {

		finalResponse, err = h.Storage.GetPostsByUser(r.Context(), storage.UserId(Id))
		if err != nil {
			http.Error(rw, "YOU SUCK AT DB LOSER", http.StatusBadRequest)
			return
		}

		startPage := 0
		for i, value := range finalResponse {
			valueId, _ := primitive.ObjectIDFromHex(pagetoken)
			if pagetoken != "" && value.Id == valueId {
				startPage = i
			}
		}
		if pagetoken != "" && startPage == 0 {
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
}
