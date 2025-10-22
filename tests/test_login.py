import json
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


def test_login_success():
    """Test login with valid credentials"""
    email = random_email()
    password = "password123"
    resp_reg = register_user("validuser", email, password)
    assert resp_reg.status_code == 201

    resp_login = login_user(email, password)
    assert resp_login.status_code == 200
    data = resp_login.json()
    assert data["message"] == "Login successful"
    assert "token" in data

def test_login_missing_fields():
    """Test login with missing fields"""
    payloads = [
        {"email": "user@example.com"},  
        {"password": "password123"},    
        {}                          
    ]
    for payload in payloads:
        resp = requests.post(f"{BASE_URL}/users/login", json=payload)
        assert resp.status_code == 400
        data = resp.json()
        assert "error" in data
        assert "Invalid request format" in data["error"]

def test_login_invalid_email_format():
    """Test login with invalid email format"""
    resp = login_user("invalid-email", "password123")
    assert resp.status_code == 400
    data = resp.json()
    assert "error" in data

def test_login_wrong_credentials():
    """Test login with incorrect credentials"""
    email = random_email()
    password = "password123"
    resp_reg = register_user("validuser", email, password)
    assert resp_reg.status_code == 201

    resp_login = login_user(email, "wrongpass")
    assert resp_login.status_code == 401
    data = resp_login.json()
    assert "error" in data
    assert "Invalid email or password" in data["error"]

if __name__ == "__main__":
    pytest.main([__file__])
