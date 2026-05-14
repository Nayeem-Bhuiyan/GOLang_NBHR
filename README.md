# Clean build
go clean -cache

# Download all dependencies
go mod tidy

# Build for production
go build -ldflags="-s -w" -o nbhr ./cmd/api

# run project in local development mode
go run ./cmd/api/main.go
