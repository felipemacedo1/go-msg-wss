# go-msg-wss

## Description

go-msg-wss is a Go-based microservice designed to manage chat rooms and their messages. This microservice provides endpoints to create and manage rooms, send and retrieve messages, and handle reactions to messages, functioning as a self-contained unit within a broader microservices architecture.

## Installation

### Prerequisites

- Go (version 1.16+)
- PostgreSQL
- Docker & Docker Compose

### Steps to Install

1. Clone the repository:
    ```bash
    git clone https://github.com/felipemacedo1/go-msg-wss.git
    cd go-msg-wss
    ```

2. Install dependencies:
    ```bash
    go mod tidy
    ```

3. Setup the database:
    - Ensure PostgreSQL is running.
    - Apply the migrations using the provided SQL files.

4. Start the microservice using Docker Compose:
    ```bash
    docker-compose up
    ```

## Configuration

The service uses environment variables for configuration. Create a `.env` file in the root directory or set these variables in your environment:

```bash
# Copy the example environment file
cp .env.example .env

# Edit the .env file with your configuration
nano .env
```

### Required Environment Variables

- `MSGWSS_DATABASE_HOST`: Database host (e.g., `localhost`)
- `MSGWSS_DATABASE_PORT`: Database port (e.g., `5432`)
- `MSGWSS_DATABASE_USER`: Database username (e.g., `postgres`)
- `MSGWSS_DATABASE_PASSWORD`: Database password
- `MSGWSS_DATABASE_NAME`: Database name (e.g., `msgwss`)
- `MSGWSS_JWT_SECRET`: Secret key for JWT token generation and validation (required for security)

### Optional Environment Variables

- `PORT`: The port on which the service will run (default: `8080`)
- `LOG_LEVEL`: Logging level (`debug`, `info`, `warn`, `error`) (default: `info`)
- `ALLOWED_ORIGINS`: Comma-separated list of allowed CORS origins (default: `*`)

You can also configure these in the `docker-compose.yml` file if you are using Docker.

## Usage

### Starting the Microservice

To start the microservice, use the following command:

```bash
go run main.go
```

### API Endpoints

The following endpoints are available in the go-msg-wss:

- **Rooms**
  - `POST /api/rooms`: Create a new room
  - `GET /api/rooms`: Get all rooms
  - `GET /api/rooms/:room_id`: Get a specific room

- **Messages**
  - `POST /api/rooms/:room_id/messages`: Create a new message in a room
  - `GET /api/rooms/:room_id/messages`: Get all messages in a room
  - `GET /api/rooms/:room_id/messages/:message_id`: Get a specific message
  - `PATCH /api/rooms/:room_id/messages/:message_id/react`: Add a reaction to a message
  - `DELETE /api/rooms/:room_id/messages/:message_id/react`: Remove a reaction from a message
  - `PATCH /api/rooms/:room_id/messages/:message_id/answer`: Mark a message as answered

## Project Structure

- **main.go**: The entry point of the microservice.
- **api.go**: Handles the routing and API endpoints.
- **db.go**: Database connection and related operations.
- **models.go**: Data models for the microservice.
- **queries.sql**: SQL queries used in the microservice.
- **001_create_rooms_table.sql**: SQL migration for creating the rooms table.
- **002_create_messages_table.sql**: SQL migration for creating the messages table.
- **tern.conf**: Configuration file for database migrations.
- **compose.yml**: Docker Compose file for containerized deployment.

## Database

The database is managed using PostgreSQL. The following SQL migrations are used to set up the database:

- **001_create_rooms_table.sql**: Creates the `rooms` table.
- **002_create_messages_table.sql**: Creates the `messages` table.

Run these migrations to set up your database before starting the microservice.
