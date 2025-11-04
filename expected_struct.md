# Backend API and Data Structures Blueprint

This document outlines the core data models and API endpoints of the chat server backend, serving as a single source of truth for frontend or SDK integration.

## 1. Model Definitions

### `User` (internal/domain/user.go)

Represents a registered user in the system.

| Field Name | Type      | Description                  | Required |
| :--------- | :-------- | :--------------------------- | :------- |
| `id`         | `int64`     | Unique identifier for the user | Yes      |
| `username`   | `string`    | User's chosen display name   | Yes      |
| `email`      | `string`    | User's email address (unique) | Yes      |
| `password`   | `string`    | Hashed password (backend only, never exposed) | Yes |
| `created_at` | `timestamp` | Time of user registration    | Yes      |

### `UserWithChatInfo` (internal/domain/user.go)

Internal model combining User data with latest chat message information, used for `UserCardResponse`.

| Field Name             | Type         | Description                                    | Required |
| :----------------------- | :----------- | :--------------------------------------------- | :------- |
| `ID`                     | `int64`      | Unique identifier for the user                 | Yes      |
| `Username`               | `string`     | User's chosen display name                     | Yes      |
| `Email`                  | `string`     | User's email address                           | Yes      |
| `CreatedAt`              | `timestamp`  | Time of user registration                      | Yes      |
| `LastMessageContent`     | `*string`    | Content of the last P2P message with the current user | Optional |
| `LastMessageTimestamp`   | `*timestamp` | Timestamp of the last P2P message with the current user | Optional |
| `LastMessageSenderID`    | `*int64`     | ID of the sender of the last P2P message       | Optional |

### `Message` (internal/domain/message.go)

Represents a chat message, either P2P or Group.

| Field Name    | Type             | Description                                  | Required |
| :------------ | :--------------- | :------------------------------------------- | :------- |
| `id`          | `int64`          | Unique identifier for the message            | Yes      |
| `sender_id`   | `int64`          | ID of the user who sent the message          | Yes      |
| `recipient_id` | `int64`          | ID of the recipient user or group            | Yes      |
| `type`        | `MessageType`    | Type of message (text, image, system, typing) | Yes      |
| `content`     | `string`         | Text content of the message                  | Yes      |
| `media_url`   | `string`         | URL to external media (e.g., image file)     | Optional |
| `timestamp`   | `timestamp`      | Time when the message was sent               | Yes      |
| `status`      | `MessageStatus`  | Current status of the message                | Yes      |

### `MessageType` (internal/domain/message.go)

Enum for message types.

| Value         | Description               |
| :------------ | :------------------------ |
| `text`        | Standard text message     |
| `image`       | Image message             |
| `system`      | System-generated message  |
| `typing`      | Typing indicator (transient) |

### `MessageStatus` (internal/domain/message.go)

Enum for message delivery status.

| Value       | Description                                 |
| :---------- | :------------------------------------------ |
| `SENT`      | Message sent to online user or group        |
| `PENDING`   | Message queued for offline P2P recipient    |
| `DELIVERED` | Pending message successfully delivered      |

### `Group` (internal/domain/group.go)

Represents a chat group.

| Field Name   | Type      | Description                    | Required |
| :----------- | :-------- | :----------------------------- | :------- |
| `id`         | `int64`     | Unique identifier for the group | Yes      |
| `name`       | `string`    | Name of the group              | Yes      |
| `owner_id`   | `int64`     | ID of the user who created the group | Yes      |
| `created_at` | `timestamp` | Time of group creation         | Yes      |

### `GroupMember` (internal/domain/group.go)

Represents a membership of a user in a group.

| Field Name | Type      | Description                        | Required |
| :--------- | :-------- | :--------------------------------- | :------- |
| `group_id`   | `int64`     | ID of the group                  | Yes      |
| `user_id`    | `int64`     | ID of the user who is a member   | Yes      |
| `joined_at`  | `timestamp` | Time when the user joined the group | Yes      |
| `is_admin`   | `bool`      | Whether the user is an admin of the group | Yes      |

### `RegisterRequest` (internal/api/models.go)

Request body for user registration.

| Field Name | Type     | Description             | Required |
| :--------- | :------- | :---------------------- | :------- |
| `username`   | `string`   | User's chosen username  | Yes      |
| `email`      | `string`   | User's email address    | Yes      |
| `password`   | `string`   | User's password (min 8 chars) | Yes      |

### `LoginRequest` (internal/api/models.go)

Request body for user login.

