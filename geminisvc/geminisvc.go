package geminisvc

import (
	"context"
	"encoding/json"
	"github.com/aiocean/wireset/cachesvc"
	"github.com/google/generative-ai-go/genai"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"os"
	"strings"
	"time"
)

type GeminiSvc struct {
	Client   *genai.Client
	Model    *genai.GenerativeModel
	Cachesvc *cachesvc.CacheService
}

func NewGeminiSvcFromEnv(ctx context.Context, cacheSvc *cachesvc.CacheService) (*GeminiSvc, func(), error) {
	apiKey := os.Getenv("VERTEX_API_KEY")
	if apiKey == "" {
		return nil, nil, errors.New("missing VERTEX_API_KEY")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to create genai client")
	}

	cleanup := func() {
		if err := client.Close(); err != nil {
			panic(err)
		}
	}

	model := client.GenerativeModel("gemini-1.0-pro-latest")
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: 0,
		},
	}

	candidateCount := int32(1)
	model.CandidateCount = &candidateCount

	return &GeminiSvc{
		Client:   client,
		Cachesvc: cacheSvc,
		Model:    model,
	}, cleanup, nil
}

// GetSession returns a chat session
func (f *GeminiSvc) GetSession(sessionID string) (*genai.ChatSession, error) {

	if sessionCache, ok := f.Cachesvc.Get(sessionID); ok {
		session := sessionCache.(*genai.ChatSession)
		return session, nil
	}

	session := f.Model.StartChat()
	f.Cachesvc.SetWithTTL(sessionID, session, time.Minute*5)

	return session, nil
}

// SaveSession saves a chat session
func (f *GeminiSvc) SaveSession(sessionID string, session *genai.ChatSession) error {
	if len(session.History) > 5 {
		session.History = session.History[len(session.History)-3:]
	}

	f.Cachesvc.SetWithTTL(sessionID, session, time.Minute*5)
	return nil
}

// SendMessage sends a message to a recipient
func (f *GeminiSvc) SendMessage(ctx context.Context, message string, sessionID string) (*genai.GenerateContentResponse, error) {
	session, err := f.GetSession(sessionID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get session")
	}

	message = message + "\n[INSTRUMENT] Answer in Vietnamese"

	resp, err := session.SendMessage(ctx, genai.Text(message))
	if err != nil {
		// remove the error message from history
		session.History = session.History[:len(session.History)-1]
		return nil, errors.WithMessage(err, "failed to send message")
	}

	f.SaveSession(sessionID, session)

	return resp, nil
}

// EmbedDoc generates an embedding for retrieval
func (f *GeminiSvc) EmbedDoc(ctx context.Context, title, content string) ([]float32, error) {
	model := f.Client.EmbeddingModel("embedding-001")

	resp, err := model.EmbedContentWithTitle(ctx, title, genai.Text(content))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to generate embedding")
	}

	return resp.Embedding.Values, nil
}

// EmbedQuery generates an embedding for a query
func (f *GeminiSvc) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	model := f.Client.EmbeddingModel("embedding-001")
	model.TaskType = genai.TaskTypeRetrievalQuery

	resp, err := model.EmbedContent(ctx, genai.Text(query))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to generate embedding")
	}

	return resp.Embedding.Values, nil
}

// GenerateAnswer generates an answer for a query
func (f *GeminiSvc) GenerateAnswer(ctx context.Context, dataContext, query string) (*genai.GenerateContentResponse, error) {
	model := f.Client.GenerativeModel("gemini-pro")
	candidateCount := int32(1)
	model.CandidateCount = &candidateCount
	chatSession := model.StartChat()

	chatSession.History = append(chatSession.History, &genai.Content{
		Role: "user",
		Parts: []genai.Part{
			genai.Text("Only use the provided Context to answer my question. Do not make up anything. Answer should be short, concise and in the provided language.\n\nContext: " + dataContext),
		},
	})
	chatSession.History = append(chatSession.History, &genai.Content{
		Role: "model",
		Parts: []genai.Part{
			genai.Text("I will only use the provided Context to answer your question. Will not make up anything."),
		},
	})

	resp, err := chatSession.SendMessage(ctx, genai.Text(query))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to generate answer")
	}

	return resp, nil
}

type ParseStructuredDataExample struct {
	Input  string `json:"question"`
	Output any    `json:"answer"`
}

// ParseStructuredData parses structured data
func (f *GeminiSvc) ParseStructuredData(ctx context.Context, examples []ParseStructuredDataExample, data string, dest interface{}) error {

	var parts []genai.Part
	for _, example := range examples {
		parts = append(parts, genai.Text("Input: "+example.Input))
		output, _ := json.Marshal(example.Output)
		parts = append(parts, genai.Text("Output: "+string(output)))
	}

	parts = append(parts, genai.Text("Input: "+data))
	parts = append(parts, genai.Text("Output: "))

	resp, err := f.Model.GenerateContent(ctx, parts...)
	if err != nil {
		return errors.WithMessage(err, "failed to generate answer")
	}

	outputText := strings.TrimSpace(string(resp.Candidates[0].Content.Parts[0].(genai.Text)))

	if err := json.Unmarshal([]byte(outputText), dest); err != nil {
		return err
	}

	return nil
}
