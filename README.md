# 💬 ChatServer
> **High-Performance Real-Time Group Chat Backend — Built in Go**

ChatServer is a **Go-based backend** for real-time, authenticated group messaging.  
It follows a **Hexagonal Architecture (Ports & Adapters)** to ensure scalability, testability, and a clean separation of concerns.

---

## 🚀 Features

- ⚡ Real-time WebSocket messaging  
- 🧠 Clean, modular Hexagonal Architecture  
- 🔐 Secure JWT authentication  
- 💾 SQLite storage (no external DB needed)  
- 🪵 Structured logging via Uber’s `zap`  
- 🧩 Minimal dependencies, easy to extend  
- 🌍 Cross-platform setup scripts (Linux, macOS, Windows)

---

## 🧱 Project Structure

```bash
chatserver/
├── main.go
├── internal/
│   ├── api/             # HTTP handlers, middleware, routes
│   ├── auth/            # JWT management
│   ├── config/          # Config loader
│   ├── domain/          # Core business logic & entities
│   ├── ports/sqlite/    # Persistence layer
│   └── ws/              # WebSocket hub & client logic
├── pkg/logger/          # Centralized zap logger
├── tests/               # API & integration tests
├── chatserver.db        # SQLite database (auto-generated)
├── setup.sh             # Linux setup script
├── setup_mac.sh         # macOS setup script
└── setup.ps1            # Windows setup script




⚙️ Prerequisites

Before running setup, make sure you have:

🐹 Go 1.21+ installed

🧰 Git installed

✅ Port 8080 free (default)

No .env required — defaults are built-in.

🧠 Quick Start
🐧 Linux
chmod +x setup.sh
./setup.sh

🍎 macOS
chmod +x setup_mac.sh
./setup_mac.sh

🪟 Windows (PowerShell)
Set-ExecutionPolicy Bypass -Scope Process -Force
.\setup.ps1


These scripts automatically:

Check for Go & Git

Install missing dependencies

Build the binary

Run the app (migrates DB automatically)

Create Tom & Jerry as default users

🌐 Once Running

Server URL:

http://localhost:8080


Default Users:

Username	Password
tom	password
jerry	password
📡 Core API Endpoints
Method	Endpoint	Description
POST	/v1/users/register	Register a new user
POST	/v1/users/login	Log in and get JWT
GET	/v1/users	List users (requires JWT)
GET	/v1/messages/history/:recipientID	Retrieve conversation
POST	/v1/messages/p2p/:recipientID	Send direct message
POST	/v1/messages/group/:groupID	Send group message
GET	/ws	WebSocket connection endpoint
🔐 Example — Login & Auth Test
# Log in as Tom
curl -s -X POST http://localhost:8080/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username": "tom", "password": "password"}'


Response:

{"token": "eyJhbGciOi..."}


Use that token for all protected endpoints:

curl -H "Authorization: Bearer <token>" http://localhost:8080/v1/users

⚡ WebSocket Usage

Endpoint:

ws://localhost:8080/ws?token=<JWT_TOKEN>

Example using websocat
websocat "ws://localhost:8080/ws?token=<JWT_TOKEN>"


Incoming messages follow this format:

{
  "type": "group_message",
  "group_id": 1,
  "sender_id": 2,
  "content": "ChatServer is alive!",
  "timestamp": "2025-10-25T15:00:00Z"
}

🧪 Testing

Run integration tests:

pytest tests/


Or use Postman / curl directly against the running server.

🪵 Logs & Data
Resource	Path	Notes
Logs	logs/app.log	All runtime events
Database	chatserver.db	Auto-created SQLite file
🤝 Contributing

Pull requests are welcome!
For major changes, open an issue first to discuss what you’d like to add or modify.

📜 License

Licensed under the MIT License
.
Feel free to fork, use, and build upon it.

💬 Support

If you find this project helpful, please ⭐ the repo on
👉 GitHub — Emmanuel326/chatserver

Made with ❤️ in Go.
