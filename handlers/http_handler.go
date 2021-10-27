package handlers

import (
	_ "Design_System/twitterLikeHW"
	"Design_System/twitterLikeHW/generator"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
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
	Id        *PostId 		`json:"id"`
	Text      string        `json:"text"`
	AuthorId  *UserId       `json:"authorId"`
	CreatedAt ISOTimestamp  `json:"createdAt"`
}

type PageToken struct {
	token string
}

type HTTPHandler struct {
	StorageMu sync.RWMutex
	Storage   map[*PostId]Post
}
type PutResponseData struct {
	Key Post `json:"Post" `
}

func HandleRoot(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("Hello from server"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "plain/text")
}

func (h *HTTPHandler) HandleCreatePost(rw http.ResponseWriter, r *http.Request) {
	tokenHeader := r.Header.Get("UserId")
	if tokenHeader == ""  {
		return
	}

	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	post.Id.Postid = generator.GetRandomKey()
	post.CreatedAt.Time = time.Now()
	post.AuthorId.Userid = tokenHeader

	h.StorageMu.Lock()
	h.Storage[post.Id] = post
	h.StorageMu.Unlock()

	response := PutResponseData{
		Key: post,
	}
	rawResponse, _ := json.Marshal(response)

	rw.Header().Set("Content-Type", "application/json")
	_, err = rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleGetPosts(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, post := range posts {
		if post.PostID
	}
}
