import pytest
import requests
import websocket
import json
import time
import os
import sqlite3

# Import common helpers from existing test files
from tests.test_integration import BASE_URL, WEBSOCKET_URL, DB_PATH, random_email, random_string, register_user, login_user, get_db_connection, create_group, add_group_member
# NOTE: get_jwt_token_and_user and send_p2p_message from test_offline_queueing.py are used as a reference.
# For pagination, a new local helper will be defined to support limit and before_id.

# Helper for WebSocket connection
def ws_connect(token):
    ws_url = f"{WEBSOCKET_URL}?token={token}"
    ws = websocket.create_connection(ws_url)
    # Read the initial "Welcome" message from the system
    initial_msg = json.loads(ws.recv())
    assert initial_msg["type"] == "system"
    return ws

# Helper to read messages from websocket
def ws_recv_messages(ws, count=1, timeout=5):
    messages = []
    ws.settimeout(timeout)
    for _ in range(count):
        try:
            msg = ws.recv()
            messages.append(json.loads(msg))
        except websocket.WebSocketTimeoutException:
            break
        except Exception as e:
            print(f"Error receiving WS message: {e}")
            break
    return messages

# Helper to login and get JWT token and user ID
def get_jwt_token_and_user_details(email, password):
    login_data = {"email": email, "password": password}
    resp = requests.post(f"{BASE_URL}/users/login", json=login_data)
    resp.raise_for_status()
    token = resp.json()["token"]
    
    # Decode JWT to get user ID for test purposes
    # For a real application, you'd have a JWT parsing library or an /me endpoint
    # Simplified approach for testing:
    header, payload, signature = token.split('.')
    user_id = json.loads(base64.b64decode(payload + '==').decode('utf-8'))['user_id']
    return token, user_id


import base64 # Required for decoding JWT

# Helper to send P2P messages
def send_p2p_message_helper(sender_token, recipient_id, content):
    headers = {"Authorization": f"Bearer {sender_token}"}
    payload = {"content": content}
    resp = requests.post(f"{BASE_URL}/messages/p2p/{recipient_id}", headers=headers, json=payload)
    resp.raise_for_status()
    return resp.json()

# Helper to get conversation history with pagination
def get_conversation_history_paginated(token, recipient_id, limit=50, before_id=0):
    headers = {"Authorization": f"Bearer {token}"}
    params = {"limit": limit}
    if before_id > 0:
        params["before_id"] = before_id
    
    resp = requests.get(f"{BASE_URL}/messages/history/{recipient_id}", headers=headers, params=params)
    resp.raise_for_status()
    return resp.json()["messages"]


# Fixture to set up multiple users and a group for various integration tests
@pytest.fixture(scope="module")
def setup_integration_users():
    # Ensure database is clean or recreate for module-scoped tests
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)
    # Start server might be needed here or assume it's run by a parent script

    user1_email = random_email()
    user1_username = random_string()
    register_user(user1_username, user1_email, "password123")
    user1_token, user1_id = get_jwt_token_and_user_details(user1_email, "password123")

    user2_email = random_email()
    user2_username = random_string()
    register_user(user2_username, user2_email, "password123")
    user2_token, user2_id = get_jwt_token_and_user_details(user2_email, "password123")

    user3_email = random_email()
    user3_username = random_string()
    register_user(user3_username, user3_email, "password123")
    user3_token, user3_id = get_jwt_token_and_user_details(user3_email, "password123")

    # Group setup
    group_name = random_string(10)
    group = create_group(user1_token, group_name) # user1 is owner
    group_id = group["group_id"]

    add_group_member(user1_token, group_id, user2_id)
    add_group_member(user1_token, group_id, user3_id)

    users_data = {
        "user1": {"id": user1_id, "token": user1_token, "email": user1_email, "username": user1_username},
        "user2": {"id": user2_id, "token": user2_token, "email": user2_email, "username": user2_username},
        "user3": {"id": user3_id, "token": user3_token, "email": user3_email, "username": user3_username},
        "group1": {"id": group_id, "name": group_name}
    }
    yield users_data

    # Cleanup: remove the DB file after tests if it was created for this module run.
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)


