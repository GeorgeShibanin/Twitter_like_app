package handlers

import (
	"Design_System/twitterLikeHW/generator"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PostId struct {
	Postid string
}

type UserId struct {
	Userid string
}

type ISOTimestamp struct {
	Time time.Time
}

type Post struct {
	Id        *PostId      `json:"id"`
	Text      string       `json:"text"`
	AuthorId  *UserId      `json:"authorId"`
	CreatedAt ISOTimestamp `json:"createdAt"`
}

type PageToken struct {
	token string
}

type HTTPHandler struct {
	StorageMu sync.RWMutex
	Storage   map[PostId]Post
}
type PutResponseData struct {
	Key Post `json:"Post" `
}

type PutRequestData struct {
	Text string `json:"text"`
}

func HandleRoot(rw http.ResponseWriter, r *http.Request) {
	_, err := rw.Write([]byte("Hello from server"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rw.Header().Set("Content-Type", "plain/text")
}

func (h *HTTPHandler) HandleCreatePost(rw http.ResponseWriter, r *http.Request) {
	tokenHeader := r.Header.Get("System-Design-User-Id")
	if tokenHeader == "" {
		http.Error(rw, "problem with token", http.StatusUnauthorized)
	}

	rw.Header().Set("Content-Type", "application/json")
	var post PutRequestData
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	newId := generator.GetRandomKey()
	Id := PostId{Postid: newId}
	newPost := Post{
		Id:        &Id,
		Text:      post.Text,
		AuthorId:  &UserId{Userid: tokenHeader},
		CreatedAt: ISOTimestamp{time.Now()},
	}

	h.StorageMu.Lock()
	h.Storage[Id] = newPost
	h.StorageMu.Unlock()

	response := PutResponseData{
		Key: newPost,
	}
	rawResponse, _ := json.Marshal(response.Key)

	_, err = rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleGetPosts(rw http.ResponseWriter, r *http.Request) {
	postId := strings.Trim(r.URL.Path, "/api/v1/posts/")
	Id := PostId{Postid: postId}
	h.StorageMu.RLock()
	postText, found := h.Storage[Id]
	h.StorageMu.RUnlock()
	if !found {
		http.NotFound(rw, r)
		return
	}
	response := PutResponseData{
		Key: postText,
	}
	rawResponse, _ := json.Marshal(response.Key)
	_, err := rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleGetUserPosts(rw http.ResponseWriter, r *http.Request) {
	userId := strings.TrimSuffix(r.URL.Path, "/api/v1/users/")
	userId = strings.TrimSuffix(userId, "/posts")
	Id := UserId{Userid: userId}
	h.StorageMu.RUnlock()
	countPosts := 0
	for _, value := range h.Storage { //итерируемся по мапу постов и выводим пост если совпал айдишник автора и юзера в запросе
		if value.AuthorId == &Id && countPosts < 10 {
			countPosts += 1

			response := PutResponseData{
				Key: value,
			}
			rawResponse, _ := json.Marshal(response.Key)
			_, err := rw.Write(rawResponse)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}
	h.StorageMu.RUnlock()
}
