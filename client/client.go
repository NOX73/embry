package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"time"
)

var (
	ErrTimeout = errors.New("Cache client time out")
)

type Client interface {
	Get(key string, result interface{}, executionLimit time.Duration, expire time.Duration, getter func()) error
}

type client struct {
	mcache *memcache.Client
}

func NewClient(server ...string) Client {
	return &client{memcache.New(server...)}
}

func (c *client) lockOrGet(key string, executionLimit time.Duration) (*memcache.Item, bool, error) {
	err := c.mcache.Add(&memcache.Item{
		Key:        fmt.Sprint(key, "_lock"),
		Value:      []byte("1"),
		Expiration: int32(executionLimit.Seconds()),
	})

	if err == nil {
		return nil, true, nil
	}

	var finish = time.Now().Add(executionLimit)

	for {
		item, err := c.mcache.Get(key)

		if err == nil {
			return item, false, nil
		}

		if time.Now().After(finish) {
			return nil, false, ErrTimeout
		}
	}

	return nil, false, ErrTimeout
}

func (c *client) Get(key string, result interface{}, executionLimit time.Duration, expire time.Duration, getter func()) error {

	item, err := c.mcache.Get(key)

	//cache hit
	if err == nil {
		json.Unmarshal(item.Value, result)
		return nil
	}

	if err != nil && err != memcache.ErrCacheMiss {
		return err
	}

	// cache miss

	item, lock, err := c.lockOrGet(key, executionLimit)

	if err != nil {
		return err
	}

	if !lock && item != nil {
		json.Unmarshal(item.Value, result)
		return nil
	}

	//got lock

	getter()

	value, err := json.Marshal(result)

	if err != nil {
		return err
	}

	c.mcache.Set(&memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(expire.Seconds()),
	})

	err = c.mcache.Delete(fmt.Sprint(key, "_lock"))

	if err != nil {
		return err
	}

	return nil
}