| Field Name | Type     | Description         | Required |
| :--------- | :------- | :------------------ | :------- |
| `email`      | `string`   | User's email address | Yes      |
| `password`   | `string`   | User's password     | Yes      |

### `AuthResponse` (internal/api/models.go)

Response body for successful authentication.

| Field Name | Type     | Description               | Required |
| :--------- | :------- | :------------------------ | :------- |
| `token`      | `string`   | JWT for authenticated user | Yes      |

### `SendGroupMessageRequest` (internal/api/models.go)

Request body for sending a message (both P2P and Group).

| Field Name | Type     | Description                        | Required |
| :--------- | :------- | :--------------------------------- | :------- |
| `content`    | `string`   | Text content of the message        | Conditional |
| `media_url`  | `string`   | URL to external media (e.g., S3 link) | Conditional |
| *Note:* Either `content` or `media_url` must be provided. |

### `UserCardResponse` (internal/api/models.go)

Response body for `GET /v1/users/with-chat-info`.

| Field Name             | Type         | Description                                    | Required |
| :----------------------- | :----------- | :--------------------------------------------- | :------- |
| `id`                     | `int64`      | Unique identifier for the user                 | Yes      |
| `username`               | `string`     | User's chosen display name                     | Yes      |
| `email`                  | `string`     | User's email address                           | Yes      |
| `last_message_content`   | `*string`    | Content of the last P2P message with the current user | Optional |
| `last_message_timestamp` | `*timestamp` | Timestamp of the last P2P message with the current user | Optional |
| `last_message_sender_id` | `*int64`     | ID of the sender of the last P2P message       | Optional |

### `WS.Message` (internal/ws/models.go)

WebSocket message structure for real-time communication.

| Field Name    | Type             | Description                                  | Required |
| :------------ | :--------------- | :------------------------------------------- | :------- |
| `id`          | `int64`          | Unique identifier for the message            | Yes      |
| `sender_id`   | `int64`          | ID of the user who sent the message          | Yes      |
| `recipient_id` | `int64`          | ID of the recipient user or group            | Yes      |
| `type`        | `domain.MessageType` | Type of message (text, image, system)        | Yes      |
| `content`     | `string`         | Text content of the message                  | Yes      |
| `media_url`   | `string`         | URL to external media (e.g., image file)     | Optional |
| `timestamp`   | `timestamp`      | Time when the message was sent               | Yes      |

### `WS.TypingNotification` (internal/ws/models.go)

WebSocket message structure for typing indicators.

| Field Name    | Type             | Description                     | Required |
| :------------ | :--------------- | :------------------------------ | :------- |
| `sender_id`   | `int64`          | ID of the user who is typing    | Yes      |
| `recipient_id` | `int64`          | ID of the recipient user or group | Yes      |
| `type`        | `domain.MessageType` | Must be `domain.TypingMessage`  | Yes      |

## 2. API Endpoint Specifications

### `POST /v1/users/register`

Description: Registers a new user account.

Authentication: No

Request Body JSON Schema: `RegisterRequest`
```json
{
    "username": "newuser",
    "email": "newuser@temp.com",
    "password": "securepassword123"
}
```

