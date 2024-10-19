package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Dollar struct {
	USD struct {
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
	} `json:"USDBRL"`
}

func GetDollarRate() (Dollar, error) {
	var cotacao Dollar
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	timeout := 200 * time.Millisecond
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %v\n", err)
		return cotacao, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %v\n", err)
		return cotacao, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
		return cotacao, err
	}

	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
		return cotacao, err
	}

	return cotacao, nil
}

func createTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS cotacao (
        id INT AUTO_INCREMENT PRIMARY KEY,
        code VARCHAR(50),
        codein VARCHAR(50),
        name VARCHAR(100),
        high VARCHAR(50),
        low VARCHAR(50),
        varBid VARCHAR(50),
        pctChange VARCHAR(50),
        bid VARCHAR(50),
        ask VARCHAR(50),
        timestamp VARCHAR(50),
        create_date VARCHAR(50)
    );`
	statement, err := db.Prepare(createTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Tabela criada com sucesso")
}

func insertCotacao(db *sql.DB, cotacao Dollar) error {
	insertCotacaoSQL := `INSERT INTO cotacao (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	statement, err := db.PrepareContext(ctx, insertCotacaoSQL)
	if err != nil {
		return err
	}
	_, err = statement.ExecContext(
		ctx,
		cotacao.USD.Code,
		cotacao.USD.Codein,
		cotacao.USD.Name,
		cotacao.USD.High,
		cotacao.USD.Low,
		cotacao.USD.VarBid,
		cotacao.USD.PctChange,
		cotacao.USD.Bid,
		cotacao.USD.Ask,
		cotacao.USD.Timestamp,
		cotacao.USD.CreateDate,
	)
	if err != nil {
		return err
	}
	fmt.Println("Cotação inserida com sucesso")
	return nil
}

func BuscaCotacao(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cotacao, err := GetDollarRate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dsn := "root:root@tcp(localhost:3306)/goexpert"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable(db)
	if err := insertCotacao(db, cotacao); err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao)
}

func main() {
	http.HandleFunc("/cotacao", BuscaCotacao)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
