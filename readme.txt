# 💬 ChatServer
> **High-Performance Real-Time Group Chat Backend (Built in Go)**

ChatServer is a **Go-based backend** for real-time, authenticated group messaging.  
It follows a **Hexagonal Architecture (Ports & Adapters)** to ensure scalability, testability, and clean separation of concerns.

---

## 🧱 System Architecture

### 🧭 Overview
The system strictly separates **core business logic (Domain)** from **external dependencies** such as the database, API frameworks, and WebSockets.

### 🧩 Core Components

| Component | Technology | Role |
|------------|-------------|------|
| **Language** | Go (Golang) | Primary application language |
| **Framework** | Gin | REST API routing and middleware |
| **Database** | SQLite | Persistent, file-based storage (`chatserver.db`) |
| **Real-Time** | WebSockets | Dedicated `ws.Hub` for broadcasting messages |
| **Auth** | JWT | Token-based authentication for APIs and WebSockets |

---

## 🔄 Data Flow — Message Broadcast

1. **API Handler** (`/v1/messages/group/:id`)  
   Receives an authenticated message via HTTP `POST`.

2. **Domain Service** (`MessageService`)  
   - Checks group membership  
   - Saves message via repository  
   - Forwards it to WebSocket Hub

3. **WebSocket Hub** (`ws.Hub`)  
   Identifies all active clients in the group and broadcasts the message in real-time.

---

## ⚙️ Setup & Running

### 🧰 Prerequisites
- Go (v1.20+)
- Git
- [`websocat`](https://github.com/vi/websocat) (for WebSocket testing)
- [`jq`](https://stedolan.github.io/jq/) (for parsing tokens)

---

### 🏗️ Installation & Build

```bash
# Clone the repository
git clone https://github.com/Emmanuel326/chatserver.git
cd chatserver

# Build the application binary
go build -o chatserver_app .



Running the Server

Running the app automatically performs DB migration and creates two default users:

| Username | Email                                 | Password |
| -------- | ------------------------------------- | -------- |
| Ava      | [ava@temp.com](mailto:ava@temp.com)   | password |
| Mike     | [mike@temp.com](mailto:mike@temp.com) | password |


# Start the server
./chatserver_app


Server log:

🚀 Server running on http://localhost:8080

🧩 API Reference

All endpoints require a JWT token via the header:
Authorization: Bearer <token>
(except registration and login).

🔐 A. Authentication & Users

| Method | Endpoint             | Description                     |
| ------ | -------------------- | ------------------------------- |
| `POST` | `/v1/users/register` | Creates a new user              |
| `POST` | `/v1/users/login`    | Authenticates user, returns JWT |


Example — Get Ava’s Token
AVA_TOKEN=$(curl -s -X POST http://localhost:8080/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email": "ava@temp.com", "password": "password"}' | jq -r '.token')

echo "Ava's Token: $AVA_TOKEN"


💬 B. Group Messaging
| Method | Endpoint                      | Auth | Description                    |
| ------ | ----------------------------- | ---- | ------------------------------ |
| `POST` | `/v1/messages/group/:groupID` | ✅    | Send & save group message      |
| `POST` | `/v1/groups/:groupID/members` | ✅    | Add user to group (owner-only) |


Example — Send Authenticated Message
GROUP_ID=1
MESSAGE_CONTENT='{"content": "This is the final test message."}'

curl -X POST http://localhost:8080/v1/messages/group/$GROUP_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AVA_TOKEN" \
  -d "$MESSAGE_CONTENT"



🌐 Real-Time WebSockets
🔗 Endpoint
ws://localhost:8080/ws?token=<JWT_TOKEN>

⚡ Connection Example
# Connect Mike to WebSocket feed
websocat "ws://localhost:8080/ws?token=$MIKE_TOKEN"

📦 Message Format
{
  "type": "group_message",
  "group_id": 1,
  "sender_id": 3,
  "content": "GROUP CHAT IS ALIVE! The journey is complete.",
  "timestamp": "2025-10-24T05:23:00.000000Z"
}


📚 Full API Reference
🧾 5.1 Authentication & Public Endpoints

| Method | Endpoint             | Description                            | Request Body                                    | Success Response               |
| ------ | -------------------- | -------------------------------------- | ----------------------------------------------- | ------------------------------ |
| `POST` | `/v1/users/register` | Create user                            | `{"username": "", "email": "", "password": ""}` | `{"id": 1, "username": "..."}` |
| `POST` | `/v1/users/login`    | Authenticate user                      | `{"email": "", "password": ""}`                 | `{"token": "eyJhbGciOi..."}`   |
| `GET`  | `/ws`                | WebSocket connection (JWT query param) | —                                               | Establishes WebSocket stream   |



🔒 5.2 Protected Endpoints

| Method | Endpoint        | Description                   | Path Params | Success Response                              |
| ------ | --------------- | ----------------------------- | ----------- | --------------------------------------------- |
| `GET`  | `/v1/users`     | List all users                | —           | `[{"id": 1, "username": "..."}, ...]`         |
| `GET`  | `/v1/test-auth` | Validate JWT & return user ID | —           | `{"message": "Access granted", "user_id": 3}` |


👥 B. Group Management
| Method | Endpoint                      | Description       | Path Params | Request Body                     |
| ------ | ----------------------------- | ----------------- | ----------- | -------------------------------- |
| `POST` | `/v1/groups`                  | Create new group  | —           | `{"name": "General Chat"}`       |
| `POST` | `/v1/groups/:groupID/members` | Add user to group | `:groupID`  | `{"user_id": 4}`                 |
| `GET`  | `/v1/groups/:groupID/members` | Get group members | `:groupID`  | `[{"id": 3, "username": "..."}]` |


🧠 Design Philosophy

Clear separation of concerns (Domain vs Infrastructure)

Fully testable and modular

Built for real-time performance

Minimal dependencies, maximum maintainability

🤝 Contributing

Pull requests are welcome!
For major changes, open an issue first to discuss what you’d like to modify.

---

## 🧭 System Architecture(section 2)

### Overview
ChatServer separates **Domain Logic** (core business rules) from **Infrastructure** (frameworks, databases, APIs).



         ┌────────────────────────┐
         │   HTTP / WebSocket API │  ← Gin, JWT, WebSockets
         └────────────┬───────────┘
                      │
            ┌─────────▼─────────┐
            │   Application     │  ← Services, Handlers
            └─────────┬─────────┘
                      │
         ┌────────────▼────────────┐
         │        Domain           │  ← Core entities & logic
         └────────────┬────────────┘
                      │
       ┌──────────────▼──────────────┐
       │      Infrastructure         │  ← SQLite, Repositories
       └─────────────────────────────┘






---

## 🧩 Core Components

| Component | Technology | Role |
|------------|-------------|------|
| **Language** | Go (Golang) | Primary application language |
| **Framework** | Gin | REST API routing and middleware |
| **Database** | SQLite | Persistent, file-based storage (`chatserver.db`) |
| **Real-Time** | WebSockets | Dedicated `ws.Hub` for broadcasting messages |
| **Auth** | JWT | Token-based authentication for APIs and WebSockets |

---





Made with ❤️ in Go.

