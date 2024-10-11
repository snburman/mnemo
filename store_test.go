package mnemo

import (
	"reflect"
	"testing"
)

func TestNewStore(t *testing.T) {
	store, _ := NewStore("test")
	storeType := reflect.TypeOf(store)
	expected := "*mnemo.Store"
	actual := storeType.String()
	if actual != expected {
		t.Errorf("expected NewStore to be of type %v; got %v", expected, actual)
	}

	if store.data == nil {
		t.Errorf("expected data map to be instantiated")
	}
}

func TestCreateStoreCache(t *testing.T) {
	var key StoreKey = "test"
	NewStore(key)
	_, err := NewCache[int](key, key)
	if err != nil {
		t.Error(err)
	}
	_, err = NewCache[int](key, key)
	if err == nil {
		t.Error("expected duplicate key error")
	}
}

// TODO: This requires set up and teardown with new key retrieval mechanism
func TestUseStoreCache(t *testing.T) {
	//TODO: Need TestMain to set up base store from init in store.go
	var key StoreKey = "test"
	NewStore(key)
	cache, err := NewCache[int](key, key)
	if err != nil {
		t.Error(err)
	}

	data := 123
	cache.Cache(key, &data)
	cache, err = UseCache[int](key, key)
	if err != nil {
		t.Error(err)
	}
	if len(cache.GetAll()) != 1 {
		t.Error("cache not retrieved")
	}

	_, err = UseCache[int]("test", StoreKey("test2"))
	if err == nil {
		t.Error("expected no data with key error")
	}

	_, err = UseCache[string]("test", StoreKey("test"))
	if err == nil {
		t.Error("expected invalid type for cache with key error")
	}
}
