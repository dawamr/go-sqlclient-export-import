# SQL Client - Export/Import Database

A web-based application for exporting and importing database content. Built with Go Fiber for the backend and Tailwind CSS for the frontend.

## Features

- Export databases from MySQL, PostgreSQL, and MariaDB
- Import SQL files into your database
- Simple and intuitive web interface
- Secure password handling
- Support for various database types

## Prerequisites

- Go 1.21 or higher
- MySQL client tools (for MySQL/MariaDB export/import)
- PostgreSQL client tools (for PostgreSQL export/import)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/dawamr/sqlclient-export-import.git
cd sqlclient-export-import
```

2. Install dependencies:

```bash
go mod download
```

3. Configure environment variables (optional):

Copy the example `.env` file and modify as needed:

```bash
cp .env.example .env
```

## Usage

1. Start the server:

```bash
go run cmd/app/main.go
```

2. Open your browser and navigate to `http://localhost:3000`

## Configuration

The application can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Port to run the server on | 3000 |
| ENVIRONMENT | Application environment (development/production) | development |
| MAX_UPLOAD_SIZE | Maximum upload file size in bytes | 52428800 (50MB) |
| EXPORT_DIR | Directory to store exported files | ./exports |
| UPLOAD_DIR | Directory to store uploaded files | ./uploads |
| TEMPLATE_DIR | Directory containing HTML templates | ./internal/templates |
| STATIC_DIR | Directory containing static files | ./static |

## Project Structure

```
sqlclient-export-import/
├── cmd/
│   └── app/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration handling
│   ├── handlers/
│   │   └── handlers.go       # HTTP request handlers
│   ├── models/
│   │   └── models.go         # Data models
│   └── templates/            # HTML templates
│       ├── layouts/
│       │   └── main.html     # Main layout template
│       ├── home.html         # Home page template
│       ├── export.html       # Export page template
│       └── import.html       # Import page template
├── static/                   # Static assets
│   ├── css/
│   │   └── styles.css        # Custom CSS styles
│   └── js/
│       └── main.js           # Client-side JavaScript
├── .env                      # Environment variables
├── .env.example              # Example environment variables
├── go.mod                    # Go module file
├── go.sum                    # Go module checksum file
└── README.md                 # Project documentation
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request 