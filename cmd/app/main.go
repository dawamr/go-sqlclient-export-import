package main

import (
	"log"
	"os"
	"time"

	"sqlclient-export-import/internal/config"
	"sqlclient-export-import/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	// Load configuration
	cfg := config.New()

	// Create required directories
	createDirectories(cfg)

	// Initialize handlers with config
	handlers.Initialize(cfg)

	// Set up the template engine
	engine := html.New(cfg.TemplateDir, ".html")
	engine.Reload(cfg.IsDevelopment()) // Enable reloading templates in development

	// Add template functions
	engine.AddFunc("currentYear", func() string {
		return time.Now().Format("2006")
	})

	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		Views:                 engine,
		ViewsLayout:           "layouts/main",         // Use a layout for all templates
		BodyLimit:             int(cfg.MaxUploadSize), // Use MaxUploadSize from config
		ReadTimeout:           10 * time.Minute,       // Increase read timeout for large file uploads
		WriteTimeout:          10 * time.Minute,       // Increase write timeout for large file uploads
		IdleTimeout:           10 * time.Minute,       // Increase idle timeout
		DisableStartupMessage: false,                  // Show startup message
		StreamRequestBody:     true,                   // Enable streaming request body for large files
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Handle 404 errors
			if err != nil {
				log.Printf("Error: %v", err)
			}

			code := fiber.StatusInternalServerError

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Return error page
			return c.Status(code).Render("error", fiber.Map{
				"Title":   "Error",
				"Code":    code,
				"Message": err.Error(),
			})
		},
	})

	// Set up middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Add middleware to handle large files first (before CORS)
	app.Use(func(c *fiber.Ctx) error {
		// Set high limits for multipart forms
		c.Set("Content-Type", "multipart/form-data")
		return c.Next()
	})

	app.Use(cors.New())

	// Add middleware to prevent caching
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("Surrogate-Control", "no-store")
		return c.Next()
	})

	// Serve static files
	app.Static("/static", cfg.StaticDir)

	// Set up routes
	setupRoutes(app)

	// Start the server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func setupRoutes(app *fiber.App) {
	// Home route
	app.Get("/", handlers.HomeHandler)

	// Database routes
	dbGroup := app.Group("/db")
	dbGroup.Get("/export", handlers.ExportPageHandler)
	dbGroup.Post("/export", handlers.ExportDatabaseHandler)
	dbGroup.Get("/download", handlers.DownloadExportHandler)
	dbGroup.Get("/import", handlers.ImportPageHandler)
	dbGroup.Post("/import", handlers.ImportDatabaseHandler)

	// Database management routes
	dbGroup.Get("/manage", handlers.ManagePageHandler)
	dbGroup.Post("/manage/list", handlers.ListDatabasesHandler)
	dbGroup.Post("/manage/operation", handlers.DatabaseOperationHandler)
}

func createDirectories(cfg *config.Config) {
	// Create export directory
	if err := os.MkdirAll(cfg.ExportDirectory, 0755); err != nil {
		log.Fatalf("Failed to create export directory: %v", err)
	}

	// Create upload directory
	if err := os.MkdirAll(cfg.UploadDirectory, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}
}
