package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Emmanuel326/chatserver/internal/api"
	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/Emmanuel326/chatserver/internal/config"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/Emmanuel326/chatserver/internal/ports/sqlite"
	"github.com/Emmanuel326/chatserver/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // Import for SQLite driver
)

// TestApp represents the test application environment
type TestApp struct {
	HTTPServer    *httptest.Server
	WSServer      *httptest.Server
	DB            *sqlx.DB
	Config        *config.Config
	UserService   domain.UserService
	MessageService domain.MessageService
	GroupService  domain.GroupService
	JWTManager    *auth.JWTManager
	Hub           *ws.Hub
}

func setupTest(t *testing.T) *TestApp {
	// Ensure Gin is in test mode to suppress logs
	gin.SetMode(gin.TestMode)

	// Create a temporary SQLite database file
	tempDBFile := filepath.Join(t.TempDir(), "test.db")
	t.Logf("Using temporary DB file: %s", tempDBFile)

	// Load configuration - override DB_FILE for testing
	cfg := &config.Config{
		DB_FILE:     tempDBFile,
		JWT_SECRET:  "test_jwt_secret_for_integration_features",
		JWT_EXPIRY:  60, // 60 minutes
		SERVER_PORT: "8080",
	}

	// Initialize Database
	db := sqlite.InitDB(cfg)
	sqlite.Migrate(db)

	// Initialize Repositories
	userRepo := sqlite.NewUserRepository(db)
	messageRepo := sqlite.NewMessageRepository(db)
	groupRepo := sqlite.NewGroupRepository(db)

	// Initialize JWT Manager
	jwtManager := auth.NewJWTManager(cfg.JWT_SECRET, time.Duration(cfg.JWT_EXPIRY)*time.Minute)

	// Initialize Services
	userService := domain.NewUserService(userRepo)
	groupService := domain.NewGroupService(groupRepo, userRepo)
	messageService := domain.NewMessageService(messageRepo, userRepo, groupRepo)

	// Initialize WebSocket Hub
	hub := ws.NewHub(messageService, groupService, userService)
	go hub.Run() // Start the hub in a goroutine

	// Initialize API Handlers
	userHandler := api.NewUserHandler(userService, jwtManager)
	groupHandler := api.NewGroupHandler(groupService, userService, jwtManager)
	messageHandler := api.NewMessageHandler(messageService, userService, jwtManager)
	wsHandler := api.NewWSHandler(hub, jwtManager)

	// Setup Gin router
	router := gin.New() // Use gin.New() to get a clean router
	router.Use(gin.Recovery())

	// API v1 group
	v1 := router.Group("/v1")
	{
		v1.POST("/register", userHandler.Register)
		v1.POST("/login", userHandler.Login)

		// Protected routes
		protected := v1.Group("/")
		protected.Use(api.AuthMiddleware(jwtManager))
		{
			protected.GET("/users", userHandler.ListUsers)
			protected.GET("/users/with-chat-info", userHandler.ListUsersWithChatInfo)
			protected.GET("/users/:userID/messages", messageHandler.GetConversationHistory) // P2P history
			protected.GET("/chats", messageHandler.GetRecentConversations) // Recent chats endpoint

			// Group routes
			protected.POST("/groups", groupHandler.CreateGroup)
			protected.POST("/groups/:groupID/members", groupHandler.AddMember)
			protected.GET("/groups/:groupID/members", groupHandler.GetMembers)
			protected.POST("/groups/:groupID/messages", messageHandler.SendGroupMessage)
			protected.GET("/groups/:groupID/messages", messageHandler.GetGroupConversationHistory)
		}
	}

	// WebSocket route
	router.GET("/ws", wsHandler.ServeWs)

	// Start an HTTP test server for the REST API
	httpServer := httptest.NewServer(router)
	t.Cleanup(func() {
		httpServer.Close()
		hub.Stop() // Signal the hub to stop
		// Give hub a moment to process stop signal, if it has cleanup logic
		time.Sleep(100 * time.Millisecond)
		db.Close()
		os.Remove(tempDBFile)
	})

	// Start a separate HTTP test server for the WebSocket endpoint
	wsRouter := gin.New()
	wsRouter.GET("/ws", wsHandler.ServeWs)
	wsServer := httptest.NewServer(wsRouter)
	t.Cleanup(func() {
		wsServer.Close()
	})

	return &TestApp{
		HTTPServer:    httpServer,
		WSServer:      wsServer, // Using this for WS connection URL
		DB:            db,
		Config:        cfg,
		UserService:   userService,
		MessageService: messageService,
		GroupService:  groupService,
		JWTManager:    jwtManager,
		Hub:           hub,
	}
}

