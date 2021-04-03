# Redis Stream Consumer Group example

This is a simple example of Redis Stream Consumer Group

## Provider

`go run provider/provider.go -password mypassword -stream mystream2 -name P-2 -time 1000`

Starts a provider on `localhost:6379` with name `P-2` writing messages into `mystream2` stream with 1000 ms time interval.

## Consumer

`go run consumer/consumer.go -password mypassword -group Group-1 -name C-1 -time 200 mystream2 ">"`

Starts a consumer on `localhost:6379` and creates a group `Group-1`, if it is not exists. Consumer listen the stream `mystream2` and it waits 200 ms before acknowledge(XACK) the message. The consumer can be set up to listen more streams with adding the streams and ID.

`... mystream2 newstream ">" ">"`

## 1 provider - N consumer setup

```
go run provider/provider.go -password mypassword -stream mystream2 -name P-2 -time 1000

go run consumer/consumer.go -password mypassword -group Group-1 -name C-1 -time 200 mystream2 ">"

go run consumer/consumer.go -password mypassword -group Group-1 -name C-2 -time 200 mystream2 ">"

go run consumer/consumer.go -password mypassword -group Group-1 -name C-3 -time 200 mystream2 ">"
```
