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
    resp = register_user("authuser", email, password)
    assert resp.status_code == 201
    resp_login = login_user(email, password)
    assert resp_login.status_code == 200
    return resp_login.json()["token"]


def test_auth_success():
    """Access /test-auth with valid token"""
    token = get_jwt_token()
    headers = {"Authorization": f"Bearer {token}"}
    
    resp = requests.get(f"{BASE_URL}/test-auth", headers=headers)
    assert resp.status_code == 200
    
    data = resp.json()
    assert data["message"] == "Access granted"
    assert "user_id" in data
    assert isinstance(data["user_id"], int)

def test_auth_missing_token():
    """Access /test-auth without token should fail"""
    resp = requests.get(f"{BASE_URL}/test-auth")
    assert resp.status_code == 401  
    data = resp.json()
    assert "error" in data

def test_auth_invalid_token():
    """Access /test-auth with invalid token should fail"""
    headers = {"Authorization": "Bearer invalidtoken123"}
    resp = requests.get(f"{BASE_URL}/test-auth", headers=headers)
    assert resp.status_code == 401
    data = resp.json()
    assert "error" in data

if __name__ == "__main__":
    pytest.main([__file__])
