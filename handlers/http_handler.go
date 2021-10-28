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
	rawResponse, _ := json.Marshal(response)

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
	_, err := rw.Write([]byte(postText.Text))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleGetUserPosts(rw http.ResponseWriter, r *http.Request) {}
