package mail

import (
	"context"
	"sync"
	"time"
)

type Message struct {
	Kind      string
	To        string
	Secret    string
	ExpiresIn time.Duration
}

type MemoryMailer struct {
	mu       sync.Mutex
	Messages []Message
}

func (m *MemoryMailer) SendRegistrationCode(_ context.Context, to, code string, expiresIn time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, Message{Kind: "register_code", To: to, Secret: code, ExpiresIn: expiresIn})
	return nil
}

func (m *MemoryMailer) SendPasswordResetProof(_ context.Context, to, proof string, expiresIn time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, Message{Kind: "password_reset_proof", To: to, Secret: proof, ExpiresIn: expiresIn})
	return nil
}

func (m *MemoryMailer) SendPasswordResetNotice(_ context.Context, to string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, Message{Kind: "password_reset_notice", To: to})
	return nil
}
