package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/jlhal/parejas/config"
	"github.com/jlhal/parejas/internal/chat"
	"github.com/jlhal/parejas/internal/models"
	"github.com/jlhal/parejas/internal/relationships"
	wsPkg "github.com/jlhal/parejas/pkg/websocket"
)

type Service interface {
	ProcessAIPrompt(userID string, userMessage string) error
}

type service struct {
	cfg         *config.Config
	chatService chat.Service
	relService  relationships.Service
}

func NewService(cfg *config.Config, chatService chat.Service, relService relationships.Service) Service {
	return &service{cfg, chatService, relService}
}

// Minimal struct for Groq completions
type GroqRequest struct {
	Model    string        `json:"model"`
	Messages []GroqMessage `json:"messages"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *service) ProcessAIPrompt(userID string, userMessage string) error {
	rel, err := s.relService.GetMyRelationship(userID)
	if err != nil || rel == nil {
		return errors.New("relationship not found")
	}

	// In a real app, you might want to fetch the last 10 messages from chatService here for context.
	systemPrompt := fmt.Sprintf(`Eres el consejero e IA de una pareja. 
Su relación empezó el: %s
Historia de cómo se conocieron: %s
Responde de forma concisa y amigable a las dudas de la pareja. Tu meta es mejorar su relación.`,
		rel.StartDate.Format("2006-01-02"), rel.HowWeMet)

	reqBody := GroqRequest{
		Model: "llama3-70b-8192", // o mixtral-8x7b-32768
		Messages: []GroqMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
	}

	jsonBytes, _ := json.Marshal(reqBody)
	
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Bearer "+s.cfg.GroqAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("groq api error %d: %s", resp.StatusCode, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return err
	}

	if len(groqResp.Choices) == 0 {
		return errors.New("empty response from Groq")
	}

	aiAnswer := groqResp.Choices[0].Message.Content

	// 1. Guardar mensaje de la IA en DB
	msg := models.Message{
		RelationshipID: rel.ID,
		SenderID:       nil, // Es la IA
		MessageType:    "ai",
		Content:        aiAnswer,
	}

	savedMsg, err := s.chatService.SaveMessage(msg)
	if err != nil {
		return err
	}

	// 2. Transmitir en vivo en WS
	if wsPkg.GlobalHub != nil {
		wsPkg.GlobalHub.Broadcast <- wsPkg.WSEvent{
			Action:         "message",
			Data:           savedMsg,
			RelationshipID: rel.ID,
		}
	}

	return nil
}
