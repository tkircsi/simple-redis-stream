package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Consumer struct {
	rc          *redis.Client
	ctx         context.Context
	Group       string
	Consumer    string
	Streams     []string
	Block       time.Duration
	SkipTimeout bool
}

type ConsumerOption func(*Consumer)

func main() {
	idx := time.Now().UnixNano()
	host := flag.String("host", "localhost:6379", "redis host")
	pwd := flag.String("password", "", "Redis password")
	name := flag.String("name", fmt.Sprintf("Consumer-%d", idx), "Consumer name")
	group := flag.String("group", "", "The name of the Reader Group")
	stime := flag.Int("time", 1000, "process interval in milliseconds for testing")
	flag.Parse()
	streams := flag.Args()

	consumer := NewConsumer(
		context.Background(),
		*host,
		*pwd,
		WithConsumer(*name),
		WithGroup(*group),
		WithStreams(streams),
	)

	consumer.CreateGroup()
	consumer.Listen(*stime)
}

func NewConsumer(ctx context.Context, host string, password string, options ...ConsumerOption) *Consumer {
	rc := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password, // "mypassword"
		DB:       0,
	})

	c := &Consumer{
		rc:          rc,
		ctx:         ctx,
		Consumer:    "consumer",
		Block:       0,
		SkipTimeout: true,
	}

	for _, op := range options {
		op(c)
	}
	return c
}

func WithGroup(group string) ConsumerOption {
	return func(c *Consumer) {
		c.Group = group
	}
}

func WithConsumer(consumer string) ConsumerOption {
	return func(c *Consumer) {
		c.Consumer = consumer
	}
}

func WithStreams(streams []string) ConsumerOption {
	return func(c *Consumer) {
		c.Streams = streams
	}
}

func WithBlock(block time.Duration) ConsumerOption {
	return func(c *Consumer) {
		c.Block = block
	}
}

func WithSkipTimeout(skip bool) ConsumerOption {
	return func(c *Consumer) {
		c.SkipTimeout = skip
	}
}

func (c *Consumer) CreateGroup() {
	for i := 0; i < len(c.Streams)/2; i++ {
		_, err := c.rc.XGroupCreate(context.Background(), c.Streams[i], c.Group, "$").Result()
		if err != nil {
			log.Printf("Error creating group %q on stream %q : %v", c.Group, c.Streams[i], err)
		}
	}
}

func (c *Consumer) Listen(stime int) {
	for {
		xstreams, err := c.rc.XReadGroup(c.ctx, &redis.XReadGroupArgs{
			Group:    c.Group,
			Consumer: c.Consumer,
			Streams:  c.Streams,
			Block:    c.Block,
			Count:    1,
		}).Result()
		if c.SkipTimeout && err == redis.Nil {
			continue
		}
		if err != nil {
			log.Fatal(err)
		}

		for _, xstream := range xstreams {
			stream := xstream.Stream
			for _, msg := range xstream.Messages {
				fmt.Printf("Group: %s, Consumer: %s, Stream: %s, Message: ID: %s, Value: %v\n", c.Group, c.Consumer, stream, msg.ID, msg.Values)

				// Process a message takes stime milliseconds
				time.Sleep(time.Duration(stime) * time.Millisecond)

				ack, err := c.rc.XAck(c.ctx, stream, c.Group, msg.ID).Result()
				if ack != 1 || err != nil {
					fmt.Printf("Error XACK ID: %s\n", msg.ID)
				}
			}
		}
	}
}
