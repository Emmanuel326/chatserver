# ChatServer

**High-Performance Real-Time Group Chat Backend (Go)**

ChatServer is a backend system built with Go for secure, scalable, and real-time group communication.  
It implements a clean, modular architecture ensuring clear boundaries, testability, and long-term maintainability.

---

## Features

- Real-time WebSocket communication  
- JWT-based authentication  
- Layered, clean architecture  
- SQLite persistence (no external DB)  
- Structured logging using Uber’s `zap`  
- Minimal dependencies and cross-platform compatibility  

---

## Project Structure


chatserver/
├── main.go
├── internal/
│ ├── api/ # HTTP routes, handlers, middleware
│ ├── auth/ # JWT generation and validation
│ ├── config/ # Configuration loader
│ ├── domain/ # Core business logic and entities
│ ├── ports/sqlite/ # SQLite persistence adapter
│ └── ws/ # WebSocket hub and client logic
├── pkg/logger/ # Centralized logging utilities
├── tests/ # Integration and API tests
├── chatserver.db # Auto-generated SQLite database
├── setup.sh # Linux setup script
├── setup_mac.sh # macOS setup script
└── setup.ps1 # Windows setup script



---

## Prerequisites

- Go 1.21 or higher  
- Git installed and available in PATH  
- Port `8080` free for the HTTP server  

> No `.env` required — defaults are built into the configuration.

---

## Quick Start

### Linux

```bash
chmod +x setup.sh
./setup.sh



---

macOS

chmod +x setup_mac.sh
./setup_mac.sh


Windows (PowerShell)

Set-ExecutionPolicy Bypass -Scope Process -Force
.\setup.ps1


Each script will:
Verify Go and Git installations

Download dependencies

Build the binary

Run the server

Create initial demo users: tom and jerry


Running the Server

Once setup is complete:

http://localhost:8080


Default Users
| Username | Password |
| -------- | -------- |
| tom      | password |
| jerry    | password |


Core Endpoints

| Method | Endpoint                            | Description                   |
| ------ | ----------------------------------- | ----------------------------- |
| POST   | `/v1/users/register`                | Register a new user           |
| POST   | `/v1/users/login`                   | Authenticate and get JWT      |
| GET    | `/v1/users`                         | List all users (JWT required) |
| GET    | `/v1/messages/history/:recipientID` | Retrieve conversation history |
| POST   | `/v1/messages/p2p/:recipientID`     | Send direct message           |
| POST   | `/v1/messages/group/:groupID`       | Send group message            |
| GET    | `/ws`                               | WebSocket connection endpoint |





Example Usage

curl -X POST http://localhost:8080/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username": "tom", "password": "password"}'


Response:
{
  "token": "eyJhbGciOi..."
}



Use the returned token for authorized requests:

curl -H "Authorization: Bearer <token>" http://localhost:8080/v1/users

WebSocket Connection

ws://localhost:8080/ws?token=<JWT_TOKEN>

Example using websocat:
websocat "ws://localhost:8080/ws?token=<JWT_TOKEN>"


Incoming messages follow this format:

{
  "type": "group_message",
  "group_id": 1,
  "sender_id": 2,
  "content": "Hello world",
  "MediaURL": "test_image.url"
  "timestamp": "2025-10-25T15:00:00Z"
}


Logs and Data

| Resource | Path            | Description                          |
| -------- | --------------- | ------------------------------------ |
| Logs     | `logs/app.log`  | Structured logs (info, error, debug) |
| Database | `chatserver.db` | SQLite data store (auto-created)     |



Testing
Run integration tests (example using pytest or Go test suite):

go test ./tests/...


Contributing

Contributions are welcome.
For large changes, please open an issue first to discuss your proposal.


License

Licensed under the MIT License
.

Maintainer

Author: Emmanuel
Repository: github.com/Emmanuel326/chatserver










