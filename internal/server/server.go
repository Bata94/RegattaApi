package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	api "github.com/bata94/RegattaApi/internal/handlers/api"
	api_v1 "github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cache"

	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/utils"
	//  _ "github.com/bata94/RegattaApi/docs"
	// "github.com/gofiber/swagger"
)

func Init(frontendEnabled, backendEnabled bool, port int) {
	appName := "RegattaApi"
	log.SetLevel(log.LevelDebug)

	// defWebCacheMid := cache.New(cache.Config{
	// 	Next:         nil,
	// 	Expiration:   2 * time.Minute,
	// 	CacheHeader:  "X-Cache",
	// 	CacheControl: false,
	// 	KeyGenerator: func(c *fiber.Ctx) string {
	// 		return utils.CopyString(c.Path())
	// 	},
	// 	ExpirationGenerator:  nil,
	// 	StoreResponseHeaders: false,
	// 	Storage:              nil,
	// 	MaxBytes:             0,
	// 	Methods:              []string{fiber.MethodGet, fiber.MethodHead},
	// })
	defApiCacheMid := cache.New(cache.Config{
		Next:         nil,
		Expiration:   1 * time.Second,
		CacheHeader:  "X-Cache",
		CacheControl: false,
		KeyGenerator: func(c *fiber.Ctx) string {
			return utils.CopyString(c.Path())
		},
		ExpirationGenerator:  nil,
		StoreResponseHeaders: false,
		Storage:              nil,
		MaxBytes:             0,
		Methods:              []string{fiber.MethodGet, fiber.MethodHead},
	})
	defAssetCacheMid := cache.New(cache.Config{
		Next:         nil,
		Expiration:   30 * time.Minute,
		CacheHeader:  "X-Cache",
		CacheControl: false,
		KeyGenerator: func(c *fiber.Ctx) string {
			return utils.CopyString(c.Path())
		},
		ExpirationGenerator:  nil,
		StoreResponseHeaders: false,
		Storage:              nil,
		MaxBytes:             0,
		Methods:              []string{fiber.MethodGet, fiber.MethodHead},
	})

	// webCompressor := compress.New()
	apiCompressor := compress.New()

	app := fiber.New(fiber.Config{
		ServerHeader:      appName,
		AppName:           appName,
		Prefork:           false,
		ErrorHandler:      ErrorHandler,
		EnablePrintRoutes: false,
		JSONEncoder:       json.Marshal,
		JSONDecoder:       json.Unmarshal,
	})

	app.Use(logger.New(logger.Config{
		Next:          nil,
		Done:          nil,
		Format:        "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat:    "15:04:05",
		TimeZone:      "Local",
		TimeInterval:  500 * time.Millisecond,
		Output:        os.Stdout,
		DisableColors: false,
	}))

	app.Static("/assets", "./assets", fiber.Static{
		Compress:      true,
		CacheDuration: 30 * time.Minute,
	})
	app.Get("/favicon.ico", defAssetCacheMid, func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendFile("./assets/favicon.ico")
	})
	app.Get("/metrics", monitor.New(monitor.Config{Title: appName + " Metrics Page", Refresh: time.Duration(1) * time.Second}))
	app.Get("/metricsApi", monitor.New(monitor.Config{APIOnly: true}))

	if backendEnabled {
		api := app.Group("/api", apiCompressor, defApiCacheMid)
		// api.Get("/docs/*", swagger.HandlerDefault) // default

		v1 := api.Group("/v1")

		athletV1 := v1.Group("/athlet")
		athletV1.Get("", api_v1.GetAllAthlet)
		athletV1.Get("/:uuid", api_v1.GetAthlet)
		athletV1.Post("", api_v1.CreateAthlet)

		leitungV1 := v1.Group("/leitung")
		leitungV1.Post("/drv_meldung_upload", api_v1.DrvMeldungUpload)
	}

	if frontendEnabled {
		// webUi := app.Group("/", defWebCacheMid, webCompressor)
		// webUi.Get("", web_routes.RootIndex)
	}

	log.Fatal(app.Listen(fmt.Sprintf(":%v", port)))
}

var ErrorHandler = func(c *fiber.Ctx, err error) error {
	log.Error("ErrorHandler:", err.Error())
	// Status statusCode defaults to 500
	statusCode := fiber.StatusInternalServerError
	code := 0
	title := "Internal Server Error"
	message := ""
	detail := ""

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		statusCode = e.Code
		message = e.Message

		if e.Code == 404 {
			title = "Not found"
		} else if e.Code == 401 {
			title = "Unauthorized"
		} else if e.Code == 403 {
			title = "Forbidden"
		}
	}

	// Check if its our custom request error
	var apiReqError *api.ReqError
	if errors.As(err, &apiReqError) {
		statusCode = apiReqError.StatusCode
		code = apiReqError.Code
		title = apiReqError.Title
		message = apiReqError.Msg
		detail = apiReqError.Details
	}

	if err == context.DeadlineExceeded {
		statusCode = fiber.StatusRequestTimeout
	}

	if detail != "" {
		log.Error("ErrorDetails: ", detail)
	}

	if code == 0 {
		code = statusCode
	}

	// Return status code with error message
	return c.Status(statusCode).JSON(fiber.Map{
		"statusCode": statusCode,
		"code":       code,
		"error":      title,
		"message":    message,
		// "detail":     detail,
	})
}
