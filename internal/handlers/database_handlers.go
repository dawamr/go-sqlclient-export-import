package handlers

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sqlclient-export-import/internal/models"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ManagePageHandler renders the database management page
func ManagePageHandler(c *fiber.Ctx) error {
	return c.Render("manage", fiber.Map{
		"Title": "Manage Databases",
	})
}

// ListDatabasesHandler handles the request to list databases
func ListDatabasesHandler(c *fiber.Ctx) error {
	// Parse form
	var connForm models.ConnectionForm
	if err := c.BodyParser(&connForm); err != nil {
		return c.Status(fiber.StatusBadRequest).Render("manage", fiber.Map{
			"Title": "Manage Databases",
			"Error": "Invalid form data: " + err.Error(),
		})
	}

	// Validate form data
	if connForm.Host == "" || connForm.Username == "" {
		return c.Status(fiber.StatusBadRequest).Render("manage", fiber.Map{
			"Title":      "Manage Databases",
			"Error":      "Please fill in all required fields",
			"Connection": connForm,
		})
	}

	// Set default port if not provided
	if connForm.Port == "" {
		switch connForm.Type {
		case "mysql", "mariadb":
			connForm.Port = "3306"
		case "postgres":
			connForm.Port = "5432"
		}
	}

	// Get list of databases
	databases, err := listDatabases(connForm)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("manage", fiber.Map{
			"Title":      "Manage Databases",
			"Error":      "Failed to list databases: " + err.Error(),
			"Connection": connForm,
		})
	}

	// Return success
	return c.Render("manage", fiber.Map{
		"Title":      "Manage Databases",
		"Success":    "Databases listed successfully",
		"Connection": connForm,
		"Databases":  databases,
	})
}

