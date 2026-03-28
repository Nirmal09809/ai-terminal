package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"ai-terminal/pkg/types"
)

type Manager struct {
	storageDir  string
	sessions    map[string]*Session
	mu          sync.RWMutex
}

type Session struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Messages  []types.Message `json:"messages"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func NewManager() *Manager {
	home, _ := os.UserHomeDir()
	storageDir := filepath.Join(home, ".ai-terminal", "sessions")

	os.MkdirAll(storageDir, 0755)

	return &Manager{
		storageDir: storageDir,
		sessions:   make(map[string]*Session),
	}
}

func (m *Manager) GetOrCreate(id string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" {
		id = uuid.New().String()
	}

	if sess, ok := m.sessions[id]; ok {
		return sess, nil
	}

	sess, err := m.loadSession(id)
	if err != nil {
		sess = &Session{
			ID:        id,
			Title:     "New Session",
			Messages:  []types.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		m.sessions[id] = sess
	}

	return sess, nil
}

func (m *Manager) Create() (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	sess := &Session{
		ID:        id,
		Title:     fmt.Sprintf("Session %s", time.Now().Format("2006-01-02 15:04")),
		Messages:  []types.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.sessions[id] = sess

	if err := m.saveSession(sess); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return sess, nil
}

func (m *Manager) Get(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if sess, ok := m.sessions[id]; ok {
		return sess, nil
	}

	return m.loadSession(id)
}

func (m *Manager) List() ([]*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, err := os.ReadDir(m.storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions: %w", err)
	}

	sessions := make([]*Session, 0)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			id := strings.TrimSuffix(entry.Name(), ".json")
			sess, err := m.loadSession(id)
			if err == nil {
				sessions = append(sessions, sess)
			}
		}
	}

	return sessions, nil
}

func (m *Manager) Save(sess *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess.UpdatedAt = time.Now()
	m.sessions[sess.ID] = sess

	return m.saveSession(sess)
}

func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, id)

	path := m.getSessionPath(id)
	return os.Remove(path)
}

func (m *Manager) UpdateTitle(id, title string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[id]
	if !ok {
		sess, err := m.loadSession(id)
		if err != nil {
			return fmt.Errorf("session not found: %w", err)
		}
		m.sessions[id] = sess
	}

	sess.Title = title
	sess.UpdatedAt = time.Now()

	return m.saveSession(sess)
}

func (m *Manager) getSessionPath(id string) string {
	return filepath.Join(m.storageDir, id+".json")
}

func (m *Manager) loadSession(id string) (*Session, error) {
	path := m.getSessionPath(id)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session: %w", err)
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	m.mu.Lock()
	m.sessions[id] = &sess
	m.mu.Unlock()

	return &sess, nil
}

func (m *Manager) saveSession(sess *Session) error {
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	path := m.getSessionPath(sess.ID)
	return os.WriteFile(path, data, 0644)
}
