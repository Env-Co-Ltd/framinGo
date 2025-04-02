package cache

import (
	"testing"
)

func TestBadgerCache_Has(t *testing.T) {
  err := testBadgerCache.Forget("foo")
  if err != nil {
    t.Error(err)
  }
  inCache, err := testBadgerCache.Has("foo")
  if err != nil {
    t.Error(err)
  }
  if inCache {
    t.Error("foo should not be in cache")
  }
  _ = testBadgerCache.Set("foo", "bar")
  inCache, err = testBadgerCache.Has("foo")
  if err != nil {
    t.Error(err)
  }
  if !inCache {
    t.Error("foo should be in cache")
  }
}

func TestBadgerCache_EmptyByMatch(t *testing.T) {
  err := testBadgerCache.Set("alpha", "beta")
  if err != nil {
    t.Error(err)
  }
  err = testBadgerCache.Set("alpha:2", "beta2")
  if err != nil {
    t.Error(err)
  }
  err = testBadgerCache.Set("beta", "beta")
  if err != nil {
    t.Error(err)
  }
  err = testBadgerCache.EmptyByMatch("alpha")
  if err != nil {
    t.Error(err)
  }
  inCache, err := testBadgerCache.Has("alpha")
  if err != nil {
    t.Error(err)
  }
  if inCache {
    t.Error("alpha should be empty")
  }

  inCache, err = testBadgerCache.Has("alpha2")
  if err != nil {
    t.Error(err)
  }
  if inCache {
    t.Error("alpha should be empty")
  }

  inCache, err = testBadgerCache.Has("beta")
  if err != nil {
    t.Error(err)
  }
  if !inCache {
    t.Error("beta should be in cache")
  }
  
}