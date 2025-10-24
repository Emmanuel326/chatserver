ChatServer: High-Performance Real-Time Group Chat Backend

The ChatServer is a Go-based backend designed for high-performance, authenticated group messaging. It utilizes a clean architectural pattern to ensure scalability, testability, and clear separation of concerns

System Architecture
The application is built around a Hexagonal Architecture (Ports & Adapters), strictly separating the core business logic (Domain) from external dependencies (Infrastructure, Database, API Frameworks).


Component,Technology,Role
Language,Go (Golang),Primary application language.
Framework,Gin,Handles all REST API routing and middleware.
Database,SQLite,"Persistent, file-based storage (chatserver.db)."
Real-Time,WebSockets,Dedicated ws.Hub for message broadcasting.
Auth,JWT,Token-based authentication for APIs and WebSockets.


Data Flow: Message Broadcast
API Handler (/v1/messages/group/:id): Receives an authenticated message via HTTP POST.

Domain Service (MessageService): Checks group membership, saves the message via the repository, and forwards the data to the WebSocket Hub.

Real-Time Hub (ws.Hub): Identifies all active, connected clients belonging to the target group and broadcasts the message in real-time.


2.  Setup and Running
Prerequisites
Go (v1.20+)

git

websocat (for testing WebSockets)

jq (for parsing tokens)

Installation and Build


bash
# Clone the repository
git clone https://github.com/Emmanuel326/chatserver.git
cd chatserver

# Build the application binary
go build -o chatserver_app .


Running the Server
Running the application automatically performs database migration and creates two default test users: Ava (ava@temp.com/password) and Mike (mike@temp.com/password).

# Start the server
./chatserver_app

(Server will log: 🚀 Server running on http://localhost:8080)


3.  API Reference
The server exposes REST endpoints, all requiring a JWT token via the Authorization: Bearer <token> header, except for user registration and login.

A. Authentication & Users


Method,Endpoint,Description
POST,/v1/users/register,Creates a new user account.
POST,/v1/users/login,Authenticates a user and returns a JWT token.


# Get Ava's Token (Using the default test user email)
AVA_TOKEN=$(curl -s -X POST http://localhost:8080/v1/users/login \
     -H "Content-Type: application/json" \
     -d '{"email": "ava@temp.com", "password": "password"}' | jq -r '.token')
echo "Ava's Token: $AVA_TOKEN"

B. Group Messaging
These endpoints are used to manage group membership and send persistent messages.


Method,Endpoint,Auth,Description
POST,/v1/messages/group/:groupID,Yes,Sends and saves a message to a specific group. Requires sender to be a member.
POST,/v1/groups/:groupID/members,Yes,Adds a user ID to a group (requires group ownership/permission).




2. Send Message Example (Authenticated)

Assuming Ava's token is in $AVA_TOKEN and she is a member of GROUP_ID=1.

GROUP_ID=1
MESSAGE_CONTENT='{"content": "This is the final test message."}'

curl -X POST http://localhost:8080/v1/messages/group/$GROUP_ID \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $AVA_TOKEN" \
     -d "$MESSAGE_CONTENT"



     4. 🌐 Real-Time WebSockets
The WebSocket endpoint provides real-time broadcast and requires the JWT token be passed as a query parameter.

Endpoint: ws://localhost:8080/ws?token=<JWT_TOKEN>




A. Connection Example

# Connect Mike to the WebSocket feed
websocat "ws://localhost:8080/ws?token=$MIKE_TOKEN"


B. Message Format
Messages broadcast through the WebSocket hub conform to the following JSON structure:

{
    "type": "group_message",
    "group_id": 1,
    "sender_id": 3,
    "content": "GROUP CHAT IS ALIVE! The journey is complete.",
    "timestamp": "2025-10-24T05:23:00.000000Z"
}




