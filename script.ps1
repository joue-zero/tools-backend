# Go RESTful API Startup Script
# This script sets up and runs your Go API with MongoDB

Write-Host "🚀 Setting up Go RESTful API with MongoDB..." -ForegroundColor Green

# Check if Go is installed
Write-Host "📋 Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✅ Go is installed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go is not installed. Please install Go from https://golang.org/dl/" -ForegroundColor Red
    exit 1
}

# Check if MongoDB is running
Write-Host "📋 Checking MongoDB connection..." -ForegroundColor Yellow
try {
    $mongoTest = Test-NetConnection -ComputerName localhost -Port 27017 -InformationLevel Quiet
    if ($mongoTest) {
        Write-Host "✅ MongoDB is running on localhost:27017" -ForegroundColor Green
    } else {
        Write-Host "⚠️  MongoDB is not running. Starting MongoDB with Docker..." -ForegroundColor Yellow
        Write-Host "🐳 Starting MongoDB container..." -ForegroundColor Cyan
        docker run -d -p 27017:27017 --name mongodb mongo:latest
        Start-Sleep -Seconds 5
        Write-Host "✅ MongoDB container started" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️  MongoDB not detected. Starting with Docker..." -ForegroundColor Yellow
    Write-Host "🐳 Starting MongoDB container..." -ForegroundColor Cyan
    docker run -d -p 27017:27017 --name mongodb mongo:latest
    Start-Sleep -Seconds 5
    Write-Host "✅ MongoDB container started" -ForegroundColor Green
}

# Create .env file if it doesn't exist
if (-not (Test-Path ".env")) {
    Write-Host "📝 Creating .env file from template..." -ForegroundColor Yellow
    Copy-Item "env.example" ".env"
    Write-Host "✅ .env file created" -ForegroundColor Green
} else {
    Write-Host "✅ .env file already exists" -ForegroundColor Green
}

# Install dependencies
Write-Host "📦 Installing Go dependencies..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Dependencies installed successfully" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to install dependencies" -ForegroundColor Red
    exit 1
}

# Build the application
Write-Host "🔨 Building the application..." -ForegroundColor Yellow
go build -o tools-backend.exe main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Application built successfully" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to build application" -ForegroundColor Red
    exit 1
}

# Start the server
Write-Host "🚀 Starting the Go API server..." -ForegroundColor Green
Write-Host "📍 Server will be available at: http://localhost:8080" -ForegroundColor Cyan
Write-Host "📚 API Documentation: Check README.md for endpoints" -ForegroundColor Cyan
Write-Host "🛑 Press Ctrl+C to stop the server" -ForegroundColor Yellow
Write-Host ""

# Run the application
./tools-backend.exe