Response JSON Schema: `AuthResponse`
```json
{
    "message": "User registered successfully",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Status Codes:
*   `201 Created`: User registered successfully.
*   `400 Bad Request`: Invalid request format or missing fields.
*   `409 Conflict`: User with email already exists.
*   `500 Internal Server Error`: Backend error.

### `POST /v1/users/login`

Description: Authenticates a user and returns a JWT.

Authentication: No

Request Body JSON Schema: `LoginRequest`
```json
{
    "email": "user@temp.com",
    "password": "securepassword123"
}
```

Response JSON Schema: `AuthResponse`
```json
{
    "message": "Login successful",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Status Codes:
*   `200 OK`: Login successful.
*   `400 Bad Request`: Invalid request format or missing fields.
*   `401 Unauthorized`: Invalid email or password.
*   `500 Internal Server Error`: Backend error.

### `GET /v1/users`

Description: Retrieves a list of all registered users in the system.

Authentication: Yes (JWT)

Request Body JSON Schema: N/A

Response JSON Schema: `[]domain.User` (password field omitted)
```json
[
    {
        "id": 1,
        "username": "tom",
        "email": "tom@temp.com",
        "created_at": "2023-10-27T10:00:00Z"
    },
    {
        "id": 2,
        "username": "jerry",
        "email": "jerry@temp.com",
        "created_at": "2023-10-27T10:01:00Z"
    }
]
```

Status Codes:
*   `200 OK`: User list retrieved successfully.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `500 Internal Server Error`: Backend error.

### `GET /v1/users/with-chat-info`

Description: Retrieves a list of all registered users, each including the latest P2P message snippet and timestamp with the authenticated user (if a conversation exists). Ideal for building chat cards.

Authentication: Yes (JWT)

Request Body JSON Schema: N/A

Response JSON Schema: `[]api.UserCardResponse`
```json
[
    {
        "id": 1,
        "username": "tom",
        "email": "tom@temp.com",
        "last_message_content": "Are you there?",
        "last_message_timestamp": "2023-10-27T10:05:00Z",
        "last_message_sender_id": 2
    },
    {
        "id": 3,
        "username": "alice",
        "email": "alice@temp.com",
        "last_message_content": null,
        "last_message_timestamp": null,
        "last_message_sender_id": null
    }
]
```

Status Codes:
*   `200 OK`: User list with chat info retrieved successfully.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `500 Internal Server Error`: Backend error.

### `GET /v1/messages/history/:recipientID`

Description: Retrieves paginated P2P message history between the authenticated user and a specific recipient. Messages are returned in chronological order (oldest to newest).

Authentication: Yes (JWT)

Request Body JSON Schema: N/A

Pagination, Sorting, Filtering:
*   `recipientID` (Path Parameter): `int64`, ID of the other user in the conversation.
*   `limit` (Query Parameter): `int`, Maximum number of messages to return (default: 50).
*   `before_id` (Query Parameter): `int64`, Messages with an ID less than this will be returned (for fetching older messages).

Response JSON Schema:
```json
{
    "messages": [
        {
            "id": 101,
            "sender_id": 1,
            "recipient_id": 2,
            "type": "text",
            "content": "Hello Jerry!",
            "media_url": "",
            "timestamp": "2023-10-27T10:00:00Z",
            "status": "SENT"
        },
        {
            "id": 102,
            "sender_id": 2,
            "recipient_id": 1,
            "type": "text",
            "content": "Hi Tom!",
            "media_url": "",
            "timestamp": "2023-10-27T10:01:00Z",
            "status": "SENT"
        }
    ],
    "count": 2
}
```

Status Codes:
*   `200 OK`: Message history retrieved successfully.
*   `400 Bad Request`: Invalid recipient ID or query parameters.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `500 Internal Server Error`: Backend error.

### `GET /v1/chats`

Description: Retrieves the latest message from each unique P2P and Group conversation the authenticated user is part of. Represents the main "recent chats" list.

Authentication: Yes (JWT)

Request Body JSON Schema: N/A

Response JSON Schema: `[]domain.Message`
```json
[
    {
        "id": 105,
        "sender_id": 1,
        "recipient_id": 2,
        "type": "text",
        "content": "See you later!",
        "media_url": "",
        "timestamp": "2023-10-27T10:15:00Z",
        "status": "SENT"
    },
    {
        "id": 203,
        "sender_id": 1,
        "recipient_id": 100,
        "type": "text",
        "content": "Meeting at 3 PM?",
        "media_url": "",
        "timestamp": "2023-10-27T09:30:00Z",
        "status": "SENT"
    }
]
```

Status Codes:
*   `200 OK`: Recent conversations retrieved successfully.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `500 Internal Server Error`: Backend error.

### `POST /v1/groups`

Description: Creates a new chat group.

Authentication: Yes (JWT)

Request Body JSON Schema:
```json
{
    "name": "Team Alpha"
}
```

Response JSON Schema: `domain.Group`
```json
{
    "id": 100,
    "name": "Team Alpha",
    "owner_id": 1,
    "created_at": "2023-10-27T11:00:00Z"
}
```

Status Codes:
*   `201 Created`: Group created successfully.
*   `400 Bad Request`: Invalid request format or missing name.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `500 Internal Server Error`: Backend error.

### `POST /v1/groups/:groupID/members`

Description: Adds a member to an existing group. The authenticated user must be the group owner.

Authentication: Yes (JWT)

Request Body JSON Schema:
```json
{
    "user_id": 3
}
```

Response JSON Schema:
```json
{
    "message": "Member added successfully"
}
```

Status Codes:
*   `200 OK`: Member added successfully.
*   `400 Bad Request`: Invalid group ID, user ID, or request format.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `403 Forbidden`: Authenticated user is not the group owner.
*   `404 Not Found`: Group or user to be added does not exist.
*   `409 Conflict`: User is already a member of the group.
*   `500 Internal Server Error`: Backend error.

### `GET /v1/groups/:groupID/members`

Description: Retrieves a list of user IDs for all members of a specific group. The authenticated user must be a member of the group.

Authentication: Yes (JWT)

Request Body JSON Schema: N/A

Response JSON Schema:
```json
{
    "members": [1, 2, 3]
}
```

Status Codes:
*   `200 OK`: Group members retrieved successfully.
*   `400 Bad Request`: Invalid group ID.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `403 Forbidden`: Authenticated user is not a member of the group.
*   `404 Not Found`: Group not found.
*   `500 Internal Server Error`: Backend error.

### `GET /v1/groups/:groupID/messages`

Description: Retrieves paginated message history for a specific group. Messages are returned in chronological order (oldest to newest). The authenticated user must be a member of the group.

Authentication: Yes (JWT)

Request Body JSON Schema: N/A

Pagination, Sorting, Filtering:
*   `groupID` (Path Parameter): `int64`, ID of the group.
*   `limit` (Query Parameter): `int`, Maximum number of messages to return (default: 50).
*   `before_id` (Query Parameter): `int64`, Messages with an ID less than this will be returned (for fetching older messages).

Response JSON Schema:
```json
{
    "messages": [
        {
            "id": 201,
            "sender_id": 1,
            "recipient_id": 100,
            "type": "text",
            "content": "Welcome everyone!",
            "media_url": "",
            "timestamp": "2023-10-27T09:00:00Z",
            "status": "SENT"
        },
        {
            "id": 202,
            "sender_id": 2,
            "recipient_id": 100,
            "type": "text",
            "content": "Hey Tom!",
            "media_url": "",
            "timestamp": "2023-10-27T09:01:00Z",
            "status": "SENT"
        }
    ],
    "count": 2
}
```

Status Codes:
*   `200 OK`: Group message history retrieved successfully.
*   `400 Bad Request`: Invalid group ID or query parameters.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `403 Forbidden`: Authenticated user is not a member of the group.
*   `404 Not Found`: Group not found.
*   `500 Internal Server Error`: Backend error.

### `POST /v1/messages/group/:groupID`

Description: Sends a new message to a specific group. The authenticated user must be a member of the group.

Authentication: Yes (JWT)

Request Body JSON Schema: `SendGroupMessageRequest`
```json
{
    "content": "Hello team!",
    "media_url": ""
}
```
OR
```json
{
    "content": "",
    "media_url": "https://example.com/image.jpg"
}
```

Response JSON Schema:
```json
{
    "message": "Message sent successfully to group"
}
```

Status Codes:
*   `200 OK`: Message sent successfully.
*   `400 Bad Request`: Invalid group ID, request body, or empty message.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `403 Forbidden`: Sender is not a member of the group.
*   `500 Internal Server Error`: Backend error.

### `POST /v1/messages/p2p/:recipientID`

Description: Sends a new P2P message to a specific user.

Authentication: Yes (JWT)

Request Body JSON Schema: `SendGroupMessageRequest` (reused)
```json
{
    "content": "Hi there!",
    "media_url": ""
}
```
OR
```json
{
    "content": "",
    "media_url": "https://example.com/image.jpg"
}
```

Response JSON Schema:
```json
{
    "message": "P2P message sent successfully"
}
```

Status Codes:
*   `200 OK`: Message sent successfully.
*   `400 Bad Request`: Invalid recipient ID, request body, or empty message. Cannot send message to self.
*   `401 Unauthorized`: Invalid or missing JWT.
*   `404 Not Found`: Recipient user not found.
*   `500 Internal Server Error`: Backend error.

### `GET /ws`

Description: Establishes a WebSocket connection for real-time messaging and typing notifications. All P2P and Group messages involving the authenticated user are pushed over this connection. Offline messages are delivered upon connection.

Authentication: Yes (JWT via `token` query parameter)

Request Body JSON Schema: N/A (WebSocket connection upgrade)

Response (WebSocket Messages - `WS.Message` / `WS.TypingNotification`):
*   **System Message (upon connection):**
    ```json
    {
        "id": 0,
        "sender_id": 0,
        "recipient_id": 0,
        "type": "system",
        "content": "Welcome to the chat server.",
        "media_url": "",
        "timestamp": "2023-10-27T10:00:00Z"
    }
    ```
*   **Incoming Chat Message:**
    ```json
    {
        "id": 106,
        "sender_id": 1,
        "recipient_id": 2,
        "type": "text",
        "content": "Are you free for lunch?",
        "media_url": "",
        "timestamp": "2023-10-27T10:30:00Z"
    }
    ```
*   **Typing Notification:** (Transient, not persisted)
    ```json
    {
        "sender_id": 1,
        "recipient_id": 2,
        "type": "typing"
    }
    ```

Status Codes:
*   `101 Switching Protocols`: WebSocket connection successful.
*   `400 Bad Request`: Missing `token` query parameter.
*   `401 Unauthorized`: Invalid or expired JWT.

### `GET /v1/test-auth`

Description: A test endpoint to verify JWT authentication.

Authentication: Yes (JWT)

Request Body JSON Schema: N/A

Response JSON Schema:
```json
{
    "message": "Access granted",
    "user_id": 1
}
```

Status Codes:
*   `200 OK`: Access granted.
*   `401 Unauthorized`: Invalid or missing JWT.

### `GET /ping`

Description: Simple health check endpoint.

Authentication: No

Request Body JSON Schema: N/A

Response JSON Schema:
```json
{
    "message": "pong",
    "db_driver": "sqlite3"
}
```

Status Codes:
*   `200 OK`: Server is responsive.

## 3. CRUD Summary per Model

### `User`

| Operation | Endpoint                     | Description                                    | Request Body                 | Response Body                  |
| :-------- | :--------------------------- | :--------------------------------------------- | :--------------------------- | :----------------------------- |
| Create    | `POST /v1/users/register`    | Registers a new user.                          | `api.RegisterRequest`        | `api.AuthResponse`             |
| Read      | `GET /v1/users`              | Lists all users.                               | N/A                          | `[]domain.User`                |
| Read      | `GET /v1/users/with-chat-info` | Lists all users with latest message previews.  | N/A                          | `[]api.UserCardResponse`       |
| Read      | `POST /v1/users/login`       | Authenticates user, returning JWT.             | `api.LoginRequest`           | `api.AuthResponse`             |
| Update    | *Not exposed*                 | User profile updates.                          | N/A                          | N/A                            |
| Delete    | *Not exposed*                 | User account deletion.                         | N/A                          | N/A                            |

### `Message`

| Operation | Endpoint                              | Description                                    | Request Body                 | Response Body                  |
| :-------- | :------------------------------------ | :--------------------------------------------- | :--------------------------- | :----------------------------- |
| Create    | `POST /v1/messages/p2p/:recipientID`  | Sends a P2P message.                           | `api.SendGroupMessageRequest`| `{"message": "..."}`           |
| Create    | `POST /v1/messages/group/:groupID`    | Sends a group message.                         | `api.SendGroupMessageRequest`| `{"message": "..."}`           |
| Read      | `GET /v1/messages/history/:recipientID` | Retrieves paginated P2P history.               | N/A                          | `{"messages": [], "count": N}` |
| Read      | `GET /v1/groups/:groupID/messages`    | Retrieves paginated group history.             | N/A                          | `{"messages": [], "count": N}` |
| Read      | `GET /v1/chats`                       | Retrieves latest message from all conversations. | N/A                          | `[]domain.Message`             |
| Read      | `GET /ws`                             | Receives real-time messages via WebSocket.     | N/A                          | `ws.Message` or `ws.TypingNotification` |
| Update    | *Internal Only*                       | Status updates (`PENDING` -> `DELIVERED`).      | N/A                          | N/A                            |
| Delete    | *Not exposed*                         | Message deletion.                              | N/A                          | N/A                            |

### `Group`

| Operation | Endpoint                     | Description                                    | Request Body                 | Response Body                  |
| :-------- | :--------------------------- | :--------------------------------------------- | :--------------------------- | :----------------------------- |
| Create    | `POST /v1/groups`            | Creates a new group.                           | `{"name": "..."}`            | `domain.Group`                 |
| Read      | *Not exposed*                 | Retrieve a single group by ID (internal to service). | N/A                          | N/A                            |
| Update    | *Not exposed*                 | Group name updates.                            | N/A                          | N/A                            |
| Delete    | *Not exposed*                 | Group deletion.                                | N/A                          | N/A                            |

### `GroupMember`

| Operation | Endpoint                     | Description                                    | Request Body                 | Response Body                  |
| :-------- | :--------------------------- | :--------------------------------------------- | :--------------------------- | :----------------------------- |
| Create    | `POST /v1/groups/:groupID/members` | Adds a user as a member to a group.            | `{"user_id": N}`             | `{"message": "..."}`           |
| Read      | `GET /v1/groups/:groupID/members` | Lists user IDs of all members in a group.      | N/A                          | `{"members": []}`              |
| Update    | *Not exposed*                 | Change member role (e.g., admin).              | N/A                          | N/A                            |
| Delete    | *Not exposed*                 | Remove member from group.                      | N/A                          | N/A                            |
