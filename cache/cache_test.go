package cache

import (
	"testing"
)

func TestCache_Has(t *testing.T) {
	err := testRedisCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo should not be in cache")
	}

	err = testRedisCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	inCache, err = testRedisCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("foo should be in cache")
	}

}

func TestCache_Get(t *testing.T) {
	err := testRedisCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	x, err := testRedisCache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if x != "bar" {
		t.Error("Did't get correct value from cache")
	}
}

func TestRedisCache_Forget(t *testing.T) {
	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Forget("alpha")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha should not be in cache")
	}
}

func TestRedisCache_Empty(t *testing.T) {
	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Empty()
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha should not be in cache")
	}
}

func TestRedisCache_EmptyByMatch(t *testing.T) {
	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}
	err = testRedisCache.Set("alpha2", "foo")
	if err != nil {
		t.Error(err)
	}
	err = testRedisCache.Set("moon", "foo")
	if err != nil {
		t.Error(err)
	}
	err = testRedisCache.EmptyByMatch("alpha")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}
	if inCache {
		t.Error("alpha should not be in cache")
	}

	inCache, err = testRedisCache.Has("alpha2")
	if err != nil {
		t.Error(err)
	}
	if inCache {
		t.Error("alpha2 should not be in cache")
	}

	inCache, err = testRedisCache.Has("moon")
	if err != nil {
		t.Error(err)
	}
	if !inCache {
		t.Error("moon should not be in cache")
	}

}

func TestEncodeDecode(t *testing.T) {
	entry := Entry{}
	entry["foo"] = "bar"
	bytes, err := encode(entry)
	if err != nil {
		t.Error(err)
	}
	_, err = decode(string(bytes))
	if err != nil {
		t.Error(err)
	}

}
