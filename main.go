package main

import (
	"log"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/oaswrap/spec/adapter/echoopenapi"
	"github.com/oaswrap/spec/option"

	"magtapp-backend/internal/config"
	"magtapp-backend/internal/db"
	"magtapp-backend/internal/dto"
	"magtapp-backend/internal/handlers"
	authmw "magtapp-backend/internal/middleware"
)

func main() {
	cfg := config.Load()

	entClient, err := db.NewEntClient(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer entClient.Close()

	redisClient := db.NewRedisClient(cfg.RedisAddr)
	defer redisClient.Close()

	e := echo.New()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	// oaswrap/spec's Echo adapter wraps Echo's own router, so every route
	// registered through `r` is both a live Echo route AND documented in
	// the generated OpenAPI spec, served at GET /docs.
	r := echoopenapi.NewRouter(e,
		option.WithTitle("MagTapp Pro API"),
		option.WithDescription("Minimal user + subscription service — founding engineer assignment"),
		option.WithVersion("1.0.0"),
		option.WithSecurity("bearerAuth", option.SecurityHTTPBearer("Bearer")),
	)

	subHandler := handlers.NewSubscriptionHandler(entClient, redisClient)

	// Every route in this group runs Auth first, which creates-or-identifies
	// the user from the Bearer token before the handler ever runs.
	v1 := r.Group("/v1", authmw.Auth(entClient)).With(
		option.GroupSecurity("bearerAuth"),
	)

	v1.GET("/me", handlers.Me).With(
		option.Summary("Identify or create the current user"),
		option.Response(200, new(dto.MeResponse)),
		option.Response(401, new(dto.ErrorResponse)),
	)

	v1.GET("/subscription", subHandler.Get).With(
		option.Summary("Get current subscription status"),
		option.Response(200, new(dto.SubscriptionResponse)),
		option.Response(401, new(dto.ErrorResponse)),
	)

	v1.PUT("/subscription", subHandler.Update).With(
		option.Summary("Update subscription status"),
		option.Description("Accepts an optional Idempotency-Key header so retried calls (e.g. from a webhook) don't double-apply."),
		option.Request(new(dto.UpdateSubscriptionRequest)),
		option.Response(200, new(dto.SubscriptionResponse)),
		option.Response(400, new(dto.ErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
	)

	log.Printf("MagTapp backend listening on :%s — docs at http://localhost:%s/docs", cfg.Port, cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
