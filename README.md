# Chatter

A real-time 1-to-1 chat backend built with Go.

Chatter is a backend-focused project designed to demonstrate practical backend development concepts including REST APIs, JWT authentication, refresh-token rotation, MariaDB, SQL migrations, sqlc, cursor-based pagination and authenticated WebSocket communication.

The project includes a small browser frontend for manually testing the API and real-time messaging functionality.

> **Important:** This is a backend project, frontend might have bugs and it's only for testing purposes, real product should not look like that.

## Features

* User registration and login
* Password hashing with bcrypt
* JWT access tokens
* Refresh tokens with rotation and revocation
* Authentication middleware
* 1-to-1 conversations
* Conversation membership validation
* Persistent message storage
* Cursor-based message pagination
* Authenticated WebSocket connections
* Real-time message broadcasting
* WebSocket message synchronization after reconnecting
* MariaDB database
* Goose database migrations
* sqlc-generated database queries
* Docker and Docker Compose support
* Environment-based configuration

## Tech Stack

| Technology          | Purpose                             |
| ------------------- | ----------------------------------- |
| Go                  | Backend                             |
| MariaDB             | Relational database                 |
| `database/sql`      | Database access                     |
| sqlc                | Type-safe SQL query generation      |
| Goose               | Database migrations                 |
| JWT                 | Access-token authentication         |
| bcrypt              | Password hashing                    |
| WebSockets          | Real-time communication             |
| Docker              | Application and database containers |
| HTML/CSS/JavaScript | Testing frontend                    |

## Project Structure

The code is organized around application domains rather than putting all handlers, services and repositories into separate global folders.

## Architecture

The application uses a layered structure:

```text
HTTP / WebSocket
       │
       ▼
   Handlers
       │
       ▼
    Services
       │
       ▼
  Repositories
       │
       ▼
     MariaDB
```

The WebSocket hub is responsible for delivering already-persisted messages to connected users.

MariaDB remains the source of truth for conversations and messages.

### REST API

REST is used for operations such as:

* authentication
* user registration
* retrieving users
* creating conversations
* listing conversations
* retrieving message history

### WebSockets

WebSockets are used for real-time operations:

* establishing an authenticated connection
* sending messages
* receiving messages
* synchronizing messages missed while disconnected

Messages are persisted before being broadcast to connected clients.

## Authentication

Authentication uses short-lived JWT access tokens and longer-lived refresh tokens.

```text
Login
  │
  ├── Access Token
  │      └── Short-lived JWT
  │
  └── Refresh Token
         └── Stored as a SHA-256 hash
```

Refresh tokens are rotated when used.

The previous refresh token is revoked before a new refresh token is issued, preventing the same refresh token from being successfully consumed multiple times concurrently.

Access tokens contain the authenticated user's UUID as the JWT subject.

## Database

The application uses MariaDB.

Main tables:

```text
users
   │
   ├── conversation_members
   │          │
   │          └── conversations
   │
   └── messages
```

Messages contain:

* message UUID
* conversation UUID
* sender UUID
* message content
* creation timestamp

## Message Pagination

Message history uses cursor-based pagination rather than page numbers.

The cursor represents:

```text
(created_at, message_id)
```

This provides a stable ordering even when multiple messages have the same timestamp.

The database query uses the same ordering:

```sql
ORDER BY created_at DESC, id DESC
```

and the cursor is used to retrieve older messages.

WebSocket synchronization uses the same cursor concept to retrieve messages created after the last known message.

## WebSocket Protocol

WebSocket endpoint:

```text
ws://localhost:8080/api/v1/ws
```

For browser clients, the access token is currently passed as a query parameter:

```text
ws://localhost:8080/api/v1/ws?token=<access_token>
```

### Send message

```json
{
  "type": "message",
  "conversation_id": "conversation-uuid",
  "content": "Hello!"
}
```

### Synchronize missed messages

```json
{
  "type": "sync",
  "conversation_id": "conversation-uuid",
  "after": "<cursor>"
}
```

### Message event

```json
{
  "type": "message",
  "message": {
    "id": "message-uuid",
    "conversation_id": "conversation-uuid",
    "sender_id": "user-uuid",
    "content": "Hello!",
    "created_at": "2026-09-05T10:00:00Z"
  }
}
```

### Error event

```json
{
  "type": "error",
  "error": "error message"
}
```