// DatabaseOperationHandler handles database operations (create, rename, drop)
func DatabaseOperationHandler(c *fiber.Ctx) error {
	// Parse form
	var dbOp models.DatabaseOperation
	if err := c.BodyParser(&dbOp); err != nil {
		return c.Status(fiber.StatusBadRequest).Render("manage", fiber.Map{
			"Title": "Manage Databases",
			"Error": "Invalid form data: " + err.Error(),
		})
	}

	// Validate form data
	if dbOp.Host == "" || dbOp.Username == "" || dbOp.Operation == "" {
		return c.Status(fiber.StatusBadRequest).Render("manage", fiber.Map{
			"Title":     "Manage Databases",
			"Error":     "Please fill in all required fields",
			"Operation": dbOp,
		})
	}

	// Set default port if not provided
	if dbOp.Port == "" {
		switch dbOp.Type {
		case "mysql", "mariadb":
			dbOp.Port = "3306"
		case "postgres":
			dbOp.Port = "5432"
		}
	}

	// Perform the operation
	var err error
	var successMsg string

	switch dbOp.Operation {
	case "create":
		if dbOp.NewDatabase == "" {
			return c.Status(fiber.StatusBadRequest).Render("manage", fiber.Map{
				"Title":     "Manage Databases",
				"Error":     "Please provide a name for the new database",
				"Operation": dbOp,
			})
		}
		err = createDatabase(dbOp)
		successMsg = fmt.Sprintf("Database '%s' created successfully", dbOp.NewDatabase)
	case "rename":
		if dbOp.Database == "" || dbOp.NewDatabase == "" {
			return c.Status(fiber.StatusBadRequest).Render("manage", fiber.Map{
				"Title":     "Manage Databases",
				"Error":     "Please provide both source and target database names",
				"Operation": dbOp,
			})
		}
		err = renameDatabase(dbOp)
		successMsg = fmt.Sprintf("Database '%s' renamed to '%s' successfully", dbOp.Database, dbOp.NewDatabase)
	case "drop":
		if dbOp.Database == "" {
			return c.Status(fiber.StatusBadRequest).Render("manage", fiber.Map{
				"Title":     "Manage Databases",
				"Error":     "Please provide the database name to drop",
				"Operation": dbOp,
			})
		}
		err = dropDatabase(dbOp)
		successMsg = fmt.Sprintf("Database '%s' dropped successfully", dbOp.Database)
	default:
		return c.Status(fiber.StatusBadRequest).Render("manage", fiber.Map{
			"Title":     "Manage Databases",
			"Error":     "Invalid operation",
			"Operation": dbOp,
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("manage", fiber.Map{
			"Title":     "Manage Databases",
			"Error":     fmt.Sprintf("Failed to %s database: %v", dbOp.Operation, err),
			"Operation": dbOp,
		})
	}

	// Get updated list of databases
	connForm := models.ConnectionForm{
		Type:     dbOp.Type,
		Host:     dbOp.Host,
		Port:     dbOp.Port,
		Username: dbOp.Username,
		Password: dbOp.Password,
	}

	databases, err := listDatabases(connForm)
	if err != nil {
		log.Printf("Failed to list databases after operation: %v", err)
	}

	// Return success
	return c.Render("manage", fiber.Map{
		"Title":      "Manage Databases",
		"Success":    successMsg,
		"Connection": connForm,
		"Databases":  databases,
	})
}

// Helper function to list databases
func listDatabases(conn models.ConnectionForm) ([]models.Database, error) {
	var cmd *exec.Cmd
	var stdout, stderr bytes.Buffer
	var databases []models.Database

	switch conn.Type {
	case "mysql", "mariadb":
		args := []string{
			"-h", conn.Host,
			"-P", conn.Port,
			"-u", conn.Username,
		}

		if conn.Password != "" {
			args = append(args, "-p"+conn.Password)
		}

		args = append(args, "-e", "SHOW DATABASES;")

		// Log the command (without password)
		log.Printf("Running mysql command: mysql -h %s -P %s -u %s -e \"SHOW DATABASES;\"",
			conn.Host, conn.Port, conn.Username)

		cmd = exec.Command("mysql", args...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

	case "postgres":
		env := []string{}
		if conn.Password != "" {
			env = append(env, "PGPASSWORD="+conn.Password)
		}

		args := []string{
			"-h", conn.Host,
			"-p", conn.Port,
			"-U", conn.Username,
			"-t", // Tuples only, no headers
			"-c", "SELECT datname FROM pg_database WHERE datistemplate = false;",
		}

		// Log the command (without password)
		log.Printf("Running psql command: psql -h %s -p %s -U %s -t -c \"SELECT datname FROM pg_database WHERE datistemplate = false;\"",
			conn.Host, conn.Port, conn.Username)

		cmd = exec.Command("psql", args...)
		cmd.Env = append(os.Environ(), env...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

	default:
		return nil, fmt.Errorf("unsupported database type: %s", conn.Type)
	}

	// Execute the command
	if err := cmd.Run(); err != nil {
		errOutput := stderr.String()
		if errOutput != "" {
			return nil, fmt.Errorf("%v: %s", err, errOutput)
		}
		return nil, err
	}

	// Parse the output
	output := stdout.String()
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Database") || strings.HasPrefix(line, "----") {
			continue
		}

		// For MySQL, the output is just the database name
		// For PostgreSQL, the output might have spaces, so we trim them
		dbName := strings.TrimSpace(line)

		// Skip system databases
		if isSystemDatabase(dbName, conn.Type) {
			continue
		}

		databases = append(databases, models.Database{
			Name: dbName,
			Size: "N/A", // Size calculation would require additional queries
		})
	}

	return databases, nil
}

// Helper function to check if a database is a system database
func isSystemDatabase(dbName string, dbType string) bool {
	switch dbType {
	case "mysql", "mariadb":
		systemDBs := []string{"information_schema", "mysql", "performance_schema", "sys"}
		for _, sysDB := range systemDBs {
			if strings.EqualFold(dbName, sysDB) {
				return true
			}
		}
	case "postgres":
		systemDBs := []string{"postgres", "template0", "template1"}
		for _, sysDB := range systemDBs {
			if strings.EqualFold(dbName, sysDB) {
				return true
			}
		}
	}
	return false
}

// Helper function to create a database
func createDatabase(dbOp models.DatabaseOperation) error {
	var cmd *exec.Cmd
	var stderr bytes.Buffer

	switch dbOp.Type {
	case "mysql", "mariadb":
		args := []string{
			"-h", dbOp.Host,
			"-P", dbOp.Port,
			"-u", dbOp.Username,
		}

		if dbOp.Password != "" {
			args = append(args, "-p"+dbOp.Password)
		}

		args = append(args, "-e", fmt.Sprintf("CREATE DATABASE `%s`;", dbOp.NewDatabase))

		// Log the command (without password)
		log.Printf("Running mysql command: mysql -h %s -P %s -u %s -e \"CREATE DATABASE `%s`;\"",
			dbOp.Host, dbOp.Port, dbOp.Username, dbOp.NewDatabase)

		cmd = exec.Command("mysql", args...)
		cmd.Stderr = &stderr

	case "postgres":
		env := []string{}
		if dbOp.Password != "" {
			env = append(env, "PGPASSWORD="+dbOp.Password)
		}

		args := []string{
			"-h", dbOp.Host,
			"-p", dbOp.Port,
			"-U", dbOp.Username,
			"-c", fmt.Sprintf("CREATE DATABASE \"%s\";", dbOp.NewDatabase),
		}

		// Log the command (without password)
		log.Printf("Running psql command: psql -h %s -p %s -U %s -c \"CREATE DATABASE \"%s\";\"",
			dbOp.Host, dbOp.Port, dbOp.Username, dbOp.NewDatabase)

		cmd = exec.Command("psql", args...)
		cmd.Env = append(os.Environ(), env...)
		cmd.Stderr = &stderr

	default:
		return fmt.Errorf("unsupported database type: %s", dbOp.Type)
	}

	// Execute the command
	if err := cmd.Run(); err != nil {
		errOutput := stderr.String()
		if errOutput != "" {
			return fmt.Errorf("%v: %s", err, errOutput)
		}
		return err
	}

	return nil
}

// Helper function to rename a database
func renameDatabase(dbOp models.DatabaseOperation) error {
	var cmd *exec.Cmd
	var stderr bytes.Buffer

	switch dbOp.Type {
	case "mysql", "mariadb":
		// MySQL doesn't have a direct RENAME DATABASE command
		// We need to create a new database and transfer all tables
		args := []string{
			"-h", dbOp.Host,
			"-P", dbOp.Port,
			"-u", dbOp.Username,
		}

		if dbOp.Password != "" {
			args = append(args, "-p"+dbOp.Password)
		}

		// Create new database
		createCmd := exec.Command("mysql", append(args, "-e", fmt.Sprintf("CREATE DATABASE `%s`;", dbOp.NewDatabase))...)
		var createStderr bytes.Buffer
		createCmd.Stderr = &createStderr

		if err := createCmd.Run(); err != nil {
			errOutput := createStderr.String()
			if errOutput != "" {
				return fmt.Errorf("failed to create target database: %v: %s", err, errOutput)
			}
			return fmt.Errorf("failed to create target database: %v", err)
		}

		// Export source database
		exportArgs := append([]string{}, args...)
		exportArgs = append(exportArgs, dbOp.Database)

		exportCmd := exec.Command("mysqldump", exportArgs...)
		var exportStdout bytes.Buffer
		exportCmd.Stdout = &exportStdout
		var exportStderr bytes.Buffer
		exportCmd.Stderr = &exportStderr

		if err := exportCmd.Run(); err != nil {
			errOutput := exportStderr.String()
			if errOutput != "" {
				return fmt.Errorf("failed to export source database: %v: %s", err, errOutput)
			}
			return fmt.Errorf("failed to export source database: %v", err)
		}

		// Import to new database
		importArgs := append([]string{}, args...)
		importArgs = append(importArgs, dbOp.NewDatabase)

		importCmd := exec.Command("mysql", importArgs...)
		importCmd.Stdin = &exportStdout
		var importStderr bytes.Buffer
		importCmd.Stderr = &importStderr

		if err := importCmd.Run(); err != nil {
			errOutput := importStderr.String()
			if errOutput != "" {
				return fmt.Errorf("failed to import to target database: %v: %s", err, errOutput)
			}
			return fmt.Errorf("failed to import to target database: %v", err)
		}

		// Drop old database
		dropArgs := append([]string{}, args...)
		dropArgs = append(dropArgs, "-e", fmt.Sprintf("DROP DATABASE `%s`;", dbOp.Database))

		dropCmd := exec.Command("mysql", dropArgs...)
		var dropStderr bytes.Buffer
		dropCmd.Stderr = &dropStderr

		if err := dropCmd.Run(); err != nil {
			errOutput := dropStderr.String()
			if errOutput != "" {
				return fmt.Errorf("failed to drop source database (rename partially completed): %v: %s", err, errOutput)
			}
			return fmt.Errorf("failed to drop source database (rename partially completed): %v", err)
		}

		return nil

	case "postgres":
		env := []string{}
		if dbOp.Password != "" {
			env = append(env, "PGPASSWORD="+dbOp.Password)
		}

		args := []string{
			"-h", dbOp.Host,
			"-p", dbOp.Port,
			"-U", dbOp.Username,
			"-c", fmt.Sprintf("ALTER DATABASE \"%s\" RENAME TO \"%s\";", dbOp.Database, dbOp.NewDatabase),
		}

		// Log the command (without password)
		log.Printf("Running psql command: psql -h %s -p %s -U %s -c \"ALTER DATABASE \"%s\" RENAME TO \"%s\";\"",
			dbOp.Host, dbOp.Port, dbOp.Username, dbOp.Database, dbOp.NewDatabase)

		cmd = exec.Command("psql", args...)
		cmd.Env = append(os.Environ(), env...)
		cmd.Stderr = &stderr

	default:
		return fmt.Errorf("unsupported database type: %s", dbOp.Type)
	}

	// Execute the command if not MySQL (MySQL is handled separately above)
	if dbOp.Type != "mysql" && dbOp.Type != "mariadb" {
		if err := cmd.Run(); err != nil {
			errOutput := stderr.String()
			if errOutput != "" {
				return fmt.Errorf("%v: %s", err, errOutput)
			}
			return err
		}
	}

	return nil
}

// Helper function to drop a database
func dropDatabase(dbOp models.DatabaseOperation) error {
	var cmd *exec.Cmd
	var stderr bytes.Buffer

	switch dbOp.Type {
	case "mysql", "mariadb":
		args := []string{
			"-h", dbOp.Host,
			"-P", dbOp.Port,
			"-u", dbOp.Username,
		}

		if dbOp.Password != "" {
			args = append(args, "-p"+dbOp.Password)
		}

		args = append(args, "-e", fmt.Sprintf("DROP DATABASE `%s`;", dbOp.Database))

		// Log the command (without password)
		log.Printf("Running mysql command: mysql -h %s -P %s -u %s -e \"DROP DATABASE `%s`;\"",
			dbOp.Host, dbOp.Port, dbOp.Username, dbOp.Database)

		cmd = exec.Command("mysql", args...)
		cmd.Stderr = &stderr

	case "postgres":
		env := []string{}
		if dbOp.Password != "" {
			env = append(env, "PGPASSWORD="+dbOp.Password)
		}

		args := []string{
			"-h", dbOp.Host,
			"-p", dbOp.Port,
			"-U", dbOp.Username,
			"-c", fmt.Sprintf("DROP DATABASE \"%s\";", dbOp.Database),
		}

		// Log the command (without password)
		log.Printf("Running psql command: psql -h %s -p %s -U %s -c \"DROP DATABASE \"%s\";\"",
			dbOp.Host, dbOp.Port, dbOp.Username, dbOp.Database)

		cmd = exec.Command("psql", args...)
		cmd.Env = append(os.Environ(), env...)
		cmd.Stderr = &stderr

	default:
		return fmt.Errorf("unsupported database type: %s", dbOp.Type)
	}

	// Execute the command
	if err := cmd.Run(); err != nil {
		errOutput := stderr.String()
		if errOutput != "" {
			return fmt.Errorf("%v: %s", err, errOutput)
		}
		return err
	}

	return nil
}
