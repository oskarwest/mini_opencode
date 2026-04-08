package chat

import "github.com/oscar/my_opencode/internal/apiclient"

// Session manages the conversation state
type Session struct {
	model    string
	messages []apiclient.Message
}

// NewSession creates a new chat session
func NewSession(model string) *Session {
	return &Session{
		model:    model,
		messages: make([]apiclient.Message, 0),
	}
}

// InitializeWithSystemPrompt adds a system prompt as the first message
func (s *Session) InitializeWithSystemPrompt(systemPrompt string) {
	// Only add if there are no messages yet
	if len(s.messages) == 0 {
		s.messages = append(s.messages, apiclient.Message{
			Role:    "system",
			Content: systemPrompt,
		})
	}
}

// AddMessage adds a message to the conversation history
func (s *Session) AddMessage(role, content string) {
	s.messages = append(s.messages, apiclient.Message{
		Role:    role,
		Content: content,
	})
}

// GetMessages returns the current conversation history
func (s *Session) GetMessages() []apiclient.Message {
	return s.messages
}

// GetModel returns the current model
func (s *Session) GetModel() string {
	return s.model
}

// SetModel changes the current model
func (s *Session) SetModel(model string) {
	s.model = model
}

// Clear resets the conversation history but keeps the system prompt
func (s *Session) Clear() {
	// Keep system prompt if it exists (first message with role "system")
	if len(s.messages) > 0 && s.messages[0].Role == "system" {
		systemPrompt := s.messages[0]
		s.messages = []apiclient.Message{systemPrompt}
	} else {
		s.messages = make([]apiclient.Message, 0)
	}
}
