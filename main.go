package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	cep            = "69304350"
	brasilApiURL   = "https://brasilapi.com.br/api/cep/v1/"
	viaCepURL      = "http://viacep.com.br/ws/"
	requestTimeout = 1 * time.Second
)

type AddressViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type AddressBrasilAPI struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

func fetchFromAPI(ctx context.Context, url string, ch chan<- *http.Response) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- nil
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ch <- nil
		return
	}
	ch <- resp
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	brasilApiCh := make(chan *http.Response)
	viaCepCh := make(chan *http.Response)

	go fetchFromAPI(ctx, brasilApiURL+cep, brasilApiCh)
	go fetchFromAPI(ctx, viaCepURL+cep+"/json/", viaCepCh)

	select {
	case resp := <-brasilApiCh:
		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var address AddressBrasilAPI
			json.Unmarshal(body, &address)
			fmt.Printf("Response from BrasilAPI:\n%+v\n", address)
			return
		}
	case resp := <-viaCepCh:
		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var address AddressViaCep
			json.Unmarshal(body, &address)
			fmt.Printf("Response from ViaCEP:\n%+v\n", address)
			return
		}
	case <-ctx.Done():
		fmt.Println("Request timeout")
	}
}
