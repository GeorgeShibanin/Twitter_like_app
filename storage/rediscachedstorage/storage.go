package rediscachedstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
	"twitterLikeHW/storage"
)

func NewStorage(redisUrl string, persistentStorage storage.Storage) *Storage {
	return &Storage{
		persistentStorage: persistentStorage,
		client:            redis.NewClient(&redis.Options{Addr: redisUrl}),
	}
}

type Storage struct {
	persistentStorage storage.Storage
	client            *redis.Client
}

var _ storage.Storage = (*Storage)(nil)

func (s *Storage) PutPost(ctx context.Context, post storage.Text, userId storage.UserId) (storage.Post, error) {
	postPut, err := s.persistentStorage.PutPost(ctx, post, userId)
	if err != nil {
		return postPut, err
	}

	fullKey := s.getFullKey(storage.PostId(postPut.Id.Hex()))
	bytes, _ := json.Marshal(postPut)
	response := s.client.HSetNX(ctx, fullKey, "test", bytes)

	if err := response.Err(); err != nil {
		//log.Printf("Failed to save key %s to redis", fullKey)
		return postPut, err
	}
	s.client.Expire(ctx, fullKey, time.Hour)
	return postPut, nil
}

func (s *Storage) GetPostById(ctx context.Context, id storage.PostId) (storage.Post, error) {

	//get, err := s.client.HGet(ctx, s.getFullKey(id), "test").Result()
	switch rawPost, err := s.client.HGet(ctx, s.getFullKey(id), "test").Result(); {
	case err == redis.Nil:
	//go to persistence
	case err != nil:
		return storage.Post{}, fmt.Errorf("%w: failed to get value from redis due to error %s", storage.StorageError, err)
	default:
		result := storage.Post{}
		err2 := json.Unmarshal([]byte(rawPost), &result)
		if err2 != nil {
			return storage.Post{}, nil
		}
		return result, nil
	}
	log.Printf("Loading url by key %s from persistent storage", id)
	postGet, err := s.persistentStorage.GetPostById(ctx, id)
	if err != nil {
		return postGet, err
	}
	return postGet, nil
}

func (s *Storage) PatchPostById(ctx context.Context, id storage.PostId, post storage.Text, userId storage.UserId) (storage.Post, error) {
	result, err := s.persistentStorage.PatchPostById(ctx, id, post, userId)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *Storage) GetPostsByUser(ctx context.Context, id storage.UserId) ([]storage.Post, error) {
	result, err := s.persistentStorage.GetPostsByUser(ctx, id)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *Storage) getFullKey(key storage.PostId) string {
	return "app:post:" + string(key)
}
