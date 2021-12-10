package storage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	StorageError = errors.New("storage")
	ErrCollision = fmt.Errorf("%w.collision", StorageError)
	ErrNotFound  = fmt.Errorf("%w.not_found", StorageError)
)

type Text string
type UserId string
type PostId string
type ISOTimestamp string
type PageToken string

type Post struct {
	Id             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Text           Text               `json:"text" bson:"text"`
	AuthorId       UserId             `json:"authorId" bson:"authorId"`
	CreatedAt      ISOTimestamp       `json:"createdAt" bson:"createdAt"`
	LastModifiedAt ISOTimestamp       `json:"lastModifiedAt"  bson:"lastModifiedAt"`
}

//type PostOld struct {
//	Id             PostId       `json:"id"`
//	Text           Text         `json:"text"`
//	AuthorId       UserId       `json:"authorId"`
//	CreatedAt      ISOTimestamp `json:"createdAt"`
//	LastModifiedAt ISOTimestamp `json:"lastModifiedAt"`
//}

type Storage interface {
	PutPost(ctx context.Context, post Text, userId UserId) (Post, error)
	GetPostById(ctx context.Context, id PostId) (Post, error)
	PatchPostById(ctx context.Context, id PostId, post Text, userId UserId) (Post, error)
	GetPostsByUser(ctx context.Context, id UserId) ([]Post, error)
}
