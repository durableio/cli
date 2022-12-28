package server

import (
	"runtime"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/gofiber/fiber/v2"

	"github.com/durableio/cli/pkg/durable"
	"github.com/durableio/cli/pkg/logging"
)

type Config struct {
	Logger  logging.Logger
	Durable durable.Durable
}

type RequestId string

type Server struct {
	logger logging.Logger
	app    *fiber.App

	durable durable.Durable
}

func New(cfg Config) (*Server, error) {

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(recover.New(recover.Config{EnableStackTrace: true, StackTraceHandler: func(c *fiber.Ctx, err interface{}) {

		buf := make([]byte, 2048)
		buf = buf[:runtime.Stack(buf, false)]
		cfg.Logger.Error().Interface("err", err).Str("stacktrace", string(buf)).Msg("recovered from panic")
	}}))

	app.Use(cors.New(cors.Config{
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "*",
		AllowOrigins:     "*",
		AllowCredentials: true,
	}))

	s := &Server{
		logger:  cfg.Logger,
		app:     app,
		durable: cfg.Durable,
	}

	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	v1 := s.app.Group("/v1")
	v1.Post("/enqueue", s.enqueue)
	v1.Get("/poll/:workflowId", s.poll)

	return s, nil

}
func (s *Server) Listen(addr string) error {
	s.logger.Info().Str("addr", addr).Msg("Listening")
	return s.app.Listen(addr)
}

func (s *Server) Close() error {
	return s.app.Server().Shutdown()
}
