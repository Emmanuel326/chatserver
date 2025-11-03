import pytest
import requests
import json
import time
from random import randint
from websocket import create_connection

BASE_URL = "http://localhost:8080/v1"
WS_URL = "ws://localhost:8080/v1/ws"

# --- Test Helper Functions ---

def random_email():
    """Generates a unique email for new user registration."""
    return f"testuser_{randint(10000, 99999)}@example.com"

def register_user(username, email, password):
    """Registers a new user via the API."""
    payload = {"username": username, "email": email, "password": password}
    return requests.post(f"{BASE_URL}/users/register", json=payload)

def login_user(email, password):
    """Logs in a user to get a JWT."""
    payload = {"email": email, "password": password}
    return requests.post(f"{BASE_URL}/users/login", json=payload)

def get_jwt_token_and_user():
    """Helper to register and login a new user, returning their token and user data."""
    email = random_email()
    password = "password123"
    username = f"user_{randint(1000, 9999)}"

    # Register the user
    resp_reg = register_user(username, email, password)
    assert resp_reg.status_code == 201, f"Failed to register user {username}: {resp_reg.text}"

    # Log in to get the token
    resp_login = login_user(email, password)
    assert resp_login.status_code == 200, f"Failed to log in user {username}: {resp_login.text}"
    token = resp_login.json()["token"]

    # Get user details (including ID)
    headers = {"Authorization": f"Bearer {token}"}
    users_resp = requests.get(f"{BASE_URL}/users", headers=headers)
    assert users_resp.status_code == 200
    
    current_user = next((u for u in users_resp.json() if u["email"] == email), None)
    assert current_user is not None, "Failed to find newly registered user in user list."
    
    return token, current_user

def send_p2p_message(sender_token, recipient_id, content):
    """Sends a P2P message from a sender to a recipient via the API."""
    headers = {"Authorization": f"Bearer {sender_token}"}
    payload = {"content": content}
    return requests.post(f"{BASE_URL}/messages/p2p/{recipient_id}", headers=headers, json=payload)

def get_conversation_history(token, recipient_id):
    """Retrieves the message history between the token owner and a recipient."""
    headers = {"Authorization": f"Bearer {token}"}
    return requests.get(f"{BASE_URL}/messages/history/{recipient_id}", headers=headers)

# --- Integration Tests ---

def test_pending_message_scenario():
    """
    Tests if a message sent to an offline user is correctly marked as 'PENDING'.
    """
    # 1. Setup: Create sender (User A) and offline recipient (User B)
    token_a, user_a = get_jwt_token_and_user()
    _, user_b = get_jwt_token_and_user()

    # 2. Action: User A sends a message to offline User B
    message_content = f"Hello offline user B from A! ({randint(100, 999)})"
    resp_send = send_p2p_message(token_a, user_b["id"], message_content)
    assert resp_send.status_code == 200, f"Failed to send message: {resp_send.text}"

    # 3. Verification: Check the message status in the database via the history endpoint
    resp_history = get_conversation_history(token_a, user_b["id"])
    assert resp_history.status_code == 200
    
    history = resp_history.json()
    assert history["count"] > 0, "History is empty, message was not saved."
    
    sent_message = history["messages"][0]
    assert sent_message["content"] == message_content
    assert sent_message["status"] == "PENDING", "Message status should be PENDING for an offline user."
    print(f"✅ Verified message to offline user has status: {sent_message['status']}")


def test_delivery_on_reconnect():
    """
    Tests if a pending message is delivered upon WebSocket connection and its status
    is updated to 'DELIVERED'.
    """
    # 1. Setup: Create sender (User A) and recipient (User B)
    token_a, user_a = get_jwt_token_and_user()
    token_b, user_b = get_jwt_token_and_user()
    
    # 2. Action: User A sends a message to User B while B is offline
    message_content = f"Message to be delivered on reconnect! ({randint(100, 999)})"
    send_p2p_message(token_a, user_b["id"], message_content)

    # 3. Action: User B connects to the WebSocket
    ws = create_connection(f"{WS_URL}?token={token_b}", timeout=5)
    print("User B connected to WebSocket.")

    # 4. Verification: Check for delivered message over WebSocket
    received_message = None
    try:
        # The server might send a "Welcome" message first
        for _ in range(2):
            raw_msg = ws.recv()
            msg_data = json.loads(raw_msg)
            if msg_data.get("type") == "text":
                received_message = msg_data
                break
    finally:
        ws.close()
    
    assert received_message is not None, "Did not receive the pending message on WebSocket connect."
    assert received_message["content"] == message_content
    print(f"✅ User B received pending message via WebSocket: {received_message['content']}")

    # 5. Verification: Check if message status is updated to 'DELIVERED' in the database
    # Give server a moment to process the status update
    time.sleep(0.5) 
    resp_history = get_conversation_history(token_a, user_b["id"])
    assert resp_history.status_code == 200

    history = resp_history.json()
    delivered_message = history["messages"][0]
    assert delivered_message["status"] == "DELIVERED", "Message status was not updated to DELIVERED after delivery."
    print(f"✅ Verified message status updated to: {delivered_message['status']}")


def test_recent_chats_endpoint():
    """
    Tests if the /chats endpoint returns the latest message from multiple conversations,
    ordered correctly.
    """
    # 1. Setup: Create three users: A, B, and C
    token_a, user_a = get_jwt_token_and_user()
    _, user_b = get_jwt_token_and_user()
    _, user_c = get_jwt_token_and_user()
    headers_a = {"Authorization": f"Bearer {token_a}"}

    # 2. Action: Create two separate conversations
    msg_to_b = "Hello B, this is the first conversation."
    send_p2p_message(token_a, user_b["id"], msg_to_b)
    
    # Ensure timestamps are distinct
    time.sleep(1) 
    
    msg_to_c = "Hello C, this is the newer conversation."
    send_p2p_message(token_a, user_c["id"], msg_to_c)

    # 3. Verification: Call the /chats endpoint
    resp_chats = requests.get(f"{BASE_URL}/chats", headers=headers_a)
    assert resp_chats.status_code == 200, f"Failed to get recent chats: {resp_chats.text}"
    
    chats = resp_chats.json()
    assert len(chats) == 2, "Expected to find 2 recent conversations."

    # 4. Verification: Check the order and content
    # The most recent message should appear first in the list.
    latest_chat = chats[0]
    older_chat = chats[1]

    assert latest_chat["content"] == msg_to_c, "The first chat in the list should be the most recent one."
    # Determine the other participant in the chat
    latest_chat_partner_id = latest_chat['recipient_id'] if latest_chat['sender_id'] == user_a['id'] else latest_chat['sender_id']
    assert latest_chat_partner_id == user_c["id"]

    assert older_chat["content"] == msg_to_b, "The second chat should be the older one."
    older_chat_partner_id = older_chat['recipient_id'] if older_chat['sender_id'] == user_a['id'] else older_chat['sender_id']
    assert older_chat_partner_id == user_b["id"]
    
    print("✅ /chats endpoint returned conversations in the correct order.")
