package mongostorage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
	"os"
	"time"
	storage2 "twitterLikeHW/storage"
)

var dbName = os.Getenv("MONGO_DBNAME")

//const dbName = "twitterPosts"
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

func (s *storage) PutPost(ctx context.Context, post storage2.Text, userId storage2.UserId) (storage2.Post, error) {
	for attempt := 0; attempt < 5; attempt++ {
		currentTime := storage2.ISOTimestamp(time.Now().UTC().Format(time.RFC3339))
		item := storage2.Post{
			Id:             primitive.NewObjectID(),
			Text:           post,
			AuthorId:       userId,
			CreatedAt:      currentTime,
			LastModifiedAt: currentTime,
		}

		_, err := s.posts.InsertOne(ctx, item)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue
			}
			return storage2.Post{}, fmt.Errorf("something went wrong - %w", storage2.StorageError)
		}
		return item, nil
	}
	return storage2.Post{}, fmt.Errorf("too much attempts - %w", storage2.ErrCollision)
}

func (s *storage) GetPostById(ctx context.Context, id storage2.PostId) (storage2.Post, error) {
	var result storage2.Post
	valueId, _ := primitive.ObjectIDFromHex(string(id))
	err := s.posts.FindOne(ctx, bson.D{{"_id", valueId}}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("no documents with key")
			return storage2.Post{}, fmt.Errorf("no documents with key %v - %w", id, storage2.ErrNotFound)
		}
		log.Printf("something went wrong")
		return storage2.Post{}, fmt.Errorf("something went wrong - %w", storage2.StorageError)
	}
	return result, nil
}

func (s *storage) PatchPostById(ctx context.Context, id storage2.PostId, post storage2.Text, userId storage2.UserId) (storage2.Post, error) {
	valueId, _ := primitive.ObjectIDFromHex(string(id))
	currentTime := storage2.ISOTimestamp(time.Now().UTC().Format(time.RFC3339))
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	filter := bson.D{{"_id", valueId}}
	update := bson.D{
		{"$set", bson.D{{"text", post}}},
		{"$set", bson.D{{"lastModifiedAt", currentTime}}},
	}

	newPost1 := s.posts.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	)

	result := storage2.Post{}
	err := newPost1.Decode(&result)
	if err != nil {
		return storage2.Post{}, fmt.Errorf("%w", storage2.StorageError)
	}
	return result, nil
}

func (s *storage) GetPostsByUser(ctx context.Context, id storage2.UserId) ([]storage2.Post, error) {
	var result []storage2.Post
	opt := options.Find().SetSort(bson.D{{"_id", -1}})
	cursor, err := s.posts.Find(ctx, bson.M{"authorId": id}, opt)
	if err != nil {
		return []storage2.Post{}, err
	}
	for cursor.Next(ctx) {
		var elem storage2.Post
		err = cursor.Decode(&elem)
		result = append(result, elem)
	}
	return result, nil
}
