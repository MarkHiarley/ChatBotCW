#!/bin/bash

echo "🤖 Coletando respostas do ChatBot..."
echo ""

echo "====== PERGUNTA 1 ======"
echo "O que é a Cloudwalk?"
echo ""
curl -s -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"question": "O que é a Cloudwalk?"}'
echo ""
echo ""

echo "====== PERGUNTA 2 ======"
echo "Quais são os principais produtos da Cloudwalk?"
echo ""
curl -s -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"question": "Quais são os principais produtos da Cloudwalk?"}'
echo ""
echo ""

echo "====== PERGUNTA 3 ======"
echo "Qual é a missão da Cloudwalk?"
echo ""
curl -s -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"question": "Qual é a missão da Cloudwalk?"}'
echo ""
echo ""

echo "✅ Respostas coletadas!"