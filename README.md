# Go RESTful API with MongoDB

A complete RESTful API built with Go, Gin framework, and MongoDB - structured similar to Laravel for easy understanding.

## 🚀 Features

- **RESTful API** with proper HTTP methods and status codes
- **MongoDB Integration** with connection pooling and proper error handling
- **JWT Authentication** with role-based access control
- **Input Validation** using struct tags (similar to Laravel validation)
- **Middleware Support** for CORS, logging, and authentication
- **Structured Project** with clear separation of concerns
- **Environment Configuration** using .env files
- **Error Handling** with consistent API responses

## 📁 Project Structure

```
tools-backend/
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── env.example            # Environment variables template
├── config/                 # Configuration management
│   └── config.go
├── database/              # Database connection
│   └── connection.go
├── models/                 # Data models (similar to Laravel Models)
│   ├── user.go
│   └── product.go
├── controllers/           # Business logic (similar to Laravel Controllers)
│   ├── auth_controller.go
│   └── product_controller.go
├── middleware/           # Middleware functions (similar to Laravel Middleware)
│   ├── auth.go
│   ├── cors.go
│   └── logger.go
├── routes/               # Route definitions (similar to Laravel Routes)
│   └── routes.go
└── utils/                # Utility functions
    ├── response.go
    ├── validation.go
    └── jwt.go
```

## 🛠️ Installation & Setup

### Prerequisites

- Go 1.25 or higher
- MongoDB 4.4 or higher
- Git

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd tools-backend
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Environment Configuration

Copy the example environment file and configure your settings:

```bash
cp env.example .env
```

Edit `.env` file with your configuration:

```env
# Database Configuration
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=tools_db

# Server Configuration
PORT=8080

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key

# Environment
APP_ENV=development
```

### 4. Start MongoDB

Make sure MongoDB is running on your system:

```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Or start your local MongoDB service
mongod
```

### 5. Run the Application

#### Option 1: Quick Start (Recommended)
```bash
# Windows PowerShell
.\script.ps1

```
#### Option 2: manual Commands
```bash
# Install dependencies
go mod tidy

# Create environment file
cp env.example .env

# Start MongoDB (if not running)
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Run the application
go run main.go
```

#### Option 3: Docker Compose (Full Stack)
```bash
# Start everything with Docker
docker-compose up -d

# View logs
docker-compose logs -f api
```

The server will start on `http://localhost:8080`

## 📚 API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/register` | Register new user | No |
| POST | `/api/v1/login` | User login | No |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Server health check |

## 🔧 Usage Examples

### 1. Register a User

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

## 🔒 Authentication & Authorization

The API uses JWT (JSON Web Tokens) for authentication:

1. **Register/Login** to get a JWT token
2. **Include token** in Authorization header: `Bearer <token>`
3. **Role-based access** for admin-only endpoints

### Token Structure

```json
{
  "user_id": "user_id_here",
  "email": "user@example.com",
  "role": "user|admin",
  "exp": 1234567890,
  "iat": 1234567890
}
```
## 📊 Database Schema

### Users Collection

```json
{
  "_id": "ObjectId",
  "name": "string",
  "email": "string (unique)",
  "password": "string (hashed)",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```
```
## 🚀 Production Deployment

### 1. Build the Application

```bash
go build -o tools-backend main.go
```

### 2. Environment Variables

Set production environment variables:

```bash
export MONGODB_URI="mongodb://your-production-db"
export JWT_SECRET="your-production-secret"
export PORT="8080"
```

### 3. Run the Application

```bash
./tools-backend
```

### Manual Development Commands

#### 1. Hot Reload (Development)
```bash
# Install air for hot reloading
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
# OR
make dev
```

#### 2. Code Formatting
```bash
go fmt ./...
go vet ./...
```

#### 3. Dependency Management
```bash
# Add new dependency
go get github.com/package/name

# Update dependencies
go mod tidy
```

#### 4. Docker Development
```bash
# Start with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop everything
docker-compose down
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Troubleshooting

### Common Issues

1. **MongoDB Connection Error**
   - Ensure MongoDB is running
   - Check connection string in `.env`

2. **JWT Token Issues**
   - Verify JWT_SECRET is set
   - Check token format in Authorization header

3. **Port Already in Use**
   - Change PORT in `.env` file
   - Kill existing process using the port


**Happy Coding! 🎉**

This API structure closely follows Laravel conventions, making it easy to transition from PHP/Laravel to Go development.