// --- Helper Functions for Tests ---

// User details struct for convenience
type TestUser struct {
	ID       int64
	Username string
	Email    string
	Password string
	Token    string
	WSConn   *websocket.Conn
	Messages []*ws.Message  // For collecting messages received over WebSocket
	Wg       sync.WaitGroup // To wait for message reception goroutine
	Done     chan struct{}  // To signal websocket reader goroutine to stop
	Mu       sync.Mutex     // Protects Messages
}

// registerAndLoginUser registers a new user and logs them in, returning their details and JWT.
func registerAndLoginUser(t *testing.T, app *TestApp, username, email, password string) *TestUser {
	// Register
	registerReq := api.RegisterRequest{Username: username, Email: email, Password: password}
	body, _ := json.Marshal(registerReq)
	resp, err := http.Post(app.HTTPServer.URL+"/v1/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to register user %s: %v", username, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d for registration, got %d. Body: %s", http.StatusCreated, resp.StatusCode, respBody)
	}

	// Login
	loginReq := api.LoginRequest{Email: email, Password: password}
	body, _ = json.Marshal(loginReq)
	resp, err = http.Post(app.HTTPServer.URL+"/v1/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to login user %s: %v", username, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d for login, got %d. Body: %s", http.StatusOK, resp.StatusCode, respBody)
	}

	var authResp api.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		t.Fatalf("Failed to decode auth response: %v", err)
	}

	// Get user ID (optional, but good for checks)
	claims, err := app.JWTManager.ValidateToken(authResp.Token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	return &TestUser{
		ID:       claims.UserID,
		Username: username,
		Email:    email,
		Password: password,
		Token:    authResp.Token,
		Done:     make(chan struct{}),
	}
}

// connectWS connects a user to the WebSocket and starts a reader goroutine.
func (tu *TestUser) connectWS(t *testing.T, app *TestApp) {
	wsURL := "ws" + strings.TrimPrefix(app.WSServer.URL, "http") + "/ws?token=" + tu.Token
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket for user %s: %v", tu.Username, err)
	}
	tu.WSConn = conn

	// Start a goroutine to read messages
	tu.Wg.Add(1)
	go func() {
		defer tu.Wg.Done()
		for {
			select {
			case <-tu.Done:
				log.Printf("User %s WS reader stopping", tu.Username)
				return
			default:
				// Set a read deadline to prevent blocking indefinitely
				tu.WSConn.SetReadDeadline(time.Now().Add(5 * time.Second))
				_, messageBytes, err := conn.ReadMessage()
				if err != nil {
					// Handle expected errors like "websocket: close 1000 (normal)"
					if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						t.Logf("User %s WS read error (expected): %v", tu.Username, err)
					} else if !strings.Contains(err.Error(), "i/o timeout") { // Ignore timeout errors
						t.Logf("User %s WS read error: %v", tu.Username, err)
					}
					return // Exit goroutine on any read error (including timeout after deadline)
				}
				var msg ws.Message
				if err := json.Unmarshal(messageBytes, &msg); err != nil {
					t.Logf("User %s failed to unmarshal WS message: %v, raw: %s", tu.Username, err, messageBytes)
					continue
				}
				tu.Mu.Lock()
				tu.Messages = append(tu.Messages, &msg)
				tu.Mu.Unlock()
				t.Logf("User %s received WS message: %s", tu.Username, msg.Content)
			}
		}
	}()
}

// closeWS closes the WebSocket connection for a user and waits for the reader goroutine to finish.
func (tu *TestUser) closeWS(t *testing.T) {
	if tu.WSConn != nil {
		// Signal reader to stop
		close(tu.Done)
		// Cleanly close the connection by sending a close message
		err := tu.WSConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			t.Logf("Write close error for %s: %v", tu.Username, err)
		}
		tu.WSConn.Close()
		tu.Wg.Wait() // Wait for reader goroutine to finish
		tu.WSConn = nil
		tu.Done = make(chan struct{}) // Reset for potential reuse
		tu.Messages = []*ws.Message{} // Clear messages
		t.Logf("User %s WS connection closed and reader stopped", tu.Username)
	}
}

