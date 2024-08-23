package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	BRASILAPI = "https://brasilapi.com.br/api/cep/v1/%s"
	VIACEP    = "http://viacep.com.br/ws/%s/json/"
)

type CEP struct {
	ZipCode      string `json:"cep"`
	Street       string `json:"logradouro"`
	Neighborhood string `json:"bairro"`
	City         string `json:"localidade"`
	State        string `json:"uf"`
}

type BrasilAPI struct {
	ZipCode      string `json:"cep"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cep := os.Args[1]
	if cep == "" {
		panic("cep is required")
	}
	cep = strings.ReplaceAll(cep, "-", "")
	if len(cep) != 8 {
		panic("cep is invalid")
	}

	// cep := "04689110"

	client := &http.Client{Timeout: time.Second * 1}

	chBrasilapi := make(chan CEP)
	brasilapi, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(BRASILAPI, cep), nil)
	if err != nil {
		panic(err)
	}

	chViacep := make(chan CEP)
	viacep, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(VIACEP, cep), nil)
	if err != nil {
		panic(err)
	}

	go getBrasilAPI(ctx, chBrasilapi, client, brasilapi)
	go getViaCEP(ctx, chViacep, client, viacep)

	var res any
	var api string
	select {
	case res = <-chBrasilapi:
		api = "brasilapi"
	case res = <-chViacep:
		api = "viacep"
	case <-time.After(time.Second * 100):
		panic("timeout")
	}
	fmt.Fprintf(os.Stdout, "Fastest API: %s\n", api)
	fmt.Fprintf(os.Stdout, "%#v\n", res)
	cancel()
}

func getViaCEP(_ context.Context, ch chan<- CEP, client *http.Client, req *http.Request) {
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var r CEP
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		panic(err)
	}

	ch <- r
}

func getBrasilAPI(_ context.Context, ch chan<- CEP, client *http.Client, req *http.Request) {
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var r BrasilAPI
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		panic(err)
	}
	addr := CEP{
		ZipCode:      r.ZipCode,
		Street:       r.Street,
		Neighborhood: r.Neighborhood,
		City:         r.City,
		State:        r.State,
	}

	ch <- addr
}