## API Endpoints

### Authentication

```text
POST /api/v1/auth/login
POST /api/v1/auth/refresh
```

### Users

```text
POST /api/v1/users
GET  /api/v1/users
```

### Conversations

```text
POST /api/v1/conversations
GET  /api/v1/conversations
```

### Messages

```text
GET /api/v1/conversations/{conversation_id}/messages
```

### WebSocket

```text
GET /api/v1/ws
```

Protected HTTP endpoints require:

```text
Authorization: Bearer <access_token>
```

## Configuration

Create a `.env` file based on `.env.example`.

Example:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=chat
DB_PASSWORD=chat
DB_NAME=chat

JWT_SECRET=your-secret

ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=720h
```

Do not commit the `.env` file.

## Running Locally

### Requirements

* Go
* Docker
* Docker Compose
* MariaDB-compatible Docker setup

### 1. Clone the repository

```bash
git clone https://github.com/ModstDev/Chatter.git
cd Chatter
```

### 2. Configure environment variables

```bash
cp .env.example .env
```

Set the required database credentials and JWT secret.

### 3. Start MariaDB

```bash
docker compose up -d
```

### 4. Run migrations

Use the project's Makefile migration commands.

```bash
make migrate-up
```

### 5. Start the API

```bash
go run ./cmd/api
```

The API will be available at:

```text
http://localhost:8080
```

## Frontend

The project includes a small static frontend for testing the application.

From the frontend directory:

```bash
cd frontend
python3 -m http.server 3000
```

Open:

```text
http://localhost:3000
```

The frontend can be used to test:

* registration
* login
* conversation selection
* message history
* WebSocket connections
* real-time messages
* WebSocket synchronization

## Docker

The project includes Docker configuration for running the application and MariaDB.

Start the services:

```bash
docker compose up -d
```

Stop them:

```bash
docker compose down
```

To rebuild the application image:

```bash
docker compose up -d --build
```

## Development Commands

The Makefile contains common development commands.

Typical commands include:

```bash
make migrate-up
make migrate-down
make migrate-status
```

Run:

```bash
make
```

or inspect the `Makefile` for the currently available commands.

## Testing the WebSocket

A WebSocket client such as `websocat` can be used to test the backend independently of the frontend.

Example:

```bash
websocat -H="Authorization: Bearer <access_token>" \
  ws://localhost:8080/api/v1/ws
```

For browser-based clients, the token is passed through the WebSocket query parameter because the browser's native `WebSocket` API does not provide a way to set arbitrary HTTP headers during the connection.

## Message Delivery Flow

When a user sends a message:

```text
Client
  │
  │ WebSocket
  ▼
WebSocket Handler
  │
  ▼
Message Service
  │
  ├── Validate conversation membership
  │
  ├── Save message
  │
  ▼
MariaDB
  │
  ▼
WebSocket Hub
  │
  ├── User A
  ├── User B
  └── Other active connections
```

If a user is disconnected, the message remains in MariaDB.

After reconnecting, the client can synchronize messages using a cursor.

## Design Decisions

### MariaDB as the source of truth

The WebSocket layer is not used as permanent message storage.

Messages are written to the database first and then delivered to connected clients.

This means a disconnected user can retrieve missed messages later.

### Manual dependency injection

Dependencies are constructed explicitly during application startup.

There is no dependency-injection framework.

This keeps the application easy to understand while still allowing services and repositories to depend on abstractions where useful.

### Domain-oriented packages

The project groups code by domain:

```text
auth
user
conversation
message
websocket
```

instead of creating large global folders such as:

```text
handlers
services
repositories
```

This keeps related functionality together as the application grows.

### No unnecessary infrastructure

The project intentionally does not introduce Redis, Kafka, NATS, microservices, or Kubernetes.

For the current single-server application, MariaDB and the in-memory WebSocket hub are sufficient.

Additional infrastructure can be introduced if the application eventually needs multiple API/WebSocket instances or significantly higher scale.

## Purpose

This project was built as a practical backend development project to gain experience with:

* Go
* REST APIs
* WebSockets
* authentication
* SQL
* relational database design
* transactions
* migrations
* sqlc
* repository/service architecture
* dependency injection
* Docker
* concurrent programming
* real-time communication

The goal is to demonstrate how these technologies can be combined into a small but realistic backend application without unnecessary complexity.
