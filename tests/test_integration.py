import pytest
import requests
import time
import random
import string
import websocket
import json
import sqlite3
import os

# --- Constants and Configuration ---
BASE_URL = "http://localhost:8080/v1"
WEBSOCKET_URL = "ws://localhost:8080/ws"
DB_PATH = os.getenv("DB_FILE", "chat.db") # Assumes the test runs from the root directory

# --- Helper Functions ---

def random_string(length=8):
    """Generate a random string of fixed length."""
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(length))

def random_email():
    """Generate a random email address."""
    return f"testuser_{random_string()}@{random_string()}.com"

def register_user(username, email, password):
    """Register a new user and return the response."""
    return requests.post(f"{BASE_URL}/register", json={
        "username": username,
        "email": email,
        "password": password
    })

def login_user(email, password):
    """Log in a user and return the response."""
    return requests.post(f"{BASE_URL}/login", json={
        "email": email,
        "password": password
    })

def create_group(token, name):
    """Creates a new group."""
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.post(f"{BASE_URL}/groups", headers=headers, json={"name": name})
    response.raise_for_status()
    return response.json()

def add_group_member(token, group_id, user_id):
    """Adds a user to a group."""
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.post(f"{BASE_URL}/groups/{group_id}/members", headers=headers, json={"user_id": user_id})
    response.raise_for_status()
    return response.json()

def get_db_connection():
    """Establishes a connection to the SQLite database."""
    return sqlite3.connect(DB_PATH, timeout=10)

def get_message_by_content(sender_id, recipient_id, content):
    """Fetches a specific message from the DB by its content."""
    conn = get_db_connection()
    cursor = conn.cursor()
    cursor.execute(
        "SELECT id, sender_id, recipient_id, content, status FROM messages WHERE sender_id = ? AND recipient_id = ? AND content = ?",
        (sender_id, recipient_id, content)
    )
    row = cursor.fetchone()
    conn.close()
    if row:
        return {"id": row[0], "sender_id": row[1], "recipient_id": row[2], "content": row[3], "status": row[4]}
    return None

# --- Pytest Fixtures ---

@pytest.fixture(scope="module")
def users():
    """Fixture to create and register 4 users for the tests."""
    user_data = []
    for _ in range(4):
        email = random_email()
        username = f"user_{random_string()}"
        password = "password123"
        
        # Register user
        register_response = register_user(username, email, password)
        assert register_response.status_code == 201
        
        # Log in to get token and user info
        login_response = login_user(email, password)
        assert login_response.status_code == 200
        user_info = login_response.json()
        
        user_data.append({
            "email": email,
            "username": username,
            "password": password,
            "id": user_info["user"]["id"],
            "token": user_info["token"]
        })
    return user_data

# --- Test Cases ---

def test_recent_chats_endpoint(users):
    """
    Tests the GET /v1/chats endpoint for fetching and ordering recent conversations.
    """
    user_a, user_b, user_c, group_owner = users[0], users[1], users[2], users[3]

    # 1. Create a group and add User A to it
    group = create_group(group_owner['token'], "Test Integration Group")
    group_id = group['id']
    add_group_member(group_owner['token'], group_id, user_a['id'])

    # 2. User A connects via WebSocket to send messages
    ws_a = websocket.create_connection(f"{WEBSOCKET_URL}?token={user_a['token']}")
    
    try:
        # 3. Send messages in a specific order with delays to ensure distinct timestamps
        # Message 1: A -> B
        msg_to_b = {"recipient_id": user_b['id'], "type": "p2p", "content": "Hello User B"}
        ws_a.send(json.dumps(msg_to_b))
        time.sleep(0.1) 
        
        # Message 2: A -> Group
        msg_to_group = {"recipient_id": group_id, "type": "group", "content": "Hello Group"}
        ws_a.send(json.dumps(msg_to_group))
        time.sleep(0.1)

        # Message 3: A -> C
        msg_to_c_content = "Hello User C, you should be first!"
        msg_to_c = {"recipient_id": user_c['id'], "type": "p2p", "content": msg_to_c_content}
        ws_a.send(json.dumps(msg_to_c))
        time.sleep(0.1) # Give server time to process

    finally:
        ws_a.close()

    # 4. Fetch recent chats for User A
    headers = {"Authorization": f"Bearer {user_a['token']}"}
    response = requests.get(f"{BASE_URL}/chats", headers=headers)
    
    # 5. Verifications
    assert response.status_code == 200
    chats = response.json()

    assert isinstance(chats, list), "Response should be a list of chats"
    assert len(chats) == 3, f"Expected 3 conversations, but got {len(chats)}"

    # Verify the order: C, then Group, then B
    assert chats[0]['content'] == msg_to_c_content, "Latest message (to User C) should be first"
    assert chats[0]['recipient_id'] == user_c['id'] or chats[0]['sender_id'] == user_c['id']

    assert chats[1]['type'] == 'group', "Second message should be the group chat"
    assert chats[1]['recipient_id'] == group_id

    assert chats[2]['type'] == 'p2p', "Third message should be the P2P chat with User B"
    assert chats[2]['recipient_id'] == user_b['id'] or chats[2]['sender_id'] == user_b['id']


def test_offline_delivery_and_status(users):
    """
    Tests that a message sent to an offline user is queued as 'PENDING'
    and delivered with status 'DELIVERED' once the user connects.
    """
    user_a, user_b = users[0], users[1]
    
    # --- Step 1: Send message while User B is offline ---
    ws_a = websocket.create_connection(f"{WEBSOCKET_URL}?token={user_a['token']}")
    offline_message_content = f"Offline message test {random_string()}"
    try:
        msg_to_b = {
            "recipient_id": user_b['id'],
            "type": "p2p",
            "content": offline_message_content
        }
        ws_a.send(json.dumps(msg_to_b))
        time.sleep(0.2) # Allow time for DB write
    finally:
        ws_a.close()

    # --- Step 2: Verify message is 'PENDING' in the database ---
    message = get_message_by_content(user_a['id'], user_b['id'], offline_message_content)
    assert message is not None, "Message was not found in the database"
    assert message['status'] == 'pending', f"Message status should be 'pending', but was '{message['status']}'"
    message_id = message['id']

    # --- Step 3: User B connects and receives the message ---
    ws_b = websocket.create_connection(f"{WEBSOCKET_URL}?token={user_b['token']}")
    try:
        # The hub should automatically send pending messages upon connection.
        # We set a timeout to avoid hanging if no message is received.
        ws_b.settimeout(2) 
        received_msg_str = ws_b.recv()
        
        # The first message might be a "Welcome" system message. We need to handle that.
        received_msg = json.loads(received_msg_str)
        if received_msg['type'] == 'system':
             # If it's a system message, receive the next one
            received_msg_str = ws_b.recv()
            received_msg = json.loads(received_msg_str)

        assert received_msg['content'] == offline_message_content, "User B did not receive the correct offline message"
        assert received_msg['sender_id'] == user_a['id'], "Received message has incorrect sender"

    except websocket.WebSocketTimeoutException:
        pytest.fail("User B did not receive the pending message within the timeout period.")
    finally:
        ws_b.close()

    # --- Step 4: Verify message status is updated to 'DELIVERED' ---
    # Allow a moment for the server to process the status update after delivery.
    time.sleep(0.5)
    
    updated_message = get_message_by_content(user_a['id'], user_b['id'], offline_message_content)
    assert updated_message is not None, "Message could not be re-fetched from the database"
    assert updated_message['id'] == message_id
    assert updated_message['status'] == 'delivered', f"Message status should be 'delivered', but was '{updated_message['status']}'"
