package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	r "github.com/otaviokr/twitch-bot-config/redis"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Warn("reading configuration file")
	readConfig(ctx)

	tp, err := jaegerTracerProvider(viper.GetString("jaeger.uri"))
	if err != nil {
		log.Fatalf("failed to initialize a tracer provider for Jaeger: %v", err)
	}
	defer func(ctx context.Context) {
		ctx, cancel = context.WithTimeout(ctx, 5 * time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown: %v", err)
		}
	}(ctx)

	// Set Global Tracer Provider and a Global Meter Provider
	otel.SetTracerProvider(tp)
	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)

	rawLevel := viper.GetString("log.level")
	logLevel, err := log.ParseLevel(rawLevel)
	if err != nil {
		log.WithFields(
			log.Fields{
				"err": err.Error(),
				"raw_level": rawLevel,
			}).Fatal("failed to read configuration file for LOG.LEVEL")
	}
	log.SetLevel(logLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	SetupCloseHandler(ctx)

	// Infinite loop until program is terminated.
	for {
		time.Sleep(2 * time.Second)
	}
}

// SetupCloseHandler creates a listener on a new goroutine which will notify the program if it receives an interrupt from the OS.
// We then handle this by calling our clean up procedure and exiting the program.
func SetupCloseHandler(ctx context.Context) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		tracer := otel.Tracer("main")
		var span trace.Span
		_, span = tracer.Start(ctx, "SetupCloseHandler")
		defer span.End()

		log.Warn("SIGTERM received. Closing program")
		span.AddEvent("SIGTERM received... closing program")
		os.Exit(0)
	}()
}

// ReadConfig will parse the properties file.
func readConfig(ctx context.Context) {
	tracer := otel.Tracer("readConfig")
	var span trace.Span
	newCtx, span := tracer.Start(ctx, "readConfig")
	defer span.End()

	span.AddEvent("Defining config file(s)")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.SetConfigName("twitch-bot")
	viper.SetConfigType("yaml")

	span.AddEvent("Setting up live reconfig")
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		innerCtx, onChangeSpan := tracer.Start(newCtx, "onChangeConfig")
		defer onChangeSpan.End()

		sendToRedis(innerCtx)
		log.WithFields(
			log.Fields{
				"file": e.Name,
				"event": e.Op.String(),
			}).Info("configuration file changed and settings have been refreshed")
	})

	span.AddEvent("Read configuration from file(s)")
	err := viper.ReadInConfig()
	if err != nil {
		log.WithFields(
			log.Fields{
				"error": err.Error(),
			}).Fatal("failed to process config file")
	}

	sendToRedis(newCtx)
	span.AddEvent("First upload of configuration values to Redis")
}

// SendToRedis will load the contents of the configuration file into the redis instance.
func sendToRedis(ctx context.Context) {
	tracer := otel.Tracer("main")
	var span trace.Span
	newCtx, span := tracer.Start(ctx, "sendToRedis")
	defer span.End()

	port, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		fmt.Printf("Failed to convert port number: %s . Setting default %d\n", os.Getenv("REDIS_PORT"), r.RedisDefaultPort)
		port = r.RedisDefaultPort
	}

	db, err := strconv.Atoi(os.Getenv("REDIS_DATABASE"))
	if err != nil {
		fmt.Printf("Failed to convert database ID: %s . Setting default %d\n", os.Getenv("REDIS_DATABASE"), r.RedisDefaultDatabase)
		db = r.RedisDefaultPort
	}

	client := r.NewClient(ctx, os.Getenv("REDIS_URI"), port, os.Getenv("REDIS_PASSWORD"), db)
	if strings.EqualFold(client.Ping(), "PONG") {
		span.AddEvent("Connected to Redis")
		client.LoadFromFile(newCtx)
	} else {
		log.WithFields(
			log.Fields{
				"error": "ping failed",
			}).Error("connection to Redis failed")
	}
}

// jaegerTracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func jaegerTracerProvider(url string) (*sdktrace.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		// Always be sure to batch in production.
		sdktrace.WithBatcher(exp),
		// Record information about this application in an Resource.
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(viper.GetString("jaeger.service.config")),
			attribute.String("environment", viper.GetString("jaeger.environment")),
			attribute.Int64("ID", viper.GetInt64("jaeger.id")),
		)),
	)
	return tp, nil
}
