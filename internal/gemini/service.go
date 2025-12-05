package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Service struct {
	apiKey string
	client *http.Client
}

// GeminiRequest representa uma requisição para a API do Gemini
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

// GeminiResponse representa uma resposta da API do Gemini
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

// NewService cria uma nova instância do serviço Gemini
func NewService() (*Service, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	return &Service{
		apiKey: apiKey,
		client: &http.Client{},
	}, nil
}

// GenerateEmbedding gera embeddings simulados para o texto
// NOTA: Implementação temporária sem embeddings reais devido a problemas de compatibilidade da SDK
func (s *Service) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Por enquanto, vamos simular embeddings usando hash simples do texto
	// Isso permite testar o resto da funcionalidade
	hash := simpleHash(text)
	embedding := make([]float32, 384) // Tamanho típico de embeddings

	// Preencher o vetor com valores baseados no hash
	for i := range embedding {
		embedding[i] = float32((hash+i*7)%1000)/1000.0 - 0.5
	}

	return embedding, nil
}

// simpleHash cria um hash simples de uma string
func simpleHash(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	return hash
}

// GenerateResponse gera uma resposta baseada no contexto e pergunta
func (s *Service) GenerateResponse(ctx context.Context, context, question string) (string, error) {
	prompt := fmt.Sprintf(`Você é um assistente virtual especializado em responder perguntas sobre a Cloudwalk.

CONTEXTO:
%s

PERGUNTA: %s

INSTRUÇÕES:
- Responda de forma clara, direta e objetiva
- Use apenas informações do contexto fornecido
- Se a pergunta for sobre "o que é a Cloudwalk", foque em: missão, produtos, serviços e área de atuação
- Organize a resposta em parágrafos curtos
- Se não houver informações suficientes no contexto, diga isso educadamente

RESPOSTA:`, context, question)

	fmt.Printf("🔍 DEBUG: Gerando resposta para pergunta: %s\n", question)
	fmt.Printf("📝 DEBUG: Contexto tem %d caracteres\n", len(context))

	request := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Usando gemini-1.5-flash com API v1
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", s.apiKey)

	fmt.Printf("🌐 DEBUG: Chamando API Gemini...\n")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ DEBUG: Erro ao fazer request: %v\n", err)
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📡 DEBUG: Status da resposta: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("📄 DEBUG: Resposta da API (primeiros 500 chars): %s\n", string(body[:min(500, len(body))]))

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ DEBUG: Erro na API - Status %d: %s\n", resp.StatusCode, string(body))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		fmt.Printf("❌ DEBUG: Erro ao fazer unmarshal: %v\n", err)
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		fmt.Printf("❌ DEBUG: Resposta vazia da API\n")
		return "", fmt.Errorf("no response generated")
	}

	answer := geminiResp.Candidates[0].Content.Parts[0].Text
	fmt.Printf("✅ DEBUG: Resposta gerada com sucesso (%d caracteres)\n", len(answer))
	return answer, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Close fecha a conexão com o cliente (não faz nada para HTTP client)
func (s *Service) Close() {
	// HTTP client não precisa ser fechado explicitamente
}