// sendP2PMessage sends a P2P message via HTTP.
func sendP2PMessage(t *testing.T, app *TestApp, senderToken string, recipientID int64, content string) domain.Message {
	url := fmt.Sprintf("%s/v1/users/%d/messages", app.HTTPServer.URL, recipientID)
	reqBody := api.SendMessageRequest{Content: content, Type: domain.MessageTypeText}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+senderToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send P2P message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d for send P2P message, got %d. Body: %s", http.StatusCreated, resp.StatusCode, respBody)
	}

	var msg domain.Message
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatalf("Failed to decode message response: %v", err)
	}
	t.Logf("Message sent: %s (ID: %d)", msg.Content, msg.ID)
	return msg
}

// getRecentConversations fetches recent conversations for a user.
func getRecentConversations(t *testing.T, app *TestApp, userToken string) []api.UserCardResponse {
	url := app.HTTPServer.URL + "/v1/chats"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get recent conversations: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d for recent conversations, got %d. Body: %s", http.StatusOK, resp.StatusCode, respBody)
	}

	var chats []api.UserCardResponse
	if err := json.NewDecoder(resp.Body).Decode(&chats); err != nil {
		t.Fatalf("Failed to decode recent conversations response: %v", err)
	}
	return chats
}

// getConversationHistory fetches P2P conversation history.
func getConversationHistory(t *testing.T, app *TestApp, userToken string, recipientID int64, limit int, beforeID int64) []domain.Message {
	url := fmt.Sprintf("%s/v1/users/%d/messages?limit=%d", app.HTTPServer.URL, recipientID, limit)
	if beforeID > 0 {
		url += fmt.Sprintf("&before_id=%d", beforeID)
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get conversation history: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d for conversation history, got %d. Body: %s", http.StatusOK, resp.StatusCode, respBody)
	}

	var messages []domain.Message
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		t.Fatalf("Failed to decode conversation history response: %v", err)
	}
	return messages
}

// getMessageStatusFromDB directly queries the database for a message's status.
func getMessageStatusFromDB(t *testing.T, app *TestApp, messageID int64) domain.MessageStatus {
	var status string
	err := app.DB.Get(&status, "SELECT status FROM messages WHERE id = ?", messageID)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("Message with ID %d not found in DB", messageID)
		}
		t.Fatalf("Failed to get message status from DB for ID %d: %v", messageID, err)
	}
	return domain.MessageStatus(status)
}

func randomString(prefix string, length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return prefix + string(b)
}

func randomEmail(username string) string {
	return fmt.Sprintf("%s@example.com", username)
}


// --- Actual Tests ---

func TestRecentChatsOrdering(t *testing.T) {
	app := setupTest(t)

	// Register users
	alice := registerAndLoginUser(t, app, randomString("alice", 5), randomEmail("alice"), "password123")
	bob := registerAndLoginUser(t, app, randomString("bob", 5), randomEmail("bob"), "password123")
	carol := registerAndLoginUser(t, app, randomString("carol", 5), randomEmail("carol"), "password123")
	dave := registerAndLoginUser(t, app, randomString("dave", 5), randomEmail("dave"), "password123")

	// Alice sends messages to Bob and Carol, interleaved.
	// Message to Bob (oldest overall last message from Alice's perspective)
	sendP2PMessage(t, app, alice.Token, bob.ID, "Hey Bob 1 (oldest)")
	time.Sleep(10 * time.Millisecond) // Ensure distinct timestamps

	// Message to Carol (older)
	sendP2PMessage(t, app, alice.Token, carol.ID, "Hi Carol 1 (older)")
	time.Sleep(10 * time.Millisecond)

	// Message to Bob (newer for Bob)
	sendP2PMessage(t, app, alice.Token, bob.ID, "Hey Bob 2 (newer)")
	time.Sleep(10 * time.Millisecond)

	// Message to Dave (will be second most recent overall)
	sendP2PMessage(t, app, alice.Token, dave.ID, "Hello Dave 1 (latest for Dave)")
	time.Sleep(10 * time.Millisecond)

	// Message to Carol (will be third most recent overall after the next message)
	sendP2PMessage(t, app, alice.Token, carol.ID, "Hi Carol 2 (latest for Carol)")
	time.Sleep(10 * time.Millisecond)

	// Message to Bob (most recent conversation overall for Alice)
	sendP2PMessage(t, app, alice.Token, bob.ID, "Hey Bob 3 (latest for Bob)")


	// Get recent conversations for Alice
	chats := getRecentConversations(t, app, alice.Token)

	// Expected order based on the LAST message's timestamp in each conversation:
	// 1. Bob ("Hey Bob 3 (latest for Bob)")
	// 2. Carol ("Hi Carol 2 (latest for Carol)")
	// 3. Dave ("Hello Dave 1 (latest for Dave)")

	if len(chats) != 3 {
		t.Fatalf("Expected 3 recent chats, got %d", len(chats))
	}

	// Verify order and content
	// Bob should be first (most recent last message overall)
	if chats[0].Username != bob.Username || *chats[0].LastMessageContent != "Hey Bob 3 (latest for Bob)" {
		t.Errorf("Expected first chat to be Bob with 'Hey Bob 3 (latest for Bob)', got %s with '%s'", chats[0].Username, *chats[0].LastMessageContent)
	}

	// Carol should be second
	if chats[1].Username != carol.Username || *chats[1].LastMessageContent != "Hi Carol 2 (latest for Carol)" {
		t.Errorf("Expected second chat to be Carol with 'Hi Carol 2 (latest for Carol)', got %s with '%s'", chats[1].Username, *chats[1].LastMessageContent)
	}

	// Dave should be third
	if chats[2].Username != dave.Username || *chats[2].LastMessageContent != "Hello Dave 1 (latest for Dave)" {
		t.Errorf("Expected third chat to be Dave with 'Hello Dave 1 (latest for Dave)', got %s with '%s'", chats[2].Username, *chats[2].LastMessageContent)
	}

	// Also check that all LastMessageSenderID's are Alice's ID (since Alice is fetching her own chats where she's the sender)
	for _, chat := range chats {
		if chat.LastMessageSenderID == nil || *chat.LastMessageSenderID != alice.ID {
			t.Errorf("Expected LastMessageSenderID to be %d (Alice), got %v for chat with %s", alice.ID, chat.LastMessageSenderID, chat.Username)
		}
	}
}

