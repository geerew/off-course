package api

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerMiddleware logs the request and response details
func loggerMiddleware(config *RouterConfig) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		started := time.Now()
		err := c.Next()

		attrs := make([]any, 0, 15)

		attrs = append(attrs, slog.String("type", "request"))

		if !started.IsZero() {
			attrs = append(attrs, slog.Float64("execTime", float64(time.Since(started))/float64(time.Millisecond)))
		}

		if err != nil {
			attrs = append(
				attrs,
				slog.String("error", err.Error()),
			)
		}

		// Get the host and extract the port
		host := string(c.Request().Host())
		port := ""
		if strings.Contains(host, ":") {
			port = strings.Split(host, ":")[1]
		}

		method := strings.ToUpper(c.Method())
		attrs = append(
			attrs,
			slog.Int("status", c.Response().StatusCode()),
			slog.String("method", method),
			slog.String("port", port),
		)

		message := method + " " + c.OriginalURL()

		// Determine if the response is an error
		isErrorResponse := err != nil

		var jsonBody map[string]interface{}
		if jsonErr := json.Unmarshal(c.Response().Body(), &jsonBody); jsonErr == nil {
			if val, exists := jsonBody["message"]; exists {
				attrs = append(attrs, slog.String("message", val.(string)))
			}

			if !isErrorResponse {
				if val, exists := jsonBody["error"]; exists {
					isErrorResponse = true
					attrs = append(
						attrs, slog.String("error", val.(string)))
				}
			}
		}

		if isErrorResponse {
			config.Logger.Error(message, attrs...)
		} else {
			config.Logger.Info(message, attrs...)
		}

		return err
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// corsMiddleWare creates a CORS middleware
func corsMiddleWare() func(c *fiber.Ctx) error {
	return cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, DELETE, HEAD, PATCH",
	})
}
