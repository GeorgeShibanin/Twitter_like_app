package storage

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type ISOTimestamp struct {
	Time time.Time
}

type PageToken struct {
	Token string
}

var (
	StorageError = errors.New("storage")
	ErrCollision = fmt.Errorf("%w.collision", StorageError)
	ErrNotFound  = fmt.Errorf("%w.not_found", StorageError)
)

type Text string
type UserId string
type PostId string
type Time string

type Post struct {
	Id        PostId `bson:"_id"`
	Text      Text   `bson:"text"`
	AuthorId  UserId `bson:"authorid"`
	CreatedAt Time   `bson:"createdat"`
}

type Storage interface {
	PutPost(ctx context.Context, post Text, userId UserId) (Post, error)
	GetPostById(ctx context.Context, id PostId) (Post, error)
	GetPostsByUser(ctx context.Context, id UserId) ([]Post, error)
}
