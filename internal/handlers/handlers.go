package handlers

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sqlclient-export-import/internal/config"
	"sqlclient-export-import/internal/models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var cfg *config.Config

// Initialize sets up the handlers with the application configuration
func Initialize(c *config.Config) {
	cfg = c
}

// HomeHandler renders the home page
func HomeHandler(c *fiber.Ctx) error {
	return c.Render("home", fiber.Map{
		"Title": "SQL Client - Export/Import Database",
	})
}

// ExportPageHandler renders the export page
func ExportPageHandler(c *fiber.Ctx) error {
	return c.Render("export", fiber.Map{
		"Title": "Export Database",
	})
}

// ExportDatabaseHandler handles the database export request
func ExportDatabaseHandler(c *fiber.Ctx) error {
	// Parse form
	var exportForm models.ExportForm
	if err := c.BodyParser(&exportForm); err != nil {
		return c.Status(fiber.StatusBadRequest).Render("export", fiber.Map{
			"Title": "Export Database",
			"Error": "Invalid form data",
		})
	}

	// Validate form data
	if exportForm.Host == "" || exportForm.Database == "" || exportForm.Username == "" {
		return c.Status(fiber.StatusBadRequest).Render("export", fiber.Map{
			"Title":  "Export Database",
			"Error":  "Please fill in all required fields",
			"Export": exportForm,
		})
	}

	// Set default port if not provided
	if exportForm.Port == "" {
		switch exportForm.Type {
		case "mysql", "mariadb":
			exportForm.Port = "3306"
		case "postgres":
			exportForm.Port = "5432"
		}
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(cfg.ExportDirectory, exportForm.Database+"_"+timestamp+".sql")
	downloadFilename := exportForm.Database + "_" + timestamp + ".sql"

	// Perform the export based on database type
	var cmd *exec.Cmd
	var stderr bytes.Buffer

	switch exportForm.Type {
	case "mysql", "mariadb":
		// For MySQL/MariaDB, we'll use a simpler approach
		args := []string{
			"-h", exportForm.Host,
			"-P", exportForm.Port,
			"-u", exportForm.Username,
			"--column-statistics=0",
		}

		if exportForm.Password != "" {
			// Pass password directly with -p option (no space between -p and password)
			args = append(args, "-p"+exportForm.Password)
		}

		args = append(args, "--databases", exportForm.Database)

		// Log the command (without password)
		log.Printf("Running mysqldump command: mysqldump -h %s -P %s -u %s --column-statistics=0 --databases %s",
			exportForm.Host, exportForm.Port, exportForm.Username, exportForm.Database)

		cmd = exec.Command("mysqldump", args...)
		cmd.Stderr = &stderr
	case "postgres":
		env := os.Environ()
		if exportForm.Password != "" {
			env = append(env, "PGPASSWORD="+exportForm.Password)
		}

		args := []string{
			"-h", exportForm.Host,
			"-p", exportForm.Port,
			"-U", exportForm.Username,
			exportForm.Database,
		}

		// Log the command (without password)
		log.Printf("Running pg_dump command: pg_dump -h %s -p %s -U %s %s",
			exportForm.Host, exportForm.Port, exportForm.Username, exportForm.Database)

		cmd = exec.Command("pg_dump", args...)
		cmd.Env = env
		cmd.Stderr = &stderr
	default:
		return c.Status(fiber.StatusBadRequest).Render("export", fiber.Map{
			"Title":  "Export Database",
			"Error":  "Unsupported database type",
			"Export": exportForm,
		})
	}

	// Open the output file
	outFile, err := os.Create(filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("export", fiber.Map{
			"Title":  "Export Database",
			"Error":  "Failed to create export file: " + err.Error(),
			"Export": exportForm,
		})
	}
	defer outFile.Close()

	// Set the output to the file
	cmd.Stdout = outFile

	// Execute the command
	if err := cmd.Run(); err != nil {
		// If the command fails, remove the file and return an error with stderr output
		os.Remove(filename)

		errorMsg := fmt.Sprintf("Failed to export database: %v", err)
		stderrOutput := stderr.String()

		if stderrOutput != "" {
			errorMsg += "\nDetails: " + stderrOutput
		}

		// Add helpful suggestions based on the error
		if strings.Contains(stderrOutput, "Access denied") {
			errorMsg += "\n\nSuggestions:\n- Check your username and password\n- Ensure the user has permission to access the database"
		} else if strings.Contains(stderrOutput, "Unknown database") {
			errorMsg += "\n\nSuggestions:\n- Check if the database name is correct\n- Ensure the database exists on the server"
		} else if strings.Contains(stderrOutput, "Connection refused") {
			errorMsg += "\n\nSuggestions:\n- Check if the host and port are correct\n- Ensure the database server is running and accessible"
		}

		log.Printf("Export error: %s", errorMsg)

		return c.Status(fiber.StatusInternalServerError).Render("export", fiber.Map{
			"Title":  "Export Database",
			"Error":  errorMsg,
			"Export": exportForm,
		})
	}

	// Log success
	log.Printf("Database exported successfully to %s", filename)

	// Check if the client wants to download the file
	if c.Query("download") == "true" {
		return c.Download(filename, downloadFilename)
	}

	// Return success with download link
	return c.Render("export", fiber.Map{
		"Title":            "Export Database",
		"Success":          "Database exported successfully to " + filename,
		"Export":           exportForm,
		"DownloadLink":     "/db/download?file=" + filepath.Base(filename),
		"DownloadFilename": downloadFilename,
	})
}

// DownloadExportHandler handles downloading exported database files
func DownloadExportHandler(c *fiber.Ctx) error {
	filename := c.Query("file")
	if filename == "" {
		return c.Status(fiber.StatusBadRequest).SendString("No file specified")
	}

	// Ensure the filename is just a basename (no path)
	filename = filepath.Base(filename)

	// Construct the full path
	fullPath := filepath.Join(cfg.ExportDirectory, filename)

	// Check if the file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).SendString("File not found")
	}

	// Send the file as a download
	return c.Download(fullPath, filename)
}

