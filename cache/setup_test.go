package cache

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dgraph-io/badger/v3"
	"github.com/gomodule/redigo/redis"
)

var testRedisCache RedisCache
var testBadgerCache BadgerCache

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	pool := redis.Pool{
		MaxIdle:     50,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.Addr())
		},
	}
	testRedisCache.Conn = &pool
	testRedisCache.Prefix = "test-framingo"

	defer testRedisCache.Conn.Close()

	_ = os.RemoveAll("./testdata/tmp/badger")

	//create badger db
	if _, err := os.Stat("./testdata/tmp"); os.IsNotExist(err) {
		err := os.Mkdir("./testdata/tmp", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = os.Mkdir("./testdata/tmp/badger", 0755)
	if err != nil {
		log.Fatal(err)
	}

	db, err := badger.Open(badger.DefaultOptions("./testdata/tmp/badger"))
	testBadgerCache.Conn = db

	os.Exit(m.Run())
}


func TestBadgerCache_Forget(t *testing.T) {
	err := testBadgerCache.Set("foo", "foo")
	if err != nil {
		t.Fatal(err)
	}
	err = testBadgerCache.Forget("foo")
	if err != nil {
		t.Fatal(err)
	}
	inCache, err := testBadgerCache.Has("foo")
	if err != nil {
		t.Fatal(err)
	}
	if inCache {
		t.Fatal("foo should be forgotten")
	}
}

func TestBadgerCache_Empty(t *testing.T) {
	err := testBadgerCache.Set("alpha", "beta")
	if err != nil {
		t.Fatal(err)
	}
	err = testBadgerCache.Empty()
	if err != nil {
		t.Fatal(err)
	}
	inCache, err := testBadgerCache.Has("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if inCache {
		t.Fatal("alpha should be empty")
	}
}