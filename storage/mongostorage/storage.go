package mongostorage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"os"
	"time"
	"twitterLikeHW/generator"
	storage2 "twitterLikeHW/storage"
)

var dbName = os.Getenv("MONGO_DBNAME")

const collName = "posts"

type storage struct {
	posts *mongo.Collection
}

func NewStorage(mongoURL string) *storage {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		panic(err)
	}

	collection := client.Database(dbName).Collection(collName)

	ensureIndexes(ctx, collection)

	return &storage{
		posts: collection,
	}
}

func ensureIndexes(ctx context.Context, collection *mongo.Collection) {
	indexModels := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "_id", Value: bsonx.Int32(1)}},
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	_, err := collection.Indexes().CreateMany(ctx, indexModels, opts)
	if err != nil {
		panic(fmt.Errorf("failed to ensure indexes %w", err))
	}
}

func (s *storage) PutPost(ctx context.Context, post string, userId string) (storage2.Post, error) {
	for attempt := 0; attempt < 5; attempt++ {
		newId, _ := generator.GenerateBase64ID(6)
		postId := storage2.PostId{Postid: newId}
		item := postItem{
			Key: postId.Postid,
			Post: storage2.Post{
				Id:        newId,
				Text:      post,
				AuthorId:  userId,
				CreatedAt: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			},
		}

		_, err := s.posts.InsertOne(ctx, item)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue
			}
			return storage2.Post{}, fmt.Errorf("something went wrong - %w", storage2.StorageError)
		}
		return item.Post, nil
	}
	return storage2.Post{}, fmt.Errorf("too much attempts - %w", storage2.ErrCollision)
}

func (s *storage) GetPostById(ctx context.Context, id storage2.PostId) (storage2.Post, error) {
	var result postItem
	err := s.posts.FindOne(ctx, bson.M{"_id": id.Postid}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return storage2.Post{}, fmt.Errorf("no documents with key %v - %w", id, storage2.ErrNotFound)
		}
		return storage2.Post{}, fmt.Errorf("something went wrong - %w", storage2.StorageError)
	}
	return result.Post, nil
}

func (s *storage) GetPostsByUser(ctx context.Context, id storage2.UserId) ([]storage2.Post, error) {
	var result []storage2.Post
	cursor, err := s.posts.Find(ctx, bson.M{"post.authorid": id.Userid})

	if err != nil {
		return []storage2.Post{}, err
	}
	/*
		if err = cursor.All(ctx, &response); err != nil {
			return []storage2.Post{}, err
		}*/

	for cursor.Next(ctx) {
		var elem postItem
		err = cursor.Decode(&elem)
		result = append(result, elem.Post)
	}

	return result, nil
}

type postItem struct {
	Key  string        `bson:"_id"`
	Post storage2.Post `bson:"post"`
}
