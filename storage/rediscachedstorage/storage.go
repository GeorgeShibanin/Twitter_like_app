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

func NewStorage(persistentStorage storage.Storage, client *redis.Client) *Storage {
	return &Storage{
		persistentStorage: persistentStorage,
		client:            client,
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
		log.Printf("WATAFAKA")
		return postPut, err
	}

	if err = s.storePost(ctx, storage.PostId(postPut.Id.Hex()), postPut); err != nil {
		log.Printf("Failed to insert key %s into cache due to an error: %s\n", storage.PostId(postPut.Id.Hex()), err)
	}
	return postPut, nil
}

func (s *Storage) GetPostById(ctx context.Context, id storage.PostId) (storage.Post, error) {
	get := s.client.Get(ctx, s.getFullKey(id))
	switch rawPost, err := get.Result(); {
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
	log.Printf("Loading post by key %s from persistent storage", id)
	postGet, err := s.persistentStorage.GetPostById(ctx, id)
	if err != nil {
		return postGet, err
	}
	if err = s.storePost(ctx, id, postGet); err != nil {
		log.Printf("Failed to insert key %s into cache due to an error: %s\n", id, err)
	}
	return postGet, nil
}

func (s *Storage) PatchPostById(ctx context.Context, id storage.PostId, post storage.Text, userId storage.UserId) (storage.Post, error) {
	fullKey := s.getFullKey(id)
	s.client.Del(ctx, fullKey)
	result, err := s.persistentStorage.PatchPostById(ctx, id, post, userId)
	if err != nil {
		return result, err
	}
	//fullKey := s.getFullKey(storage.PostId(result.Id.Hex()))
	if err = s.storePost(ctx, id, result); err != nil {
		log.Printf("Failed to insert key %s into cache due to an error: %s\n", id, err)
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
	return "twitter:" + string(key)
}

func (s *Storage) storePost(ctx context.Context, id storage.PostId, post storage.Post) error {
	fullKey := s.getFullKey(id)
	bytes, _ := json.Marshal(post)
	return s.client.Set(ctx, fullKey, bytes, time.Minute).Err()
}