// ImportPageHandler renders the import page
func ImportPageHandler(c *fiber.Ctx) error {
	return c.Render("import", fiber.Map{
		"Title": "Import Database",
	})
}

// ImportDatabaseHandler handles the database import request
func ImportDatabaseHandler(c *fiber.Ctx) error {
	log.Println("Starting database import process")

	// Get the uploaded file
	file, err := c.FormFile("sqlFile")
	if err != nil {
		log.Printf("Error getting uploaded file: %v", err)
		return c.Status(fiber.StatusBadRequest).Render("import", fiber.Map{
			"Title": "Import Database",
			"Error": "Please upload a SQL file: " + err.Error(),
		})
	}

	log.Printf("Received file: %s, size: %d bytes", file.Filename, file.Size)

	// Check file size - log but don't reject immediately
	if file.Size > cfg.MaxUploadSize {
		log.Printf("Warning: File size (%d bytes) exceeds configured limit (%d bytes), but will attempt to process",
			file.Size, cfg.MaxUploadSize)
	}

	// Parse form
	var importForm models.ImportForm
	if err := c.BodyParser(&importForm); err != nil {
		log.Printf("Error parsing form: %v", err)
		return c.Status(fiber.StatusBadRequest).Render("import", fiber.Map{
			"Title": "Import Database",
			"Error": "Invalid form data: " + err.Error(),
		})
	}

	// Validate form data
	if importForm.Host == "" || importForm.Database == "" || importForm.Username == "" {
		log.Println("Missing required fields in form data")
		return c.Status(fiber.StatusBadRequest).Render("import", fiber.Map{
			"Title":  "Import Database",
			"Error":  "Please fill in all required fields",
			"Import": importForm,
		})
	}

	// Set default port if not provided
	if importForm.Port == "" {
		switch importForm.Type {
		case "mysql", "mariadb":
			importForm.Port = "3306"
		case "postgres":
			importForm.Port = "5432"
		}
	}

	// Save the file
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(cfg.UploadDirectory, timestamp+"_"+file.Filename)
	log.Printf("Saving file to: %s", filename)

	if err := c.SaveFile(file, filename); err != nil {
		log.Printf("Error saving file: %v", err)
		return c.Status(fiber.StatusInternalServerError).Render("import", fiber.Map{
			"Title":  "Import Database",
			"Error":  "Failed to save uploaded file: " + err.Error(),
			"Import": importForm,
		})
	}

	log.Printf("File saved successfully: %s", filename)

	// Perform the import based on database type
	var cmd *exec.Cmd
	var stderr bytes.Buffer

	switch importForm.Type {
	case "mysql", "mariadb":
		args := []string{
			"-h", importForm.Host,
			"-P", importForm.Port,
			"-u", importForm.Username,
			"--max_allowed_packet=1G", // Increase max allowed packet size
		}

		if importForm.Password != "" {
			// Pass password directly with -p option (no space between -p and password)
			args = append(args, "-p"+importForm.Password)
		}

		args = append(args, importForm.Database)

		// Log the command (without password)
		log.Printf("Running mysql command: mysql -h %s -P %s -u %s --max_allowed_packet=1G %s",
			importForm.Host, importForm.Port, importForm.Username, importForm.Database)

		cmd = exec.Command("mysql", args...)
		cmd.Stderr = &stderr

		// Open the input file
		inFile, err := os.Open(filename)
		if err != nil {
			log.Printf("Error opening file for import: %v", err)
			return c.Status(fiber.StatusInternalServerError).Render("import", fiber.Map{
				"Title":  "Import Database",
				"Error":  "Failed to open import file: " + err.Error(),
				"Import": importForm,
			})
		}
		defer inFile.Close()

		// Set the input from the file
		cmd.Stdin = inFile
	case "postgres":
		env := os.Environ()
		if importForm.Password != "" {
			env = append(env, "PGPASSWORD="+importForm.Password)
		}

		args := []string{
			"-h", importForm.Host,
			"-p", importForm.Port,
			"-U", importForm.Username,
			"-d", importForm.Database,
			"-f", filename,
		}

		// Log the command (without password)
		log.Printf("Running psql command: psql -h %s -p %s -U %s -d %s -f %s",
			importForm.Host, importForm.Port, importForm.Username, importForm.Database, filename)

		cmd = exec.Command("psql", args...)
		cmd.Env = env
		cmd.Stderr = &stderr
	default:
		log.Printf("Unsupported database type: %s", importForm.Type)
		return c.Status(fiber.StatusBadRequest).Render("import", fiber.Map{
			"Title":  "Import Database",
			"Error":  "Unsupported database type: " + importForm.Type,
			"Import": importForm,
		})
	}

	// Execute the command
	log.Println("Executing import command...")
	if err := cmd.Run(); err != nil {
		errorMsg := fmt.Sprintf("Failed to import database: %v", err)
		stderrOutput := stderr.String()

		if stderrOutput != "" {
			errorMsg += "\nDetails: " + stderrOutput
		}

		// Add helpful suggestions based on the error
		if strings.Contains(stderrOutput, "Access denied") {
			errorMsg += "\n\nSuggestions:\n- Check your username and password\n- Ensure the user has permission to access the database"
		} else if strings.Contains(stderrOutput, "Unknown database") {
			errorMsg += "\n\nSuggestions:\n- Check if the database name is correct\n- Ensure the database exists on the server"
		} else if strings.Contains(stderrOutput, "Connection refused") {
			errorMsg += "\n\nSuggestions:\n- Check if the host and port are correct\n- Ensure the database server is running and accessible"
		}

		log.Printf("Import error: %s", errorMsg)

		return c.Status(fiber.StatusInternalServerError).Render("import", fiber.Map{
			"Title":  "Import Database",
			"Error":  errorMsg,
			"Import": importForm,
		})
	}

	// Return success
	log.Printf("Database imported successfully from %s", file.Filename)
	return c.Render("import", fiber.Map{
		"Title":   "Import Database",
		"Success": "Database imported successfully from " + file.Filename,
		"Import":  importForm,
	})
}
