package api

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/geerew/off-course/ui"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (r *Router) bindUi() {
	if r.config.IsProduction {
		r.App.Use(filesystem.New(filesystem.Config{
			Root: http.FS(ui.Assets()),
		}))

		// Create a fallback for 'dynamic' pages
		fallback := func(c *fiber.Ctx) error {
			if strings.HasPrefix(c.OriginalURL(), "/api") {
				return c.Next()
			}

			f, err := ui.Assets().Open("200.html")
			if err != nil {
				return c.Status(500).SendString("could not open SPA fallback")
			}

			data, _ := io.ReadAll(f)
			c.Type("html", "utf-8")

			return c.Send(data)
		}

		r.App.Get("/*", fallback)
		r.App.Head("/*", fallback)
	} else {
		r.App.Use(func(c *fiber.Ctx) error {
			if strings.HasPrefix(c.OriginalURL(), "/api") {
				return c.Next()
			}

			// This is the port that the UI dev server exposes. If this changes, update the
			// DEV_UI_PORT environment variable in the .air.toml file
			port := os.Getenv("DEV_UI_PORT")
			if port == "" {
				port = "5173"
			}

			uri := "http://localhost:" + port + c.OriginalURL()

			var err error
			err = proxy.Do(c, uri)
			if err != nil {
				// Sometimes svelte closes the connection before returning the first response
				// byte. This just attempts the proxy again
				err = proxy.Do(c, uri)
			}
			return err
		})
	}
}
