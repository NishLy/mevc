package rtc

import (
	"sync"
	"time"
)

var MAX_CHAT_HISTORY = 1000
var MAX_CHAT_MESSAGE_LENGTH = 1000

type ChatMessageType string

const (
	ChatMessageTypeText   ChatMessageType = "text"
	ChatMessageTypeSystem ChatMessageType = "system"
)

type ChatMessage struct {
	SenderID   string          `json:"senderId"`
	SenderName string          `json:"senderName,omitempty"`
	Message    string          `json:"message"`
	Timestamp  time.Time       `json:"timestamp"`
	Type       ChatMessageType `json:"type"`
}

type ChatService struct {
	history []ChatMessage
	mu      sync.Mutex
}

func NewChatService() *ChatService {
	return &ChatService{
		history: make([]ChatMessage, 0),
	}
}

func (cs *ChatService) GetHistory(lastN int, skip int) []ChatMessage {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	hLen := len(cs.history)
	if hLen == 0 || skip >= hLen {
		return nil
	}

	// 1. Calculate the 'end' point (excluding the most recent 'skip' items)
	end := hLen - skip

	// 2. Calculate the 'start' point
	start := 0
	if lastN > 0 && lastN < end {
		start = end - lastN
	}

	// 3. Create a clean copy to prevent memory leaks/race conditions
	resultLen := end - start
	if resultLen <= 0 {
		return nil
	}

	result := make([]ChatMessage, resultLen)
	copy(result, cs.history[start:end])

	return result
}

func (cs *ChatService) AddMessage(msg ChatMessage) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Limit history to last 1000 messages
	if len(cs.history) >= MAX_CHAT_HISTORY {
		cs.history = cs.history[1:]
	}

	cs.history = append(cs.history, ChatMessage{
		SenderID:   msg.SenderID,
		SenderName: msg.SenderName,
		Message:    msg.Message,
		Timestamp:  msg.Timestamp,
		Type:       msg.Type,
	})
}

func (cs *ChatService) ClearHistory() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.history = make([]ChatMessage, 0)
}
