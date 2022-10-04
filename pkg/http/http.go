package http

import (
	"fmt"
	"net"

	"github.com/goccy/go-json"
	"github.com/minio/simdjson-go"
	"github.com/rs/zerolog/log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	fiber_utils "github.com/gofiber/fiber/v2/utils"

	"github.com/nano-interactive/go-utils"
	"github.com/nano-interactive/go-utils/environment"

	"github.com/nano-interactive/GoFiber-Boilerplate/pkg/constants"
	"github.com/nano-interactive/GoFiber-Boilerplate/pkg/container"
	"github.com/nano-interactive/GoFiber-Boilerplate/pkg/http/middleware"
)

func CreateApplication(c *container.Container, appName string, env environment.Env, displayInfo, enableMonitor bool, errorHandler fiber.ErrorHandler) *fiber.App {
	var (
		jsonEncoder fiber_utils.JSONMarshal = json.Marshal
		jsonDecoder fiber_utils.JSONUnmarshal
	)

	if simdjson.SupportedCPU() {
		jsonDecoder = func(data []byte, v interface{}) error {
			parsed, err := simdjson.Parse(data, nil)
			if err != nil {
				return err
			}

			parsed.Iter()
			return nil
		}
	} else {
		log.Warn().
			Str("app", appName).
			Str("env", string(env)).
			Msg("simdjson is not supported on this CPU, application performance might suffer")

		jsonDecoder = json.Unmarshal
	}

	staticConfig := fiber.Config{
		StrictRouting:                true,
		EnablePrintRoutes:            displayInfo,
		Prefork:                      false,
		DisableStartupMessage:        !displayInfo,
		DisableDefaultDate:           true,
		DisableHeaderNormalizing:     false,
		DisablePreParseMultipartForm: true,
		AppName:                      appName,
		ErrorHandler:                 errorHandler,
		JSONEncoder:                  jsonEncoder,
		JSONDecoder:                  jsonDecoder,
	}

	app := fiber.New(staticConfig)

	switch env {
	case environment.Development:
		app.Use(pprof.New())
	case environment.Production:
		app.Use(recover.New())
	}

	app.Use(middleware.Context)
	app.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return utils.RandomString(32)
		},
		ContextKey: constants.RequestIdKey,
	}))

	if env == environment.Development {
		app.Use(logger.New(logger.Config{
			TimeZone: "UTC",
		}))
	}

	if enableMonitor {
		app.Get("/monitor", monitor.New(monitor.Config{
			Title: constants.AppName + " Monitor",
		}))
	}

	registerHandlers(app, c, env)

	return app
}

func RunServer(ip string, port int, app *fiber.App) {
	addr := fmt.Sprintf("%s:%d", ip, port)

	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		log.
			Fatal().
			Err(err).
			Msg("Error while creating net.Listener for HTTP Server")
	}

	err = app.Listener(listener)

	if err != nil {
		log.
			Fatal().
			Err(err).
			Msg("Cannot start Fiber HTTP Server")
	}
}
