# ChatBot Cloudwalk

Um chatbot inteligente que responde perguntas sobre a Cloudwalk usando a API do Google Gemini.

## Configuração

1. **Obter API Key do Gemini**
   - Acesse [Google AI Studio](https://makersuite.google.com/app/apikey)
   - Crie uma nova API key
   - Copie a chave

2. **Configurar Variáveis de Ambiente**
   ```bash
   # Copie o arquivo de exemplo
   cp .env.example .env
   
   # Edite o arquivo .env e adicione sua GEMINI_API_KEY
   # Exemplo:
   # GEMINI_API_KEY=sua_chave_aqui
   ```

3. **Instalar Dependências**
   ```bash
   go mod tidy
   ```

## Executando

### Opção 1: Com Docker (Recomendado) 🐳

```bash
# Build e executar com docker-compose
docker-compose up -d

# Ver logs
docker-compose logs -f

# Parar o serviço
docker-compose down
```

### Opção 2: Localmente com Go

```bash
# Com arquivo .env (recomendado)
# Basta ter o arquivo .env configurado com a GEMINI_API_KEY
go run cmd/api/main.go

# Ou com variável de ambiente
export GEMINI_API_KEY="sua_api_key_aqui"
go run cmd/api/main.go
```

O servidor será iniciado na porta 8080.

**⚡ Sistema de Cache**: Na primeira execução, o sistema irá fazer scraping e gerar embeddings (pode demorar alguns minutos). Os dados são salvos em `cache/documents_cache.json` e nas próximas execuções serão carregados automaticamente, tornando a inicialização muito mais rápida!

## Endpoints

### POST /chat
Envia uma pergunta para o chatbot.

**Exemplo de requisição:**
```json
{
  "question": "O que é a Cloudwalk?"
}
```

**Exemplo de resposta:**
```json
{
  "answer": "A Cloudwalk é uma empresa de tecnologia financeira..."
}
```

### GET /health
Health check do serviço.

### POST /debug-search
Endpoint de debug para ver quais documentos foram encontrados (apenas desenvolvimento).

## Arquitetura e Tecnologias

### Stack Tecnológico
- **Backend**: Go (Golang) 1.23+
- **IA**: Google Gemini API (gemini-1.5-flash)
- **Web Scraping**: Colly
- **Containerização**: Docker & Docker Compose
- **Cache**: Sistema de persistência em JSON

### Fluxo RAG (Retrieval-Augmented Generation)

1. **Ingestão de Dados**
   - Web scraping dos sites oficiais (cloudwalk.io, infinitepay.io)
   - Coleta de ~2500 documentos com informações relevantes
   - Filtros de qualidade (tamanho mínimo, limpeza de texto)

2. **Geração de Embeddings**
   - Conversão de texto em vetores numéricos
   - Sistema de embeddings simulados (hash-based) para testes
   - Cache persistente para evitar reprocessamento

3. **Busca Híbrida**
   - Similaridade de cosseno entre vetores
   - Busca por palavras-chave com boost
   - Priorização de documentos informativos
   - Retorna top 15 documentos mais relevantes

4. **Geração de Resposta**
   - Contexto montado a partir dos documentos encontrados
   - Prompt engineering otimizado
   - API do Gemini para gerar respostas naturais
   - Validação e formatação da resposta

### Estrutura do Projeto
```
├── cmd/api/main.go              # Servidor HTTP principal
├── internal/
│   ├── cache/cache.go           # Sistema de cache
│   ├── gemini/service.go        # Cliente Gemini API
│   ├── scraper/scraper.go       # Web scraping
│   └── vectorstore/store.go     # Busca por similaridade
├── pkg/models/document.go       # Modelos de dados
├── cache/                       # Cache persistente (gitignored)
├── Dockerfile                   # Container da aplicação
└── docker-compose.yaml          # Orquestração
```

## 3 Exemplos de Conversas

### Conversa 1: Sobre a Empresa

**Pergunta:** "O que é a Cloudwalk?"

**Resposta:**
```json
{
  "answer": "A CloudWalk é uma empresa global de tecnologia financeira (fintech), fundada em maio de 2013 no Brasil. Sua missão é transformar a forma como comerciantes lidam com dinheiro, criando produtos e serviços financeiros mais acessíveis e inovadores, com o objetivo de democratizar o acesso a esses serviços e construir uma nova arquitetura financeira.\n\nA empresa é a criadora da plataforma de serviços financeiros InfinitePay, que oferece soluções para pequenas e médias empresas (PMEs) e autônomos. Entre os serviços do InfinitePay estão o processamento de pagamentos (incluindo a tecnologia Tap to Pay em iPhones e Android, que transforma smartphones em terminais de pagamento), concessão de crédito e empréstimos (como empréstimos inteligentes e adiantamento de recebíveis), conta digital (InfiniteBank com conta remunerada), gestão de cobranças (SuperCobra), e gestão de vendas e estoque.\n\nA CloudWalk se destaca pelo uso intensivo de Inteligência Artificial (IA) e tecnologia blockchain. A IA é integrada em todas as áreas da empresa, desde atendimento ao cliente (\"Claudio Walker\") e prevenção de fraudes (prevenindo perdas de bilhões), até concessão de crédito, desenvolvimento de produtos e operações internas. A empresa também desenvolveu sua própria blockchain de código aberto, a Stratus, para garantir segurança e agilidade nas transações.\n\nCom um alcance que abrange 100% das cidades brasileiras, a CloudWalk expandiu suas operações para os Estados Unidos com o aplicativo Jim.com, que funciona como um assistente financeiro inteligente para comerciantes. Há planos de expansão para o Sudeste Asiático e o Oriente Médio. Em 2023, a CloudWalk registrou sua primeira lucratividade, com receita de US$ 320.5 milhões e lucro líquido de US$ 22.3 milhões, e em 2024 fechou com US$497 milhões em receita. Em 2025, superou US$1.2 bilhão em receita anualizada e atingiu mais de 5 milhões de clientes ativos no InfinitePay."
}
```

### Conversa 2: Sobre Produtos

**Pergunta:** "Quais são os principais produtos da Cloudwalk?"

**Resposta:**
```json
{
  "answer": "Os principais produtos da CloudWalk são:\n\n*   **InfinitePay**: Uma plataforma de serviços financeiros que oferece uma variedade de soluções, incluindo a tecnologia Tap to Pay (que transforma smartphones em terminais de pagamento), conta digital, Pix com crédito, cartões (InfiniteCard), operações de crédito/empréstimos (Intelligent Credit, Smart Loans), e o produto de investimento CloudWalk InfinitePay CDB. Também oferece ferramentas de gestão como o Magic POS e o sistema de cobrança SuperCobra, além de uma solução Buy Now, Pay Later.\n*   **JIM (e JIM.com)**: Um assistente inteligente integrado ao aplicativo InfinitePay, que atua como um \"funcionário digital\" para empreendedores, ajudando em tarefas como revisão de preços, organização de recebíveis, criação de campanhas de marketing e geração de conteúdo. JIM.com é a marca com a qual a CloudWalk expandiu para os Estados Unidos, oferecendo ferramentas de inteligência artificial e soluções de pagamento instantâneo para comerciantes.\n*   **Stratus**: A blockchain proprietária de próxima geração da CloudWalk, lançada publicamente no final de 2024. Ela foi desenvolvida para garantir segurança e agilidade nas transações, com capacidade para mais de mil transações por segundo e compatibilidade com smart contracts via EVM.\n*   **InfiniteBank**: Um banco digital que oferece uma gama completa de soluções bancárias e de pagamento, indo além das máquinas de cartão e focando em serviços bancários completos para seus usuários."
}
```

### Conversa 3: Sobre Missão

**Pergunta:** "Qual é a missão da Cloudwalk?"

**Resposta:**
```json
{
  "answer": "A missão da CloudWalk é entregar as soluções financeiras mais inovadoras para empreendedores e negócios globalmente.\n\nA empresa se dedica a democratizar o acesso a serviços financeiros inovadores, visando construir uma nova arquitetura financeira. Essa nova lógica de sistema financeiro busca ser mais ágil, inteligente e acessível a todos, utilizando tecnologias como inteligência artificial, blockchain e o desenvolvimento de talentos para substituir estruturas lentas, caras e centralizadas."
}
```

## Testando

```bash
# Health check
curl http://localhost:8080/health

# Teste de chat
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"question": "O que é a Cloudwalk?"}'
```

## Como Funciona

1. **Scraping**: O sistema coleta informações dos sites da Cloudwalk
2. **Embeddings**: Usando Gemini, converte o texto em vetores numéricos
3. **Vector Store**: Armazena os documentos com seus embeddings em memória
4. **Busca**: Para cada pergunta, encontra documentos similares
5. **Geração**: Usa Gemini para gerar respostas baseadas no contexto encontrado

## Estrutura do Projeto

```
├── cmd/api/main.go              # Servidor HTTP principal
├── internal/
│   ├── gemini/service.go        # Cliente Gemini
│   ├── scraper/scraper.go       # Web scraping
│   └── vectorstore/store.go     # Armazenamento de vetores
├── pkg/models/document.go       # Modelos de dados
└── docker-compose.yaml         # Configuração Docker
```