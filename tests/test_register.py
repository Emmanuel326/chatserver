import json
import pytest
import requests
from random import randint

BASE_URL = "http://localhost:8080/v1"

def random_email():
    return f"testuser{randint(1000,9999)}@example.com"

def register_user(username, email, password):
    payload = {
        "username": username,
        "email": email,
        "password": password
    }
    return requests.post(f"{BASE_URL}/users/register", json=payload)


def test_register_success():
    """Test successful registration with valid payload"""
    email = random_email()
    resp = register_user("validuser", email, "password123")
    
    assert resp.status_code == 201
    data = resp.json()
    assert data["message"] == "User registered successfully"
    assert "token" in data

def test_register_missing_fields():
    """Test registration with missing fields"""
    payloads = [
        {"username": "useronly"},  
        {"email": "user@example.com"},  
        {"password": "password123"}  
    ]
    
    for payload in payloads:
        resp = requests.post(f"{BASE_URL}/users/register", json=payload)
        assert resp.status_code == 400
        data = resp.json()
        assert "error" in data
        assert "Invalid request format" in data["error"]

def test_register_invalid_email():
    """Test registration with invalid email format"""
    resp = register_user("user", "invalid-email", "password123")
    assert resp.status_code == 400
    data = resp.json()
    assert "error" in data

def test_register_short_username():
    """Test registration with username less than 3 characters"""
    resp = register_user("ab", random_email(), "password123")
    assert resp.status_code == 400
    data = resp.json()
    assert "error" in data

def test_register_short_password():
    """Test registration with password less than 8 characters"""
    resp = register_user("validuser", random_email(), "short")
    assert resp.status_code == 400
    data = resp.json()
    assert "error" in data

def test_register_existing_user():
    """Test registration with email that already exists"""
    email = random_email()
    resp1 = register_user("user1", email, "password123")
    assert resp1.status_code == 201
    
    resp2 = register_user("user2", email, "password123")
    assert resp2.status_code == 500
    data = resp2.json()
    print(data)
    assert "error" in data
    assert "Failed to create user" in data["error"]

if __name__ == "__main__":
    pytest.main([__file__])
