package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Dollar struct {
	USD struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func getDollarRate() (string, error) {
	url := "http://localhost:8080/cotacao"
	timeout := 3000 * time.Millisecond
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro: status code %d", res.StatusCode)
	}

	var cotacao Dollar
	err = json.NewDecoder(res.Body).Decode(&cotacao)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer parse da resposta: %v", err)
	}

	return cotacao.USD.Bid, nil
}

func main() {
	bid, err := getDollarRate()
	if err != nil {
		log.Printf("Erro ao obter cotação: %v\n", err)
		if os.IsTimeout(err) {
			log.Printf("A requisição ultrapassou o tempo limite de 300ms\n")
		}
	} else {
		fmt.Printf("Cotação do Dólar: %s\n", bid)

		
		file, err := os.Create("cotacao.txt")
		if err != nil {
			log.Fatalf("Erro ao criar arquivo: %v\n", err)
		}
		defer file.Close()

		_, err = file.WriteString(fmt.Sprintf("Dólar: %s\n", bid))
		if err != nil {
			log.Fatalf("Erro ao escrever no arquivo: %v\n", err)
		}
	}
}