func TestOfflineDeliveryAndStatus(t *testing.T) {
	app := setupTest(t)

	// Register users
	alice := registerAndLoginUser(t, app, randomString("alice", 5), randomEmail("alice"), "password123")
	bob := registerAndLoginUser(t, app, randomString("bob", 5), randomEmail("bob"), "password123")

	// Alice sends a message to Bob while Bob is NOT connected to WS
	messageContent := "Hello Bob, are you there?"
	sentMsg := sendP2PMessage(t, app, alice.Token, bob.ID, messageContent)

	// Verify message status in DB is PENDING
	status := getMessageStatusFromDB(t, app, sentMsg.ID)
	if status != domain.MessageStatusPending {
		t.Fatalf("Expected message status to be PENDING, got %s", status)
	}
	t.Logf("Message %d status is PENDING as expected.", sentMsg.ID)

	// Bob connects to WebSocket
	bob.connectWS(t, app)
	defer bob.closeWS(t) // Ensure WS connection is closed at test end

	// Give some time for message delivery
	time.Sleep(500 * time.Millisecond)

	// Verify Bob received the message via WebSocket
	bob.Mu.Lock()
	if len(bob.Messages) == 0 {
		t.Fatal("Bob did not receive any messages via WebSocket")
	}
	receivedMsg := bob.Messages[0]
	bob.Mu.Unlock()

	if receivedMsg.SenderID != alice.ID || receivedMsg.Content != messageContent {
		t.Errorf("Received message mismatch. Expected sender %d, content '%s'. Got sender %d, content '%s'",
			alice.ID, messageContent, receivedMsg.SenderID, receivedMsg.Content)
	}
	t.Logf("Bob received message via WS: '%s'", receivedMsg.Content)

	// Verify message status in DB is updated to DELIVERED
	status = getMessageStatusFromDB(t, app, sentMsg.ID)
	if status != domain.MessageStatusDelivered {
		t.Fatalf("Expected message status to be DELIVERED after reconnection, got %s", status)
	}
	t.Logf("Message %d status updated to DELIVERED as expected.", sentMsg.ID)
}

