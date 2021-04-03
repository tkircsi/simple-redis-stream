package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Provider struct {
	rc     *redis.Client
	ctx    context.Context
	Stream string
	Name   string
}

func main() {
	idx := time.Now().UnixNano()
	host := flag.String("host", "localhost:6379", "redis host")
	pwd := flag.String("password", "", "Redis password")
	name := flag.String("name", fmt.Sprintf("Provider-%d", idx), "Provider name")
	stream := flag.String("stream", "", "Redis Stream name")
	stime := flag.Int("time", 2000, "Message sending frequency in milliseconds")
	flag.Parse()

	provider := NewProvider(
		context.Background(),
		*host,
		*pwd,
		*name,
		*stream)

	err := provider.healthCheck()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to Redis...")

	for {
		go provider.SendMessage([]string{"message1", "message2"})
		time.Sleep(time.Duration(*stime) * time.Millisecond)
	}
}

func NewProvider(ctx context.Context, host, password, name, stream string) *Provider {
	rc := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password, // "mypassword"
		DB:       0,
	})

	return &Provider{
		rc:     rc,
		ctx:    ctx,
		Stream: stream,
		Name:   name,
	}
}

func (p *Provider) healthCheck() error {
	res, err := p.rc.Ping(p.ctx).Result()
	if err != nil || res != "PONG" {
		return fmt.Errorf("can not connect to Redis server")
	}
	return nil
}

func (p *Provider) SendMessage(msg []string) error {
	id, err := p.rc.XAdd(p.ctx, &redis.XAddArgs{
		Stream: p.Stream,
		Values: msg,
	}).Result()
	if err != nil {
		return err
	}
	fmt.Printf("Provider: %s, Stream: %s, Message: ID: %s, Values: %v\n", p.Name, p.Stream, id, msg)
	return nil
}
