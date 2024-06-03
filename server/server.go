package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/valyala/fastjson"
)

var (
	endpointCotacaoDolar string = "http://economia.awesomeapi.com.br/json/last/USD-BRL"
	globalDB             *sql.DB
)

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

func (c CotacaoDolar) String() string {
	data, err := json.Marshal(c)
	if err != nil {
		log.Println("Erro ao converter contação para string.", c)
		return ""
	}
	return string(data)
}

func main() {
	initDB()
	defer globalDB.Close()
	createTableCotacao()

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
	cotacao, err := contacaoService()
	if err != nil {
		serverError(w, err)
		return
	}
	serverSuccess(w)
	json.NewEncoder(w).Encode(cotacao)
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("ERROR: %s", err.Error())
	http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
}

func serverSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func contacaoService() (*CotacaoDolar, error) {
	reqCtx, reqCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer reqCancel()
	cotacao, err := RealizaCotacao(reqCtx)
	if err != nil {
		return nil, err
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer dbCancel()
	persist(dbCtx, cotacao.String())

	return cotacao, err
}

func RealizaCotacao(reqCtx context.Context) (*CotacaoDolar, error) {
	req, err := http.NewRequestWithContext(reqCtx, "GET", endpointCotacaoDolar, nil)
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
	cotacaoString := value.Get("USDBRL").String()

	if err := json.Unmarshal([]byte(cotacaoString), &cotacao); err != nil {
		return err
	}

	return nil
}

func initDB() {
	fmt.Println("Init DB ...")
	var err error
	globalDB, err = sql.Open("sqlite3", "./cotacao.db")
	if globalDB == nil || err != nil {
		log.Fatal(err)
	}
	fmt.Println("Testing db connection ...")
	if err = globalDB.Ping(); err != nil {
		log.Fatal("Error on opening database connection: %s", err.Error())
	}
}

func createTableCotacao() {
	_, err := globalDB.Exec(`CREATE TABLE IF NOT EXISTS cotacao (tag jsonb)`)
	if err != nil {
		log.Fatal(err)
	}
}

func persist(ctx context.Context, data string) {
	stmt, err := globalDB.Prepare("insert into cotacao(tag) values(?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database storage complete!")
}
