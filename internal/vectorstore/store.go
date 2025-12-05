package vectorstore

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/markHiarley/chatbotCW/pkg/models"
)

type Store struct {
	Data []models.Document
}

func NewStore(docs []models.Document) *Store {
	return &Store{Data: docs}
}

// SearchWithKeywords - BUSCA PRINCIPAL (melhorada)
func (s *Store) SearchWithKeywords(queryVector []float32, keywords []string) []models.Document {
	type result struct {
		doc         models.Document
		score       float64
		keywordHits int
		debugInfo   string
	}

	var results []result

	// Palavras-chave por categoria
	productKeywords := []string{"produto", "serviço", "solução", "oferece", "maquininha", "infinitepay", "stratus", "conta", "digital", "pagamento", "pix", "cartão"}
	companyKeywords := []string{"cloudwalk", "empresa", "fintech", "tecnologia", "missão", "objetivo", "atua"}

	for _, doc := range s.Data {
		contentLower := strings.ToLower(doc.Content)

		// 1. Score base (cosine similarity)
		baseScore := CosineSimilarity(queryVector, doc.Vector)

		// 2. Conta keywords da pergunta do usuário
		queryKeywordHits := 0
		for _, keyword := range keywords {
			if strings.Contains(contentLower, strings.ToLower(keyword)) {
				queryKeywordHits++
			}
		}

		// 3. Conta keywords de produtos
		productHits := 0
		for _, keyword := range productKeywords {
			if strings.Contains(contentLower, keyword) {
				productHits++
			}
		}

		// 4. Conta keywords de empresa
		companyHits := 0
		for _, keyword := range companyKeywords {
			if strings.Contains(contentLower, keyword) {
				companyHits++
			}
		}

		// 5. Boost por comprimento (documentos maiores = mais informação)
		lengthScore := math.Min(float64(len(doc.Content))/500.0, 2.0)

		// 6. Boost se contém múltiplas informações relevantes
		densityBoost := 0.0
		if productHits >= 3 || companyHits >= 3 {
			densityBoost = 1.0
		}

		// CÁLCULO FINAL DO SCORE
		// Pesos ajustados para priorizar keywords sobre embeddings simulados
		finalScore := baseScore*0.1 + // embedding tem pouco peso
			float64(queryKeywordHits)*2.0 + // keywords da pergunta são importantes
			float64(productHits)*1.5 + // keywords de produtos
			float64(companyHits)*1.0 + // keywords de empresa
			lengthScore*0.3 + // documentos maiores
			densityBoost // bonus por densidade

		debugInfo := fmt.Sprintf("base:%.2f query:%d prod:%d comp:%d len:%.2f",
			baseScore, queryKeywordHits, productHits, companyHits, lengthScore)

		if finalScore > 0.1 { // threshold mínimo
			results = append(results, result{
				doc:         doc,
				score:       finalScore,
				keywordHits: queryKeywordHits + productHits + companyHits,
				debugInfo:   debugInfo,
			})
		}
	}

	// Ordenar por score
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Debug: mostrar top 5 resultados
	fmt.Println("\n🔍 TOP 5 DOCUMENTOS ENCONTRADOS:")
	for i := 0; i < min(5, len(results)); i++ {
		fmt.Printf("%d. Score: %.2f | Hits: %d | %s\n",
			i+1, results[i].score, results[i].keywordHits,
			truncate(results[i].doc.Content, 100))
	}

	// Retorna top 20 para ter contexto suficiente
	limit := 20
	if len(results) < limit {
		limit = len(results)
	}

	finalDocs := make([]models.Document, limit)
	for i := 0; i < limit; i++ {
		finalDocs[i] = results[i].doc
	}

	return finalDocs
}

// Search - busca básica por similaridade
func (s *Store) Search(queryVector []float32) []models.Document {
	// Extrai keywords básicas do vetor (simulação)
	keywords := []string{"cloudwalk", "produto", "serviço"}
	return s.SearchWithKeywords(queryVector, keywords)
}

func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
