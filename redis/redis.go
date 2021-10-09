package redis

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
)

const (
	RedisDefaultPort = 6379
	RedisDefaultDatabase = 0
)

var (
	varStrings = []string{
		"jaeger.uri",
		"jaeger.service",
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

type Redis struct {
	client *redis.Client
}

func NewClient(uri string, port int, pwd string, db int) (*Redis) {
	return &Redis{
		client: redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", viper.GetString("redis.uri"), viper.GetInt("redis.port")),
		Password: viper.GetString("redis.password"),
		DB: viper.GetInt("redis.database"),
	})}
}

func (c *Redis) SaveToFile(ctx context.Context) {
	tracer := otel.Tracer("redis")
	var span trace.Span
	_, span = tracer.Start(ctx, "SaveValues")
	defer span.End()

	span.AddEvent("Client defined")

	c.saveStrings()
	span.AddEvent("Strings saved to configuration file")

	c.saveInts()
	span.AddEvent("Integer saved to configuration file")

	c.saveBools()
	span.AddEvent("Booleans saved to configuration file")

	c.saveSliceStrings()
	span.AddEvent("Slice of Strings saved to configuration file")
}

func (c *Redis) LoadFromFile(ctx context.Context) {
	tracer := otel.Tracer("redis")
	var span trace.Span
	_, span = tracer.Start(ctx, "LoadValues")
	defer span.End()

	span.AddEvent("Client defined")

	c.saveStrings()
	span.AddEvent("Strings loaded from configuration file")

	c.saveInts()
	span.AddEvent("Integer loaded from configuration file")

	c.saveBools()
	span.AddEvent("Booleans loaded from configuration file")

	c.saveSliceStrings()
	span.AddEvent("Slice of Strings loaded from configuration file")
}

func (c *Redis) LoadToViper(ctx context.Context) {
	for _, key := range varStrings {
		viper.Set(key, c.GetString(key))
	}

	for _, key := range varInts {
		viper.Set(key, c.GetInt(key))
	}

	for _, key := range varBools {
		viper.Set(key, c.GetBool(key))
	}

	for _, key := range varSliceStrings {
		viper.Set(key, c.GetSliceString(key))
	}
}

func (c *Redis) GetAllKeys() ([]string) {
	return viper.AllKeys()
}

func (c *Redis) GetString(k string) (string) {
	return c.client.Get(k).String()
}

func (c *Redis) GetInt(k string) (int) {
	i, _ := c.client.Get(k).Int()
	return i
}

func (c *Redis) GetBool(k string) (bool) {
	i, _ := c.client.Get(k).Int()
	return i == 1
}

func (c *Redis) GetSliceString(k string) ([]string) {
	return []string{c.client.Get(k).String()}
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
		c.client.Set(key, viper.GetStringSlice(key), 0)
	}
}