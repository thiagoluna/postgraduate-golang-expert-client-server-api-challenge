package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	serverAddress      = ":8080"
	externalAPIURL     = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	externalAPITimeout = 200 * time.Millisecond
	dbTimeout          = 10 * time.Millisecond
)

type awesomeAPIResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type cotacaoResponse struct {
	Bid string `json:"bid"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		log.Fatalf("erro ao abrir banco de dados: %v", err)
	}
	defer db.Close()

	if err := createTable(); err != nil {
		log.Fatalf("erro ao criar tabela: %v", err)
	}

	http.HandleFunc("/cotacao", cotacaoHandler)
	log.Printf("servidor iniciado em %s", serverAddress)
	if err := http.ListenAndServe(serverAddress, nil); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	bid, err := fetchDollarBid(r.Context())
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("timeout ao consultar API externa: %v", err)
			http.Error(w, "timeout ao consultar cotacao", http.StatusGatewayTimeout)
			return
		}

		log.Printf("erro ao consultar API externa: %v", err)
		http.Error(w, "erro ao consultar cotacao", http.StatusInternalServerError)
		return
	}

	if err := saveBid(r.Context(), bid); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("timeout ao salvar cotacao no banco: %v", err)
			http.Error(w, "timeout ao salvar cotacao", http.StatusGatewayTimeout)
			return
		}

		log.Printf("erro ao salvar cotacao no banco: %v", err)
		http.Error(w, "erro ao salvar cotacao", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cotacaoResponse{Bid: bid}); err != nil {
		log.Printf("erro ao serializar resposta: %v", err)
	}
}

func fetchDollarBid(parent context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(parent, externalAPITimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, externalAPIURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", fmt.Errorf("api externa retornou status %d e nao foi possivel ler body: %w", resp.StatusCode, readErr)
		}

		bodyMessage := strings.TrimSpace(string(bodyBytes))
		if bodyMessage == "" {
			bodyMessage = "<body vazio>"
		}

		const maxLogBodySize = 300
		if len(bodyMessage) > maxLogBodySize {
			bodyMessage = bodyMessage[:maxLogBodySize] + "..."
		}

		return "", fmt.Errorf("api externa retornou status %d com mensagem: %s", resp.StatusCode, bodyMessage)
	}

	var parsed awesomeAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	if parsed.USDBRL.Bid == "" {
		return "", errors.New("campo bid vazio na resposta da api")
	}

	return parsed.USDBRL.Bid, nil
}

func saveBid(parent context.Context, bid string) error {
	ctx, cancel := context.WithTimeout(parent, dbTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO cotacoes (bid) VALUES (?)", bid)
	return err
}

func createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(query)
	return err
}
