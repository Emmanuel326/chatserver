import pytest
import requests
from random import randint

BASE_URL = "http://localhost:8080/v1"
default_history_limit = 50  

def random_email():
    return f"testuser{randint(1000,9999)}@example.com"

def register_user(username, email, password):
    payload = {"username": username, "email": email, "password": password}
    return requests.post(f"{BASE_URL}/users/register", json=payload)

def login_user(email, password):
    payload = {"email": email, "password": password}
    return requests.post(f"{BASE_URL}/users/login", json=payload)

def get_jwt_token_and_user():
    """Register + Login and return JWT token and full user dict"""
    email = random_email()
    password = "password123"
    username = "testuser"

    resp = register_user(username, email, password)
    assert resp.status_code in (201, 409)

    resp_login = login_user(email, password)
    assert resp_login.status_code == 200
    token = resp_login.json()["token"]

    headers = {"Authorization": f"Bearer {token}"}
    users_resp = requests.get(f"{BASE_URL}/users", headers=headers)
    assert users_resp.status_code == 200
    users = users_resp.json()
    user = next(u for u in users if u["email"] == email)
    return token, user

def test_list_users_success():
    token, _ = get_jwt_token_and_user()
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(f"{BASE_URL}/users", headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    if data:
        user = data[0]
        assert "id" in user
        assert "username" in user
        assert "email" in user

def test_list_users_unauthorized():
    resp = requests.get(f"{BASE_URL}/users")
    assert resp.status_code == 401
    data = resp.json()
    assert "error" in data

def test_get_conversation_history_success():
    sender_token, sender_user = get_jwt_token_and_user()
    recipient_token, recipient_user = get_jwt_token_and_user()
    headers_sender = {"Authorization": f"Bearer {sender_token}"}
    recipient_id = recipient_user["id"]

    resp = requests.get(f"{BASE_URL}/messages/history/{recipient_id}", headers=headers_sender)
    assert resp.status_code == 200
    data = resp.json()
    assert "messages" in data
    assert "count" in data
    assert isinstance(data["messages"], list)

def test_get_conversation_history_missing_auth():
    resp = requests.get(f"{BASE_URL}/messages/history/1")
    assert resp.status_code == 401
    data = resp.json()
    assert "error" in data

def test_get_conversation_history_invalid_recipient():
    token, _ = get_jwt_token_and_user()
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(f"{BASE_URL}/messages/history/abc", headers=headers)
    assert resp.status_code == 400
    data = resp.json()
    assert "error" in data

def test_get_conversation_history_self():
    token, user = get_jwt_token_and_user()
    headers = {"Authorization": f"Bearer {token}"}
    user_id = user["id"]

    resp = requests.get(f"{BASE_URL}/messages/history/{user_id}", headers=headers)
    assert resp.status_code == 400
    data = resp.json()
    assert "error" in data

def test_get_conversation_history_with_limit():
    sender_token, _ = get_jwt_token_and_user()
    recipient_token, recipient_user = get_jwt_token_and_user()
    headers_sender = {"Authorization": f"Bearer {sender_token}"}
    recipient_id = recipient_user["id"]

    resp = requests.get(f"{BASE_URL}/messages/history/{recipient_id}?limit=5", headers=headers_sender)
    assert resp.status_code == 200
    data = resp.json()
    assert len(data["messages"]) <= 5
    assert "count" in data

if __name__ == "__main__":
    pytest.main([__file__])
