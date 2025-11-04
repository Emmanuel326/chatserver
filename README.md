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


# 1. Clone the repository
git clone https://github.com/Emmanuel326/chatserver.git

# 2. Move into the project directory
cd chatserver
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
| Username | Email            | Password      |
| :------- | :--------------- | :------------ |
| tom      | tom@temp.com     | `password123` |
| jerry    | jerry@temp.com   | `password123` |
| alice    | alice@temp.com   | `password123` |
| bob      | bob@temp.com     | `password123` |
| charlie  | charlie@temp.com | `password123` |
| diana    | diana@temp.com   | `password123` |
| eve      | eve@temp.com     | `password123` |


Core Endpoints

A detailed API specification is available in `expected_struct.md`. Below is a summary of the most important endpoints.

| Method | Endpoint                            | Description                                       | Auth Required |
| :----- | :---------------------------------- | :------------------------------------------------ | :------------ |
| `POST` | `/v1/users/register`                | Register a new user account.                      | No            |
| `POST` | `/v1/users/login`                   | Authenticate and receive a JWT.                   | No            |
| `GET`  | `/ws`                               | Establish a real-time WebSocket connection.       | Yes (token)   |
| `GET`  | `/v1/users`                         | Get a list of all users.                          | Yes (Bearer)  |
| `GET`  | `/v1/users/:userID`                 | Get public details for a single user.             | Yes (Bearer)  |
| `GET`  | `/v1/users/with-chat-info`          | Get users with last message previews (for cards). | Yes (Bearer)  |
| `GET`  | `/v1/chats`                         | Get a list of recent conversations (P2P & Group). | Yes (Bearer)  |
| `GET`  | `/v1/groups`                        | Get a list of all groups the user is a member of. | Yes (Bearer)  |
| `POST` | `/v1/groups`                        | Create a new group.                               | Yes (Bearer)  |
| `POST` | `/v1/groups/:groupID/members`       | Add a member to a group.                          | Yes (Bearer)  |
| `GET`  | `/v1/groups/:groupID/members`       | Get a list of all members in a group.             | Yes (Bearer)  |
| `GET`  | `/v1/groups/:groupID/messages`      | Get message history for a group.                  | Yes (Bearer)  |
| `GET`  | `/v1/messages/history/:recipientID` | Get P2P message history with another user.        | Yes (Bearer)  |





Example Usage

**1. Log in to get a token:**

```bash
curl -X POST http://localhost:8080/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email": "tom@temp.com", "password": "password123"}'
```

Response:
```json
{
    "message": "Login successful",
    "token": "eyJhbGciOi..."
}
```



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