class TestChatFeatures:

    def test_recent_chats_ordering(self, setup_integration_users):
        """
        Verify that the GET /v1/chats API endpoint returns the latest message for multiple
        conversations, correctly ordered by the most recent timestamp.
        """
        u1 = setup_integration_users["user1"]
        u2 = setup_integration_users["user2"]
        u3 = setup_integration_users["user3"]
        group = setup_integration_users["group1"]

        # User1 sends message to User2
        send_p2p_message_helper(u1["token"], u2["id"], "Hey User2 from User1!")
        time.sleep(0.1) # Ensure timestamps are distinct

        # User3 sends message to User1 (more recent than u1->u2)
        send_p2p_message_helper(u3["token"], u1["id"], "Hi User1 from User3!")
        time.sleep(0.1)

        # User1 sends message to Group (most recent)
        requests.post(
            f"{BASE_URL}/messages/group/{group['id']}",
            headers={"Authorization": f"Bearer {u1['token']}"},
            json={"content": "Group message from User1!"}
        ).raise_for_status()
        time.sleep(0.1)

        # Get recent chats for User1
        resp = requests.get(
            f"{BASE_URL}/chats",
            headers={"Authorization": f"Bearer {u1['token']}"}
        )
        assert resp.status_code == 200
        recent_chats = resp.json()

        # User1 has 2 recent chats: one with U3 (as recipient), one with Group (as sender).
        # Ordered by newest first.
        assert len(recent_chats) == 2

        # Verify ordering and content
        assert recent_chats[0]["content"] == "Group message from User1!"
        assert recent_chats[0]["recipient_id"] == group["id"] # Should be group ID as recipient for group message

        assert recent_chats[1]["content"] == "Hi User1 from User3!"
        assert recent_chats[1]["sender_id"] == u3["id"] # Should be U3's ID as sender


    def test_offline_delivery_and_status(self, setup_integration_users):
        """
        Verify the full cycle: a message sent to an offline user is marked 'PENDING',
        and upon reconnection, the message is delivered, and its status is updated to 'DELIVERED' in the database.
        """
        u1 = setup_integration_users["user1"]
        u2 = setup_integration_users["user2"]

        # Ensure User2 is offline (no WS connection yet)

        # User1 sends a message to User2 (who is offline)
        sent_message_content = f"Message to offline User2 from U1 - {random_string()}"
        send_p2p_message_helper(u1["token"], u2["id"], sent_message_content)

        time.sleep(1) # Give server time to process and save

        # 1. Verify message status in DB is 'PENDING'
        conn = get_db_connection()
        cursor = conn.cursor()
        cursor.execute(
            "SELECT status FROM messages WHERE sender_id = ? AND recipient_id = ? AND content = ?",
            (u1["id"], u2["id"], sent_message_content)
        )
        status_db = cursor.fetchone()
        conn.close()

        assert status_db is not None
        assert status_db[0] == "PENDING"

        # 2. User2 connects to WebSocket
        ws_u2 = ws_connect(u2["token"])

        # Expect the pending message to be delivered
        received_messages = ws_recv_messages(ws_u2, count=1, timeout=5)
        ws_u2.close()

        assert len(received_messages) == 1
        assert received_messages[0]["content"] == sent_message_content
        assert received_messages[0]["sender_id"] == u1["id"]
        assert received_messages[0]["recipient_id"] == u2["id"]

        time.sleep(1) # Give server time to update status in DB

        # 3. Verify message status in DB is 'DELIVERED'
        conn = get_db_connection()
        cursor = conn.cursor()
        cursor.execute(
            "SELECT status FROM messages WHERE sender_id = ? AND recipient_id = ? AND content = ?",
            (u1["id"], u2["id"], sent_message_content)
        )
        status_db_after_delivery = cursor.fetchone()
        conn.close()

        assert status_db_after_delivery is not None
        assert status_db_after_delivery[0] == "DELIVERED"


    def test_p2p_pagination(self, setup_integration_users):
        """
        Verify that calling the history endpoint with 'limit' and 'before_id'
        correctly fetches the next older page of messages.
        """
        u1 = setup_integration_users["user1"]
        u2 = setup_integration_users["user2"]

        # Send more messages than the default limit (e.g., 50), let's send 30 for testing
        num_messages = 30
        sent_messages_content = []
        for i in range(num_messages):
            content = f"Pagination Test Msg {i}"
            send_p2p_message_helper(u1["token"], u2["id"], content)
            sent_messages_content.append(content)
            time.sleep(0.01) # Small delay to ensure distinct timestamps/IDs

        # The messages are retrieved in chronological order (oldest first)
        # So sent_messages_content[0] is the oldest, sent_messages_content[29] is the newest.

        # Get the first page with a limit of 10
        limit = 10
        history_page1 = get_conversation_history_paginated(u1["token"], u2["id"], limit=limit)
        assert len(history_page1) == limit
        
        # history_page1 should contain messages 20-29 (newest 10 messages)
        # history_page1[0] should be content at index num_messages - limit
        # history_page1[9] should be content at index num_messages - 1
        for i in range(limit):
            expected_content = sent_messages_content[num_messages - limit + i]
            assert history_page1[i]["content"] == expected_content
        
        first_page_oldest_id = history_page1[0]["id"] # This is the ID of message "Pagination Test Msg 20"

        # Get the second page using before_id (messages older than the oldest on page 1)
        history_page2 = get_conversation_history_paginated(u1["token"], u2["id"], limit=limit, before_id=first_page_oldest_id)
        assert len(history_page2) == limit
        
        # history_page2 should contain messages 10-19
        for i in range(limit):
            expected_content = sent_messages_content[num_messages - (2 * limit) + i]
            assert history_page2[i]["content"] == expected_content

        second_page_oldest_id = history_page2[0]["id"] # This is the ID of message "Pagination Test Msg 10"

        # Get the third page
        history_page3 = get_conversation_history_paginated(u1["token"], u2["id"], limit=limit, before_id=second_page_oldest_id)
        assert len(history_page3) == limit
        
        # history_page3 should contain messages 0-9
        for i in range(limit):
            expected_content = sent_messages_content[i]
            assert history_page3[i]["content"] == expected_content

        # Try to get a fourth page, should be empty
        history_page4 = get_conversation_history_paginated(u1["token"], u2["id"], limit=limit, before_id=history_page3[0]["id"])
        assert len(history_page4) == 0
