# Backend API and Data Structures Blueprint

This document is the definitive guide for any client-side application (Java, web, mobile) interacting with the ChatServer backend. It provides a detailed, narrative explanation of the data models and API endpoints, designed to make integration as straightforward as possible.

## 1. Core Concepts: The Data Models

Before diving into the API, it's essential to understand the building blocks of our chat system. These models represent the core entities you will be working with.

### `User`
This is the heart of the system, representing a person who can send and receive messages. When you fetch a user's details, you will receive an object with this structure. Note that the `password` is never sent to the client.

| Field Name   | Type        | Description                                       |
| :----------- | :---------- | :------------------------------------------------ |
| `id`         | `int64`     | The unique, permanent number identifying this user. |
| `username`   | `string`    | The user's public display name.                   |
| `email`      | `string`    | The user's private email, used for login.         |
| `created_at` | `time.Time` | The exact moment this user registered.            |

### `Group`
This represents a chat room where multiple users can converse. Each group has a name and is owned by the user who created it.

| Field Name   | Type        | Description                                    |
| :----------- | :---------- | :--------------------------------------------- |
| `id`         | `int64`     | The unique, permanent number identifying this group. |
| `name`       | `string`    | The public display name of the group.          |
| `owner_id`   | `int64`     | The `id` of the user who created the group.    |
| `created_at` | `time.Time` | The exact moment this group was created.       |

### `Message`
This is a single communication sent by a user. Its destination is determined by the `recipient_id`, which can be either another user's ID (for a private chat) or a group's ID (for a group chat).

| Field Name     | Type            | Description                                                     |
| :------------- | :-------------- | :-------------------------------------------------------------- |
| `id`           | `int64`         | The unique, permanent number for this message.                  |
| `sender_id`    | `int64`         | The `id` of the user who sent the message.                      |
| `recipient_id` | `int64`         | The `id` of the target `User` or `Group`.                       |
| `type`         | `MessageType`   | The kind of message (e.g., `text`, `image`).                    |
| `content`      | `string`        | The actual text content of the message.                         |
| `media_url`    | `string`        | A URL pointing to an image or file, if applicable.              |
| `timestamp`    | `time.Time`     | The exact moment the server received the message.               |
| `status`       | `MessageStatus` | The delivery status (e.g., `SENT`, `PENDING`).                  |

### `UserCardResponse`
This is a special, combined model designed specifically for building the main chat list UI. It's a `User` object enriched with details about the very last message exchanged between them and the currently logged-in user. This allows you to build a "chat card" preview without making extra API calls.

| Field Name               | Type         | Description                                                          |
| :----------------------- | :----------- | :------------------------------------------------------------------- |
| `id`                     | `int64`      | The user's unique ID.                                                |
| `username`               | `string`     | The user's display name.                                             |
| `email`                  | `string`     | The user's email.                                                    |
| `last_message_content`   | `*string`    | The text of the last message. It will be `null` if no conversation exists. |
| `last_message_timestamp` | `*time.Time` | The timestamp of the last message. `null` if no conversation exists.   |
| `last_message_sender_id` | `*int64`     | The ID of who sent the last message. `null` if no conversation exists. |

---

## 2. The Grand Tour: API Endpoint Specifications

Here is a detailed walkthrough of every available API endpoint.

### Authentication (`/v1/users`)

These endpoints are your entry point. They are public and do not require an authentication token.

#### `POST /v1/users/register`
**Purpose:** To create a new user account. On success, the server automatically logs the user in and provides a JWT, ready for immediate use.

*   **Request Body:**
    ```json
    {
        "username": "newuser",
        "email": "newuser@temp.com",
        "password": "a-very-secure-password"
    }
    ```
*   **Success Response (`201 Created`):**
    ```json
    {
        "message": "User registered successfully",
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```
*   **Error Responses:**
    *   `400 Bad Request`: One of the fields was missing or invalid.
    *   `409 Conflict`: A user with that email address already exists.

#### `POST /v1/users/login`
**Purpose:** To authenticate an existing user. On success, the server provides a JWT to be used in the `Authorization` header for all subsequent protected requests.

*   **Request Body:**
    ```json
    {
        "email": "tom@temp.com",
        "password": "password123"
    }
    ```
*   **Success Response (`200 OK`):**
    ```json
    {
        "message": "Login successful",
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```
*   **Error Responses:**
    *   `401 Unauthorized`: The email or password was incorrect.

### User Management (`/v1/users`)

These endpoints require a valid JWT in the `Authorization: Bearer <token>` header.

#### `GET /v1/users`
**Purpose:** To fetch a list of every registered user in the system. This is useful for a "new chat" or "contacts" screen where a user can search for anyone to talk to.

*   **Response (`200 OK`):** An array of `User` objects.
    ```json
    [
        { "id": 1, "username": "tom", "email": "tom@temp.com", "created_at": "..." },
        { "id": 2, "username": "jerry", "email": "jerry@temp.com", "created_at": "..." }
    ]
    ```

#### `GET /v1/users/:userID`
**Purpose:** To fetch the public profile of a single user by their ID. This is useful when you have a `user_id` (e.g., from a message object) and need to display their name.

*   **Path Parameter:** `userID` - The unique ID of the user you want to retrieve.
*   **Response (`200 OK`):** A single `User` object.
    ```json
    { "id": 2, "username": "jerry", "email": "jerry@temp.com", "created_at": "..." }
    ```
*   **Error Responses:**
    *   `404 Not Found`: No user exists with that ID.

#### `GET /v1/users/with-chat-info`
**Purpose:** The powerhouse endpoint for building your main chat list. It returns a list of all users, but each user is enriched with details of the last private message exchanged with the *authenticated* user.

