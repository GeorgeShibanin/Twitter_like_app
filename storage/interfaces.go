package storage

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	StorageError = errors.New("storage")
	ErrCollision = fmt.Errorf("%w.collision", StorageError)
	ErrNotFound  = fmt.Errorf("%w.not_found", StorageError)
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
	Id        string `json:"id"`
	Text      string `json:"text"`
	AuthorId  string `json:"authorId"`
	CreatedAt string `json:"createdAt"`
}

type PageToken struct {
	Token string
}

type Storage interface {
	PutPost(ctx context.Context, post string, userId string) (Post, error)
	GetPostById(ctx context.Context, id string) (Post, error)
	GetPostsByUser(ctx context.Context, id string) ([]Post, error)
}
