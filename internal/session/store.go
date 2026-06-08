package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Session represents a conversation session.
type Session struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id,omitempty"`
	Messages     []Message              `json:"messages,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	LastActiveAt time.Time              `json:"last_active_at"`
}

// Message represents a single message in a session.
type Message struct {
	Role      string    `json:"role"`      // user, assistant, tool
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Store provides session persistence.
type Store struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// StoreOption configures the session store.
type StoreOption func(*Store)

// WithSessionTTL sets the default session TTL.
func WithSessionTTL(ttl time.Duration) StoreOption {
	return func(s *Store) {
		s.defaultTTL = ttl
	}
}

// NewStore creates a new session store with Redis backend.
func NewStore(client *redis.Client, opts ...StoreOption) *Store {
	s := &Store{
		client:     client,
		defaultTTL: 24 * time.Hour, // Default 24 hour TTL
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Get retrieves a session by ID.
func (s *Store) Get(ctx context.Context, sessionID string) (*Session, error) {
	key := s.sessionKey(sessionID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Session not found
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &session, nil
}

// Create creates a new session.
func (s *Store) Create(ctx context.Context, userID string) (*Session, error) {
	now := time.Now()
	session := &Session{
		ID:           generateSessionID(),
		UserID:       userID,
		Messages:     make([]Message, 0),
		Context:      make(map[string]interface{}),
		CreatedAt:    now,
		LastActiveAt: now,
	}

	if err := s.Save(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// Save persists a session.
func (s *Store) Save(ctx context.Context, session *Session) error {
	session.LastActiveAt = time.Now()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	key := s.sessionKey(session.ID)
	if err := s.client.Set(ctx, key, data, s.defaultTTL).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

// AppendMessage appends a message to a session.
func (s *Store) AppendMessage(ctx context.Context, sessionID string, role, content string) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Messages = append(session.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})

	// Keep only last 50 messages to prevent memory bloat
	if len(session.Messages) > 50 {
		session.Messages = session.Messages[len(session.Messages)-50:]
	}

	return s.Save(ctx, session)
}

// SetContext sets a context value for a session.
func (s *Store) SetContext(ctx context.Context, sessionID string, key string, value interface{}) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Context[key] = value
	return s.Save(ctx, session)
}

// GetContext retrieves a context value from a session.
func (s *Store) GetContext(ctx context.Context, sessionID string, key string) (interface{}, error) {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session.Context[key], nil
}

// Delete removes a session.
func (s *Store) Delete(ctx context.Context, sessionID string) error {
	key := s.sessionKey(sessionID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

// Exists checks if a session exists.
func (s *Store) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := s.sessionKey(sessionID)
	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists: %w", err)
	}
	return count > 0, nil
}

// RefreshTTL extends the session TTL.
func (s *Store) RefreshTTL(ctx context.Context, sessionID string) error {
	key := s.sessionKey(sessionID)
	if err := s.client.Expire(ctx, key, s.defaultTTL).Err(); err != nil {
		return fmt.Errorf("redis expire: %w", err)
	}
	return nil
}

func (s *Store) sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// NewRedisClient creates a Redis client from environment variables.
func NewRedisClient(addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	slog.Info("Redis connected", "addr", addr)
	return client, nil
}
