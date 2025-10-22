import pytest
import requests
from random import randint

BASE_URL = "http://localhost:8080/v1"

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
    resp = register_user("userlist", email, password)
    assert resp.status_code == 201
    resp_login = login_user(email, password)
    assert resp_login.status_code == 200
    return resp_login.json()["token"]

def test_list_users_success():
    """Retrieve all users with valid JWT"""
    token = get_jwt_token()
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
    """Access /users without token should fail"""
    resp = requests.get(f"{BASE_URL}/users")
    assert resp.status_code == 401
    data = resp.json()
    assert "error" in data

if __name__ == "__main__":
    pytest.main([__file__])
