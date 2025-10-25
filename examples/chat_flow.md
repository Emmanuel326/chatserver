🧩 ChatServer Example Flow

This document provides a step-by-step breakdown of how clients interact with the ChatServer, from registration and login, to establishing WebSocket connections, and finally exchanging peer-to-peer (P2P) messages in real time.

🚀 Stage 1: Account Setup (Registration & Login)

This stage establishes the user’s identity and generates the JWT Token required for all subsequent authenticated actions.


🧍‍♂️ A. User Registration — POST /v1/users/register

| Step | Component                              | Action                                                           | Details                                                            |
| ---- | -------------------------------------- | ---------------------------------------------------------------- | ------------------------------------------------------------------ |
| 1    | **Client**                             | Sends HTTP `POST` request to `/v1/users/register`                | Body: <br>`{"email": "...", "password": "...", "username": "..."}` |
| 2    | **API Handler** (`user_handler.go`)    | Parses JSON body                                                 | Calls the `UserService`                                            |
| 3    | **Domain Service** (`user_service.go`) | Validates input (e.g., checks for unique email), hashes password | Calls the `UserRepository`                                         |
| 4    | **Repository** (`user_repository.go`)  | Persists `domain.User` (including hashed password)               | Saves into **SQLite** (`chatserver.db`)                            |
| 5    | **API Handler**                        | Returns `201 Created`                                            | Includes the new user ID                                           |



🔑 B. User Login — POST /v1/users/login

| Step | Component                              | Action                                                           | Details                                                            |
| ---- | -------------------------------------- | ---------------------------------------------------------------- | ------------------------------------------------------------------ |
| 1    | **Client**                             | Sends HTTP `POST` request to `/v1/users/register`                | Body: <br>`{"email": "...", "password": "...", "username": "..."}` |
| 2    | **API Handler** (`user_handler.go`)    | Parses JSON body                                                 | Calls the `UserService`                                            |
| 3    | **Domain Service** (`user_service.go`) | Validates input (e.g., checks for unique email), hashes password | Calls the `UserRepository`                                         |
| 4    | **Repository** (`user_repository.go`)  | Persists `domain.User` (including hashed password)               | Saves into **SQLite** (`chatserver.db`)                            |
| 5    | **API Handler**                        | Returns `201 Created`                                            | Includes the new user ID                                           |




🔸 The token is the client’s credential for all secure operations.

🌐 Stage 2: Establishing Real-Time Connection

Before sending messages, the user must connect to the real-time WebSocket infrastructure.




🧠 WebSocket Connection — GET /ws?token=<JWT>

| Step | Component                         | Action                              | Details                                           |
| ---- | --------------------------------- | ----------------------------------- | ------------------------------------------------- |
| 1    | **Client**                        | Opens WebSocket connection to `/ws` | Example: <br>`ws://localhost:8080/ws?token=<JWT>` |
| 2    | **API Handler** (`ws_handler.go`) | Extracts token from URL             | Passes it to `JWTManager` for verification        |
| 3    | **Auth Manager** (`auth/jwt.go`)  | Verifies signature & expiration     | Extracts `user_id` from payload                   |
| 4    | **API Handler**                   | Upgrades connection to WebSocket    | Creates `ws.Client` linked to the user ID         |
| 5    | **WebSocket Hub** (`ws/hub.go`)   | Registers the new client            | Stores it in Hub’s map of active clients          |
| 6    | **Client**                        | Connection established              | Begins listening for real-time messages           |





💬 Stage 3: Sending a P2P Message (With Media)

Tom (User ID 1) sends a message with a media URL to Jerry (User ID 2).

📤 P2P Send — POST /v1/messages/p2p/:recipientID

| Step | Component                                  | Action                                                     | Details                                                                                                 |
| ---- | ------------------------------------------ | ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| 1    | **Client (Tom)**                           | Sends HTTP `POST /v1/messages/p2p/2`                       | Header: `Authorization: Bearer <JWT>` <br>Body: `{"content": "Hey Jerry!", "media_url": "https://..."}` |
| 2    | **Auth Middleware** (`middleware/auth.go`) | Validates JWT, extracts `user_id`                          | Adds it to Gin Context                                                                                  |
| 3    | **API Handler** (`message_handler.go`)     | Fetches `sender_id` from Context & `recipient_id` from URL | Calls `MessageService`                                                                                  |
| 4    | **Domain Service** (`message_service.go`)  | Constructs `domain.Message` object                         | Calls `MessageRepository`                                                                               |
| 5    | **Repository** (`message_repository.go`)   | Persists message data                                      | Includes: `sender_id`, `recipient_id`, `media_url`                                                      |
| 6    | **Domain Service**                         | Calls `ws.Hub.BroadcastP2PMessage`                         | Tells Hub to deliver message                                                                            |
| 7    | **WebSocket Hub** (`ws/hub.go`)            | Finds connected clients for Tom & Jerry                    | Sends JSON message via WebSocket                                                                        |
| 8    | **Clients (Tom & Jerry)**                  | Both receive the message instantly                         | Media message delivery complete ✅                                                                       |


