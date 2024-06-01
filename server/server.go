package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/valyala/fastjson"
)

var endpointCotacaoDolar string = "http://economia.awesomeapi.com.br/json/last/USD-BRL"

type CotacaoDolar struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HomeHandler)
	mux.HandleFunc("/cotacao", CotacaoHandler)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Could not listen on %s: %v\n", ":8080", err)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("# Cotação do dólar"))
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	cotacao, err := RealizaCotacao(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(cotacao)
}

func RealizaCotacao(ctx context.Context) (*CotacaoDolar, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpointCotacaoDolar, nil)
	if err != nil {
		log.Fatalf("Erro ao criar requisição para o endpoint: %s. Erro: %s\n", endpointCotacaoDolar, err)
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Erro ao requisitar o endpoint: %s. Erro: %s\n", endpointCotacaoDolar, err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Erro ao ler a responsta do endpoint: %s. Erro: %s\n", endpointCotacaoDolar, err)
		return nil, err
	}

	var cotacao CotacaoDolar
	err = CotacaoResponseParser(body, &cotacao)
	if err != nil {
		log.Fatalf("Erro ao realizar a converção do objeto de cotação. Erro: %s\n", err)
		return nil, err
	}

	return &cotacao, nil
}

func CotacaoResponseParser(body []byte, cotacao *CotacaoDolar) error {
	var p fastjson.Parser
	value, err := p.Parse(string(body))
	if err != nil {
		return err
	}
	cotacaoJSON := value.Get("USDBRL").String()

	if err := json.Unmarshal([]byte(cotacaoJSON), &cotacao); err != nil {
		return err
	}

	return nil
}
