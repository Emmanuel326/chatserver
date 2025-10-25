#!/usr/bin/env bash
set -euo pipefail

GREEN="\033[0;32m"
RED="\033[0;31m"
YELLOW="\033[1;33m"
CYAN="\033[0;36m"
RESET="\033[0m"

echo -e "${CYAN}🚀 Starting ChatServer setup for macOS...${RESET}"

# --- Step 1: Check Homebrew ---
if ! command -v brew &> /dev/null; then
  echo -e "${RED}❌ Homebrew not found. Please install it from https://brew.sh first.${RESET}"
  exit 1
fi

# --- Step 2: Check Go installation ---
if ! command -v go &> /dev/null; then
  echo -e "${YELLOW}⚙️ Installing Go via Homebrew...${RESET}"
  brew install go || { echo -e "${RED}❌ Failed to install Go.${RESET}"; exit 1; }
else
  echo -e "${GREEN}✅ Go is installed${RESET}"
fi

# --- Step 3: Check Git ---
if ! command -v git &> /dev/null; then
  echo -e "${YELLOW}⚙️ Installing Git via Homebrew...${RESET}"
  brew install git || { echo -e "${RED}❌ Failed to install Git.${RESET}"; exit 1; }
else
  echo -e "${GREEN}✅ Git is installed${RESET}"
fi

# --- Step 4: Build project ---
echo -e "${CYAN}🔨 Building ChatServer...${RESET}"
go mod tidy
go build -o chatserver main.go || { echo -e "${RED}❌ Build failed.${RESET}"; exit 1; }

# --- Step 5: Kill existing process on port 8080 ---
if lsof -i:8080 &> /dev/null; then
  echo -e "${YELLOW}⚙️ Port 8080 in use. Killing existing process...${RESET}"
  kill -9 "$(lsof -t -i:8080)" || true
fi

# --- Step 6: Run server ---
echo -e "${CYAN}🗃️ Running ChatServer (Ctrl+C to stop)...${RESET}"
./chatserver &

sleep 3
echo -e "${GREEN}✅ Server running at http://localhost:8080${RESET}"
echo -e "${CYAN}👥 Default users created: tom / jerry${RESET}"

