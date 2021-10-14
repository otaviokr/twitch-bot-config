package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	RedisDefaultPort = 6379
	RedisDefaultDatabase = 0
)

var (
	varStrings = []string{
		"jaeger.uri",
		"jaeger.service.config",
		"jaeger.service.bot",
		"jaeger.environment",
		"irc.target",
		"irc.nickname",
		"irc.password",
		"mqtt.broker",
		"mqtt.port",
		"mqtt.clientId",
		"redis.uri",
		"redis.password",
		"log.level",
		"log.path",
		"triggers.guestbook.topic",
		"triggers.bot.owner",
		"triggers.bot.repository",
		"triggers.socialmedia.github",
		"triggers.socialmedia.twitter",
		"triggers.socialmedia.#youtube",
	}

	varInts = []string{
		"jaeger.id",
		"prometheus.port",
		"redis.port",
		"redis.database",
	}

	varBools = []string{
		"irc.ssl",
	}

	varSliceStrings = []string{
		"irc.channels",
		"triggers.streamholics.friends",
	}
)

// Redis is a struct that interacts with a Redis instance.
type Redis struct {
	client *redis.Client
}

// NewClient connects to the Redis instance described and returns a connected object.
// Context is used for tracing, if you are not working with Tracing (eg, Jaeger), just use the default.
func NewClient(ctx context.Context, uri string, port int, pwd string, db int) (*Redis) {
	tracer := otel.Tracer("redis")
	var span trace.Span
	_, span = tracer.Start(ctx, "NewClient",
		trace.WithAttributes(
			attribute.String("uri", uri),
			attribute.Int("port", port),
			attribute.String("pwd", pwd),
			attribute.Int("db", db),
	))
	defer span.End()

	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d", uri, port),
			Password: pwd,
			DB: db,
		}),
	}
}

// LoadFromFile will read the variables from the configuration file and load them into the Redis instance.
// Context is used for tracing, if you are not working with Tracing (eg, Jaeger), just use the default.
func (c *Redis) LoadFromFile(ctx context.Context) {
	log.Info("Loading variables from file")
	tracer := otel.Tracer("redis")
	var span trace.Span
	_, span = tracer.Start(ctx, "LoadValues")
	defer span.End()

	c.saveStrings()
	span.AddEvent("Strings loaded from configuration file")

	c.saveInts()
	span.AddEvent("Integer loaded from configuration file")

	c.saveBools()
	span.AddEvent("Booleans loaded from configuration file")

	c.saveSliceStrings()
	span.AddEvent("Slice of Strings loaded from configuration file")
}

// Ping will check if communication with Redis instance is active.
//
// Expected response is "PONG".
func (c *Redis) Ping() (string) {
	return c.client.Ping().Val()
}

// GetString will fetch the given key in Redis.
func (c *Redis) GetString(k string) (string) {
	return c.client.Get(k).Val()
}

// GetInt will fetch the given key in Redis.
func (c *Redis) GetInt(k string) (int) {
	i, _ := c.client.Get(k).Int()
	return i
}

// GetBool will fetch the given key in Redis.
func (c *Redis) GetBool(k string) (bool) {
	i, _ := c.client.Get(k).Int()
	return i == 1
}

// GetSliceString will fetch the given key in Redis.
func (c *Redis) GetSliceString(k string) ([]string) {
	return c.client.SMembers(k).Val()
}

func (c *Redis) saveStrings() {
	for _, key := range varStrings {
		c.client.Set(key, viper.GetString(key), 0)
	}
}

func (c *Redis) saveInts() {
	for _, key := range varInts {
		c.client.Set(key, viper.GetInt(key), 0)
	}
}

func (c *Redis) saveBools() {
	for _, key := range varBools {
		c.client.Set(key, viper.GetBool(key), 0)
	}
}

func (c *Redis) saveSliceStrings() {
	for _, key := range varSliceStrings {
		c.client.SAdd(key, viper.GetStringSlice(key))
	}
}