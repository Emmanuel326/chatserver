#!/usr/bin/env bash
set -euo pipefail

GREEN="\033[0;32m"
RED="\033[0;31m"
YELLOW="\033[1;33m"
CYAN="\033[0;36m"
RESET="\033[0m"

echo -e "${CYAN}🚀 Starting ChatServer setup for Linux...${RESET}"

# --- Step 1: Check Go installation ---
if ! command -v go &> /dev/null; then
  echo -e "${YELLOW}⚙️ Go not found. Installing...${RESET}"
  if command -v apt &> /dev/null; then
    sudo apt update && sudo apt install -y golang-go
  elif command -v dnf &> /dev/null; then
    sudo dnf install -y golang
  else
    echo -e "${RED}❌ Could not detect package manager. Please install Go manually.${RESET}"
    exit 1
  fi
else
  echo -e "${GREEN}✅ Go is installed${RESET}"
fi

# --- Step 2: Check Git ---
if ! command -v git &> /dev/null; then
  echo -e "${YELLOW}⚙️ Git not found. Installing...${RESET}"
  sudo apt install -y git || { echo -e "${RED}❌ Failed to install Git.${RESET}"; exit 1; }
else
  echo -e "${GREEN}✅ Git is installed${RESET}"
fi

# --- Step 3: Dependency Setup ---
echo -e "${CYAN}📦 Setting up Go modules...${RESET}"
go mod tidy || { echo -e "${RED}❌ go mod tidy failed.${RESET}"; exit 1; }

# --- Step 4: Build the server ---
echo -e "${CYAN}🔨 Building ChatServer...${RESET}"
go build -o chatserver main.go || { echo -e "${RED}❌ Build failed.${RESET}"; exit 1; }
echo -e "${GREEN}✅ Build successful${RESET}"

# --- Step 5: Kill existing process on port 8080 ---
if lsof -i:8080 &> /dev/null; then
  echo -e "${YELLOW}⚙️ Port 8080 in use. Killing existing process...${RESET}"
  kill -9 "$(lsof -t -i:8080)" || true
fi

# --- Step 6: Start the server ---
echo -e "${CYAN}🗃️ Running ChatServer (Ctrl+C to stop)...${RESET}"
./chatserver &

sleep 3
echo -e "${GREEN}✅ Server running at http://localhost:8080${RESET}"
echo -e "${CYAN}👥 Default users created: tom / jerry${RESET}"

