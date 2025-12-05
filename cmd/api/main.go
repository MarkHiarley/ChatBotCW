package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/markHiarley/chatbotCW/internal/cache"
	"github.com/markHiarley/chatbotCW/internal/gemini"
	"github.com/markHiarley/chatbotCW/internal/scraper"
	"github.com/markHiarley/chatbotCW/internal/vectorstore"
	"github.com/markHiarley/chatbotCW/pkg/models"
)

// ChatRequest representa a estrutura da requisição de chat
type ChatRequest struct {
	Question string `json:"question"`
}

// ChatResponse representa a estrutura da resposta de chat
type ChatResponse struct {
	Answer string `json:"answer"`
	Error  string `json:"error,omitempty"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// 0. Carregar variáveis de ambiente do arquivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	} else {
		log.Println("✅ Arquivo .env carregado com sucesso")
	}

	// 1. Inicializar serviço Gemini
	geminiService, err := gemini.NewService()
	if err != nil {
		log.Fatalf("Erro ao inicializar serviço Gemini: %v", err)
	}
	defer geminiService.Close()

	var docsWithEmbeddings []models.Document

	// 2. Tentar carregar do cache primeiro
	if cache.IsCacheValid() {
		fmt.Println("Carregando dados do cache...")
		cachedDocs, err := cache.LoadFromCache()
		if err == nil {
			docsWithEmbeddings = cachedDocs
			fmt.Printf("✅ Cache carregado com sucesso! %d documentos encontrados.\n", len(docsWithEmbeddings))
		} else {
			fmt.Printf("❌ Erro ao carregar cache: %v\n", err)
		}
	}

	// 3. Se não tiver cache válido, fazer scraping e gerar embeddings
	if len(docsWithEmbeddings) == 0 {
		// Fazer scraping
		fmt.Println("Iniciando scraping...")
		rawDocs, err := scraper.ScraperCloudwalk()
		if err != nil {
			log.Fatalf("Erro no scraping: %v", err)
		}

		// Gerar embeddings para todos os documentos
		fmt.Println("Gerando embeddings...")
		ctx := context.Background()
		fmt.Println("Gerando embeddings (pode demorar um pouco)...")
		bar := 0
		total := len(rawDocs)
		for i, doc := range rawDocs {

			cleanContent := strings.TrimSpace(doc.Content)
			if len(cleanContent) < 20 {
				continue
			}

			embedding, err := geminiService.GenerateEmbedding(ctx, cleanContent)
			if err != nil {

				log.Printf("⚠️ Erro no doc %d: %v", i, err)
				continue
			}

			doc.Vector = embedding
			doc.ID = fmt.Sprintf("doc_%d", i)
			docsWithEmbeddings = append(docsWithEmbeddings, doc)

			// Feedback visual simples
			bar++
			if bar%10 == 0 {
				fmt.Printf("Processados: %d/%d\n", bar, total)
			}

			time.Sleep(100 * time.Millisecond)
		}

		fmt.Printf("Processados %d documentos com embeddings\n", len(docsWithEmbeddings))

		// 4. Salvar no cache para próximas execuções
		fmt.Println("Salvando dados no cache...")
		if err := cache.SaveToCache(docsWithEmbeddings); err != nil {
			log.Printf("⚠️ Erro ao salvar cache: %v", err)
		} else {
			fmt.Println("✅ Cache salvo com sucesso!")
		}
	}

	// 4. Inicializar a Vector Store
	store := vectorstore.NewStore(docsWithEmbeddings)

	// 5. Contexto para as operações
	ctx := context.Background()

	// 6. Endpoint de chat
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response := ChatResponse{Error: "Erro ao decodificar JSON"}
			json.NewEncoder(w).Encode(response)
			return
		}

		if strings.TrimSpace(req.Question) == "" {
			response := ChatResponse{Error: "Pergunta não pode estar vazia"}
			json.NewEncoder(w).Encode(response)
			return
		}

		// A. Gerar embedding da pergunta
		queryEmbedding, err := geminiService.GenerateEmbedding(ctx, req.Question)
		if err != nil {
			response := ChatResponse{Error: "Erro ao processar pergunta"}
			json.NewEncoder(w).Encode(response)
			return
		}

		// B. Extrair palavras-chave da pergunta
		keywords := strings.Fields(strings.ToLower(req.Question))

		// C. Buscar contexto no store (híbrido: vetores + keywords)
		relevantDocs := store.SearchWithKeywords(queryEmbedding, keywords)

		// D. Montar contexto
		var contextParts []string
		for _, doc := range relevantDocs {
			if len(strings.TrimSpace(doc.Content)) > 0 {
				contextParts = append(contextParts, doc.Content)
			}
		}
		contextText := strings.Join(contextParts, "\n\n")

		// D. Gerar resposta com Gemini
		answer, err := geminiService.GenerateResponse(ctx, contextText, req.Question)
		if err != nil {
			response := ChatResponse{Error: "Erro ao gerar resposta"}
			json.NewEncoder(w).Encode(response)
			return
		}

		response := ChatResponse{Answer: answer}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/clear-cache", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		if err := cache.ClearCache(); err != nil {
			http.Error(w, "Erro ao limpar cache", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Cache limpo com sucesso!"))
	})

	// 9. Endpoint de debug para ver documentos encontrados
	http.HandleFunc("/debug-search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
			return
		}

		queryEmbedding, _ := geminiService.GenerateEmbedding(ctx, req.Question)
		keywords := strings.Fields(strings.ToLower(req.Question))
		relevantDocs := store.SearchWithKeywords(queryEmbedding, keywords)

		type DebugDoc struct {
			Content string `json:"content"`
			Source  string `json:"source"`
			Length  int    `json:"length"`
		}

		var debugDocs []DebugDoc
		for _, doc := range relevantDocs {
			debugDocs = append(debugDocs, DebugDoc{
				Content: doc.Content[:min(200, len(doc.Content))] + "...",
				Source:  doc.Source,
				Length:  len(doc.Content),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"question":      req.Question,
			"keywords":      keywords,
			"documents":     debugDocs,
			"total_context": len(relevantDocs),
		})
	})

	fmt.Println("🚀 Servidor iniciado na porta 8080")
	fmt.Println("📋 Endpoints disponíveis:")
	fmt.Println("  POST /chat - Chat com o bot")
	fmt.Println("  GET  /health - Health check")
	fmt.Println("  POST /clear-cache - Limpar cache de documentos")
	fmt.Println("  POST /debug-search - Debug: ver documentos encontrados (dev only)")
	fmt.Printf("💾 Documentos carregados: %d\n", len(docsWithEmbeddings))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
