package handlers

import (
	"Design_System/twitterLikeHW/generator"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var authorIdPattern = regexp.MustCompile(`[0-9a-f]+`)

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
	Id        string `json:"id"`
	Text      string `json:"text"`
	AuthorId  string `json:"authorId"`
	CreatedAt string `json:"createdAt"`
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
type PutAllPostsResponseData struct {
	Posts    []Post `json:"posts"`
	NextPage string `json:"nextpage"`
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
	if tokenHeader == "" || !authorIdPattern.MatchString(tokenHeader) {
		http.Error(rw, "problem with token", http.StatusUnauthorized)
	}

	rw.Header().Set("Content-Type", "application/json")
	var post PutRequestData
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	newId, _ := generator.GenerateBase64ID(6)
	newPost := Post{
		Id:        newId,
		Text:      post.Text,
		AuthorId:  tokenHeader,
		CreatedAt: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}
	idPost := PostId{
		Postid: newPost.Id,
	}
	h.StorageMu.Lock()
	h.Storage[idPost] = newPost
	h.StorageMu.Unlock()

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
	Id := PostId{Postid: postId}
	h.StorageMu.RLock()
	postText, found := h.Storage[Id]
	h.StorageMu.RUnlock()
	if !found {
		http.NotFound(rw, r)
		return
	}
	rawResponse, _ := json.Marshal(postText)
	_, err := rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, "Поста с указанным идентификатором не существует", http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleGetUserPosts(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	pagetoken := PageToken{token: r.URL.Query().Get("page")}
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

	h.StorageMu.RLock()
	var finalResponse []Post
	for _, value := range h.Storage { //итерируемся по мапу постов и выводим пост если совпал айдишник автора и юзера в запросе
		if value.AuthorId == Id {
			finalResponse = append(finalResponse, value)
		}
	}
	h.StorageMu.RUnlock()

	//if len(finalResponse) == 0 {
	//	http.Error(rw, "NoUserPostsOrInvalidUserName", http.StatusBadRequest)
	//	return
	//}

	sort.Slice(finalResponse, func(i, j int) bool {
		layout := "2006-01-02T15:04:05.000Z"
		first, _ := time.Parse(layout, finalResponse[i].CreatedAt)
		second, _ := time.Parse(layout, finalResponse[j].CreatedAt)
		return first.After(second)
	})

	rawResponse := PutAllPostsResponseData{}
	startPage := 0
	for i, value := range finalResponse {
		if pagetoken.token != "" && value.Id == pagetoken.token {
			startPage = i
		}
	}
	if pagetoken.token != "" && startPage == 0 {
		http.Error(rw, "InvalidPageToken", http.StatusBadRequest)
		return
	}
	finalResponse = finalResponse[startPage:]
	if len(finalResponse) >= sizepage+1 {
		rawResponse.Posts = finalResponse[0:sizepage]
		rawResponse.NextPage = finalResponse[sizepage].Id
	} else {
		rawResponse.Posts = finalResponse
		rawResponse.NextPage = ""
	}

	returnResponse, _ := json.Marshal(rawResponse)
	_, err := rw.Write(returnResponse)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}
