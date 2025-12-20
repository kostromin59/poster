package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
)

func TestRedisSetWithExpiration(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := Redis{
		client: db,
	}

	t.Run("successful", func(t *testing.T) {
		key := "some_key"
		data := []byte("some_data")
		exp := 1 * time.Minute

		mock.ExpectSet(key, any(data), exp).RedisNil()

		err := cache.SetWithExpiration(t.Context(), key, data, exp)
		if err != nil {
			t.Errorf("unexpected error: %q", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		key := "some_key"
		data := []byte("some_data")
		exp := 1 * time.Minute

		mock.ExpectSet(key, any(data), exp).SetErr(errors.New("some err"))

		err := cache.SetWithExpiration(t.Context(), key, data, exp)
		if err == nil {
			t.Error("unexpected error but got nil")
		}
	})
}

func TestRedisGet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := Redis{
		client: db,
	}

	t.Run("successful", func(t *testing.T) {
		key := "some_key"
		expectedData := "some_data"

		mock.ExpectGet(key).SetVal(expectedData)

		data, err := cache.Get(t.Context(), key)
		if err != nil {
			t.Errorf("unexpected error: %q", err)
		}

		if string(data) != expectedData {
			t.Errorf("expected %#v but got %#v", expectedData, string(data))
		}
	})

	t.Run("not found", func(t *testing.T) {
		key := "some_key"

		mock.ExpectGet(key).RedisNil()

		data, err := cache.Get(t.Context(), key)
		if err != nil {
			t.Errorf("unexpected error: %q", err)
		}

		if len(data) != 0 {
			t.Errorf("expected data length %d but got %d", 0, len(data))
		}
	})

	t.Run("error", func(t *testing.T) {
		key := "some_key"

		mock.ExpectGet(key).SetErr(errors.New("some err"))

		_, err := cache.Get(t.Context(), key)
		if err == nil {
			t.Error("expected error but got nil")
		}
	})
}

func TestRedisDelete(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := Redis{
		client: db,
	}

	t.Run("successful", func(t *testing.T) {
		key := "some_key"

		mock.ExpectDel(key).RedisNil()

		err := cache.Delete(t.Context(), key)
		if err != nil {
			t.Errorf("unexpected error: %q", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		key := "some_key"
		expectedErr := errors.New("delete error")

		mock.ExpectDel(key).SetErr(expectedErr)

		err := cache.Delete(t.Context(), key)
		if err != expectedErr {
			t.Errorf("expected error %q but got %q", expectedErr, err)
		}
	})
}