func TestP2PPagination(t *testing.T) {
	app := setupTest(t)

	// Register users
	alice := registerAndLoginUser(t, app, randomString("alice", 5), randomEmail("alice"), "password123")
	bob := registerAndLoginUser(t, app, randomString("bob", 5), randomEmail("bob"), "password123")

	// Send 25 messages from Alice to Bob
	const totalMessages = 25
	sentMessages := make([]domain.Message, totalMessages)
	for i := 0; i < totalMessages; i++ {
		content := fmt.Sprintf("Message %d from Alice", i+1)
		sentMessages[i] = sendP2PMessage(t, app, alice.Token, bob.ID, content)
		time.Sleep(5 * time.Millisecond) // Ensure distinct timestamps
	}

	// Messages are returned in reverse chronological order (newest first).
	// So, the newest message is `sentMessages[totalMessages-1]`, and the oldest is `sentMessages[0]`.

	// Page 1: limit=10 (newest 10 messages)
	t.Run("Page 1", func(t *testing.T) {
		history := getConversationHistory(t, app, alice.Token, bob.ID, 10, 0)

		if len(history) != 10 {
			t.Fatalf("Expected 10 messages for page 1, got %d", len(history))
		}

		// Verify order and content (newest first)
		for i := 0; i < 10; i++ {
			expectedMsg := sentMessages[totalMessages-1-i] // Sent messages: 24, 23, ..., 15
			if history[i].ID != expectedMsg.ID || history[i].Content != expectedMsg.Content {
				t.Errorf("Page 1, message %d: Expected ID %d ('%s'), got ID %d ('%s')",
					i, expectedMsg.ID, expectedMsg.Content, history[i].ID, history[i].Content)
			}
		}

		// The before_id for the next page should be the ID of the oldest message on this page.
		// That's history[9].ID
		t.Logf("Page 1: Oldest message ID: %d", history[9].ID)
	})

	// Page 2: limit=10, before_id = ID of 10th newest message (which is history[9].ID from above)
	t.Run("Page 2", func(t *testing.T) {
		// First get the before_id from page 1's last message
		page1History := getConversationHistory(t, app, alice.Token, bob.ID, 10, 0)
		beforeID := page1History[9].ID // ID of the 10th newest message (index 9)

		history := getConversationHistory(t, app, alice.Token, bob.ID, 10, beforeID)

		if len(history) != 10 {
			t.Fatalf("Expected 10 messages for page 2, got %d", len(history))
		}

		// Verify order and content (messages 14 down to 5)
		for i := 0; i < 10; i++ {
			expectedMsg := sentMessages[totalMessages-1-10-i] // Sent messages: 14, 13, ..., 5
			if history[i].ID != expectedMsg.ID || history[i].Content != expectedMsg.Content {
				t.Errorf("Page 2, message %d: Expected ID %d ('%s'), got ID %d ('%s')",
					i, expectedMsg.ID, expectedMsg.Content, history[i].ID, history[i].Content)
			}
		}
		t.Logf("Page 2: Oldest message ID: %d", history[9].ID)
	})

	// Page 3: limit=10, before_id = ID of 20th newest message
	t.Run("Page 3", func(t *testing.T) {
		page1History := getConversationHistory(t, app, alice.Token, bob.ID, 10, 0)
		page2History := getConversationHistory(t, app, alice.Token, bob.ID, 10, page1History[9].ID)
		beforeID := page2History[9].ID // ID of the 20th newest message (index 9 of page 2)

		history := getConversationHistory(t, app, alice.Token, bob.ID, 10, beforeID)

		// Expected 5 remaining messages (0 to 4)
		if len(history) != 5 {
			t.Fatalf("Expected 5 messages for page 3, got %d", len(history))
		}

		// Verify order and content (messages 4 down to 0)
		for i := 0; i < 5; i++ {
			expectedMsg := sentMessages[totalMessages-1-20-i] // Sent messages: 4, 3, ..., 0
			if history[i].ID != expectedMsg.ID || history[i].Content != expectedMsg.Content {
				t.Errorf("Page 3, message %d: Expected ID %d ('%s'), got ID %d ('%s')",
					i, expectedMsg.ID, expectedMsg.Content, history[i].ID, history[i].Content)
			}
		}
		t.Logf("Page 3: Oldest message ID: %d", history[4].ID)
	})

	// Page 4: No more messages
	t.Run("Page 4 (empty)", func(t *testing.T) {
		page1History := getConversationHistory(t, app, alice.Token, bob.ID, 10, 0)
		page2History := getConversationHistory(t, app, alice.Token, bob.ID, 10, page1History[9].ID)
		page3History := getConversationHistory(t, app, alice.Token, bob.ID, 10, page2History[9].ID)
		beforeID := page3History[4].ID // ID of the oldest message on page 3

		history := getConversationHistory(t, app, alice.Token, bob.ID, 10, beforeID)

		if len(history) != 0 {
			t.Fatalf("Expected 0 messages for page 4, got %d", len(history))
		}
	})
}
