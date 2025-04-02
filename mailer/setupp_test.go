package mailer

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var pool *dockertest.Pool
var resource *dockertest.Resource

var mailer = Mail{
	Domain:      "localhost",
	Templates:   "./testdata/mail",
	FromAddress: "test@example.com",
	FromName:    "Test",
	Port:        1026,
	Jobs:        make(chan Message, 1),
	Results:     make(chan Result, 1),
}

func TestMain(m *testing.M) {
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	pool = p
	opts := &dockertest.RunOptions{
		Repository:   "mailhog/mailhog",
		Tag:          "latest",
		Env:          []string{},
		ExposedPorts: []string{"1025", "8025"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"1025": {
				{HostIP: "0.0.0.0", HostPort: "1026"},
			},
			"8025": {
				{HostIP: "0.0.0.0", HostPort: "8025"},
			},
		},
	}
	resource, err = pool.RunWithOptions(opts)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
		_ = pool.Purge(resource)
		// os.Exit(1)
	}

	time.Sleep(3 * time.Second)
	go mailer.ListenForMail()
	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Printf("Warning: Could not purge resource: %s", err)
	}
	os.Exit(code)
}
