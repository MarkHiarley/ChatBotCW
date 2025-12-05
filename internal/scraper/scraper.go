package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/markHiarley/chatbotCW/pkg/models"
)

func ScraperCloudwalk() ([]models.Document, error) {
	var docs []models.Document
	visited := make(map[string]bool)

	c := colly.NewCollector(
		colly.AllowedDomains("www.cloudwalk.io", "cloudwalk.io", "infinitepay.io", "help.infinitepay.io"),
		colly.MaxDepth(3), // Aumentado para 3
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5, // Mais threads
		RandomDelay: 300 * time.Millisecond,
	})

	// Captura blocos de texto maiores e mais contextualizados
	c.OnHTML("article, section, main, div.content, div[class*='content'], div[class*='text']", func(e *colly.HTMLElement) {
		// Pega todo o texto do bloco
		text := strings.TrimSpace(e.Text)

		// Remove espaços extras e quebras de linha excessivas
		text = strings.Join(strings.Fields(text), " ")

		// Só adiciona se for substancial (mais de 50 caracteres)
		if len(text) > 50 && !strings.Contains(text, "Cookie") {
			docs = append(docs, models.Document{
				Content: text,
				Source:  e.Request.URL.String(),
			})
		}
	})

	// Também captura parágrafos individuais como backup
	c.OnHTML("p, li", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		text = strings.Join(strings.Fields(text), " ")

		if len(text) > 30 {
			docs = append(docs, models.Document{
				Content: text,
				Source:  e.Request.URL.String(),
			})
		}
	})

	// Captura títulos com contexto
	c.OnHTML("h1, h2, h3", func(e *colly.HTMLElement) {
		title := strings.TrimSpace(e.Text)
		if len(title) > 5 {
			// Tenta pegar o parágrafo seguinte para dar contexto
			nextText := ""
			if p := e.DOM.Next().Text(); len(p) > 0 {
				nextText = " - " + strings.TrimSpace(p)
			}

			docs = append(docs, models.Document{
				Content: title + nextText,
				Source:  e.Request.URL.String(),
			})
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if !visited[link] && link != "" {
			visited[link] = true
			e.Request.Visit(link)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Scraping:", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Erro ao acessar %s: %v\n", r.Request.URL, err)
	})

	// URLs expandidas - adicione TODAS as páginas importantes
	startUrls := []string{
		// CloudWalk principal
		"https://www.cloudwalk.io/",
		"https://www.cloudwalk.io/en",
		"https://www.cloudwalk.io/en/mission",
		"https://www.cloudwalk.io/en/about",
		"https://www.cloudwalk.io/en/products",
		"https://www.cloudwalk.io/en/solutions",

		// Produtos específicos
		"https://www.cloudwalk.io/infinitepay",
		"https://www.cloudwalk.io/stratus",
		"https://www.cloudwalk.io/en/infinitepay",
		"https://www.cloudwalk.io/en/stratus",

		// InfinitePay
		"https://infinitepay.io/",
		"https://infinitepay.io/maquininha",
		"https://infinitepay.io/produtos",
		"https://infinitepay.io/conta-digital",

		// Help/Docs
		"https://help.infinitepay.io/",
	}

	for _, url := range startUrls {
		c.Visit(url)
	}

	c.Wait()

	fmt.Printf("✅ Total de documentos coletados: %d\n", len(docs))
	return docs, nil
}