*   **Response (`200 OK`):** An array of `UserCardResponse` objects.
    ```json
    [
        {
            "id": 2,
            "username": "jerry",
            "email": "jerry@temp.com",
            "last_message_content": "See you tomorrow!",
            "last_message_timestamp": "2025-11-05T10:05:00Z",
            "last_message_sender_id": 1
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

### Chat & Message Management (`/v1/chats`, `/v1/messages`)

#### `GET /v1/chats`
**Purpose:** To get the most recent message from every single conversation (both P2P and group) that the user is a part of. This is the ideal endpoint for populating the main "Chats" screen, showing a list of all ongoing conversations sorted by recent activity.

*   **Response (`200 OK`):** An array of `Message` objects.
    ```json
    [
        // The latest message from the chat with user 2
        { "id": 105, "sender_id": 1, "recipient_id": 2, "content": "Sounds good!", ... },
        // The latest message from the chat in group 5
        { "id": 203, "sender_id": 3, "recipient_id": 5, "content": "Let's start the meeting.", ... }
    ]
    ```

#### `GET /v1/messages/history/:recipientID`
**Purpose:** To fetch the full conversation history for a private (P2P) chat with another user. This is what you call when a user clicks on a chat to open it.

*   **Path Parameter:** `recipientID` - The `id` of the *other* user in the conversation.
*   **Query Parameters (for pagination):**
    *   `limit` (optional, default `50`): How many messages to retrieve.
    *   `before_id` (optional): To fetch older messages, provide the ID of the oldest message you currently have. The server will return messages created before that one.
*   **Response (`200 OK`):** An array of `Message` objects.
    ```json
    [
        { "id": 101, "sender_id": 1, "recipient_id": 2, "content": "Hello Jerry!", ... },
        { "id": 102, "sender_id": 2, "recipient_id": 1, "content": "Hi Tom!", ... }
    ]
    ```

### Group Management (`/v1/groups`)

#### `GET /v1/groups`
**Purpose:** To fetch a list of every group that the authenticated user is currently a member of.

*   **Response (`200 OK`):** An array of `Group` objects.
    ```json
    [
        { "id": 5, "name": "Project Team", "owner_id": 1, "created_at": "..." },
        { "id": 8, "name": "Weekend Plans", "owner_id": 3, "created_at": "..." }
    ]
    ```

#### `POST /v1/groups`
**Purpose:** To create a new group. The user making the request automatically becomes the owner and first member of the group.

*   **Request Body:**
    ```json
    {
        "name": "New Project Group"
    }
    ```
*   **Success Response (`201 Created`):** A detailed object about the newly created group.
    ```json
    {
        "message": "Group created successfully",
        "group_id": 12,
        "name": "New Project Group",
        "owner_id": 1
    }
    ```

#### `GET /v1/groups/:groupID/members`
**Purpose:** To get a list of all users who are members of a specific group.

*   **Path Parameter:** `groupID` - The ID of the group to inspect.
*   **Response (`200 OK`):**
    ```json
    {
        "group_id": 12,
        "members": [1, 3, 5]
    }
    ```

#### `POST /v1/groups/:groupID/members`
**Purpose:** To add another user to a group. *(Note: Currently, any member can add another user. In the future, this will be restricted to group admins).*

*   **Path Parameter:** `groupID` - The ID of the group to add someone to.
*   **Request Body:**
    ```json
    {
        "user_id": 4
    }
    ```
*   **Success Response (`200 OK`):**
    ```json
    { "message": "Member added successfully" }
    ```

#### `GET /v1/groups/:groupID/messages`
**Purpose:** To fetch the full conversation history for a group chat.

*   **Path Parameter:** `groupID` - The ID of the group.
*   **Query Parameters (for pagination):**
    *   `limit` (optional, default `50`): How many messages to retrieve.
    *   `before_id` (optional): For fetching older messages.
*   **Response (`200 OK`):** An array of `Message` objects, where `recipient_id` will be the `groupID`.
    ```json
    [
        { "id": 201, "sender_id": 1, "recipient_id": 12, "content": "Welcome!", ... }
    ]
    ```

### Real-Time Communication (`/ws`)

This is not a standard REST endpoint but the gateway to our real-time messaging system.

#### `GET /ws`
**Purpose:** To upgrade the HTTP connection to a persistent WebSocket connection. All real-time events (new messages, typing indicators) will be pushed from the server to the client over this connection.

*   **Authentication:** The JWT must be passed as a query parameter.
    *   `ws://localhost:8080/ws?token=<JWT_TOKEN>`
*   **Messages Sent from Client to Server:** To send a message, the client sends a JSON string.
    *   **P2P Message:**
        ```json
        {
          "type": "p2p_message",
          "sender_id": 1,
          "recipient_id": 2,
          "content": "This is a live message!"
        }
        ```
    *   **Group Message:**
        ```json
        {
          "type": "group_message",
          "sender_id": 1,
          "group_id": 12,
          "content": "Hello everyone in the group!"
        }
        ```
*   **Messages Received by Client from Server:** The server will push messages with the following structure.
    *   **P2P Message:**
        ```json
        {
            "id": 108,
            "sender_id": 2,
            "recipient_id": 1,
            "group_id": 0,
            "type": "text",
            "content": "I got your message!",
            "media_url": "",
            "timestamp": "2025-11-05T12:00:00Z"
        }
        ```
    *   **Group Message:**
        ```json
        {
            "id": 205,
            "sender_id": 3,
            "recipient_id": 12,
            "group_id": 12,
            "type": "text",
            "content": "Let's sync up.",
            "media_url": "",
            "timestamp": "2025-11-05T12:01:00Z"
        }
        ```
