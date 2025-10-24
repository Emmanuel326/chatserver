import json
import pytest
import requests
from websocket import create_connection
from random import randint

BASE_URL = "http://localhost:8080/v1"
WS_URL = "ws://localhost:8080/v1/ws"

def random_email():
    return f"testuser{randint(1000,9999)}@example.com"

def register_user(username, email, password):
    payload = {"username": username, "email": email, "password": password}
    return requests.post(f"{BASE_URL}/users/register", json=payload)

def login_user(email, password):
    payload = {"email": email, "password": password}
    return requests.post(f"{BASE_URL}/users/login", json=payload)

def get_jwt_token():
    """Register + Login and return JWT token"""
    email = random_email()
    password = "password123"
    resp = register_user("wsuser", email, password)
    assert resp.status_code == 201
    resp_login = login_user(email, password)
    assert resp_login.status_code == 200
    return resp_login.json()["token"]


def test_ws_connect_success():
    """Test WebSocket connection with valid token"""
    token = get_jwt_token()
    ws = create_connection(f"{WS_URL}?token={token}")
    
    ws.send(json.dumps({"action": "ping"}))
    result = ws.recv()
    
    assert result is not None
    print("Received from WS:", result)
    
    ws.close()

def test_ws_connect_missing_token():
    """Test WebSocket connection fails without token"""
    import websocket
    with pytest.raises(websocket.WebSocketBadStatusException) as exc_info:
        create_connection(f"{WS_URL}")  # No token provided
    assert "401" in str(exc_info.value)

def test_ws_connect_invalid_token():
    """Test WebSocket connection fails with invalid token"""
    import websocket
    with pytest.raises(websocket.WebSocketBadStatusException) as exc_info:
        create_connection(f"{WS_URL}?token=invalidtoken123")
    assert "401" in str(exc_info.value)

if __name__ == "__main__":
    pytest.main([__file__])
