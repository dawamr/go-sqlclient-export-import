package models

// ExportForm represents the form data for exporting a database
type ExportForm struct {
	Type     string `form:"type"`
	Host     string `form:"host"`
	Port     string `form:"port"`
	Database string `form:"database"`
	Username string `form:"username"`
	Password string `form:"password"`
}

// ImportForm represents the form data for importing a database
type ImportForm struct {
	Type     string `form:"type"`
	Host     string `form:"host"`
	Port     string `form:"port"`
	Database string `form:"database"`
	Username string `form:"username"`
	Password string `form:"password"`
}

// ConnectionForm represents the form data for database connection
type ConnectionForm struct {
	Type     string `form:"type"`
	Host     string `form:"host"`
	Port     string `form:"port"`
	Username string `form:"username"`
	Password string `form:"password"`
}

// DatabaseOperation represents the form data for database operations
type DatabaseOperation struct {
	Type        string `form:"type"`
	Host        string `form:"host"`
	Port        string `form:"port"`
	Username    string `form:"username"`
	Password    string `form:"password"`
	Database    string `form:"database"`
	NewDatabase string `form:"newDatabase"`
	Operation   string `form:"operation"`
}

// Database represents a database in the list
type Database struct {
	Name string
	Size string
}
