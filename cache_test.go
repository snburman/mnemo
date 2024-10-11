package mnemo

import (
	"reflect"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	cache := newCache[int]()
	cacheType := reflect.TypeOf(cache)
	expected := "*mnemo.cache[int]"
	if cacheType.String() != expected {
		t.Errorf("invalid return type %v; expected %v", cacheType, expected)
	}
}

func TestGet(t *testing.T) {
	cache := newCache[int]()
	expected := 1
	key := "one"
	err := cache.Cache(key, &expected)
	if err != nil {
		t.Error(err)
	}
	data, ok := cache.Get(key)
	if !ok {
		t.Errorf("could not get cache with key %v; expected value %v", key, expected)
	}
	if data.Data != &expected {
		t.Errorf("got wrong cache value %v, expected %v", data.Data, &expected)
	}

	noData, ok := cache.Get("invalid_key")
	if ok {
		t.Errorf("expected ok to be false with invalid key")
	}
	if noData.Data != nil {
		t.Errorf("expected data result with invalid key to be nil")
	}
}

func TestAll(t *testing.T) {
	cache := newCache[int]()
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	for k := range nums {
		err := cache.Cache(k, &nums[k])
		if err != nil {
			t.Error(err)
		}
	}
	data := cache.GetAll()
	if len(data) != len(nums) {
		t.Error("returned incorrect number of items")
	}

	for k, num := range nums {
		if num != *data[k].Data {
			t.Error("returned incorrect values")
		}
	}
}

func TestCache(t *testing.T) {
	cache := newCache[int]()
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	for k := range nums {
		err := cache.Cache(k, &nums[k])
		if err != nil {
			t.Error(err)
		}
	}
	for k := range nums {
		err := cache.Cache(k, &nums[k])
		if err == nil {
			t.Error("expected duplicate key error")
		}
	}
}

func TestCacheWithTimeout(t *testing.T) {
	cache := newCache[int]()
	data := 1
	key := "one"
	ch := make(chan int)
	timeoutFunc := func(data *int) {
		ch <- *data
	}
	cfg := NewCacheTimeoutConfig[int](&data, key, timeoutFunc, time.Microsecond*1)
	err := cache.CacheWithTimeout(cfg)
	if err != nil {
		t.Error(err)
	}

	result := <-ch
	if result != data {
		t.Error("incorrect result return from timeoutFunc")
	}

	cfg.timeout = 0
	err = cache.CacheWithTimeout(cfg)
	if err == nil {
		t.Error("expected timeout length must be greater than 0 error")
	}
}

func TestDelete(t *testing.T) {
	cache := newCache[int]()
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	for k := range nums {
		err := cache.Cache(k, &nums[k])
		if err != nil {
			t.Error(err)
		}
	}
	err := cache.Delete(0)
	if err != nil {
		t.Error(err)
	}
	if len(cache.GetAll()) != 7 {
		t.Error("expected one item to be deleted from cache")
	}
	err = cache.Delete(0)
	if err == nil {
		t.Error("expected invalid key error after item already deleted")
	}
}
