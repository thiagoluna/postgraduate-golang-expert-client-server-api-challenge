package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	serverURL      = "http://localhost:8080/cotacao"
	clientTimeout  = 300 * time.Millisecond
	outputFileName = "cotacao.txt"
)

type cotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		log.Fatalf("erro ao criar requisicao: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Printf("timeout ao chamar servidor: %v", err)
			return
		}

		log.Fatalf("erro ao chamar servidor: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("servidor retornou status %d", resp.StatusCode)
	}

	var parsed cotacaoResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		log.Fatalf("erro ao ler resposta do servidor: %v", err)
	}

	if parsed.Bid == "" {
		log.Fatal("resposta do servidor sem campo bid")
	}

	content := fmt.Sprintf("Dólar: %s", parsed.Bid)
	if err := os.WriteFile(outputFileName, []byte(content), 0644); err != nil {
		log.Fatalf("erro ao salvar arquivo: %v", err)
	}

	log.Printf("cotacao salva em %s: %s", outputFileName, content)
}
