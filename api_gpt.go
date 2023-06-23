package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	GPT3URL    = "https://api.openai.com/v1/completions"
	MODEL      = "text-davinci-003"
	MAX_TOKENS = 500
	ROLE       = "Aja como um profissional especializado em evitar que pessoas cometam suicídios."
)

type GPT3Request struct {
	Prompt    string `json:"prompt"`
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/buddy", buddyHandler)

	fmt.Println("Servidor iniciando na porta 8080")
	http.ListenAndServe(":8080", nil)
}

// Rest of the code...

func buddyHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		w.Write([]byte("Welcome to Lifeline Buddy!"))

	case http.MethodPost:
		// Ler o corpo da requisição
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Erro ao ler o corpo da requisição", http.StatusInternalServerError)
			return
		}

		// Fechar o corpo da requisição depois de ler
		defer r.Body.Close()

		// Chamada à API GPT-3
		prompt := string(body)
		response, err := makeGPT3Request(prompt)
		if err != nil {
			http.Error(w, "Erro ao chamar a API GPT-3", http.StatusInternalServerError)
			return
		}

		// Responder com o resultado da chamada à API GPT-3
		w.Write([]byte(response))

	default:
		http.Error(w, "Método de requisição inválido", http.StatusMethodNotAllowed)
	}
}

func makeGPT3Request(prompt string) (string, error) {
	// Recuperar a chave da API da variável de ambiente
	openAIKey := os.Getenv("OPENAI_KEY")
	if openAIKey == "" {
		return "", fmt.Errorf("OPENAI_KEY não definido")
	}

	// Adicionar o papel do profissional especializado ao prompt
	prompt = ROLE + "\n" + prompt

	// Criar a requisição para o GPT-3
	reqBody, err := json.Marshal(GPT3Request{Prompt: prompt, Model: MODEL, MaxTokens: MAX_TOKENS})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", GPT3URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	// Adicionar cabeçalhos necessários
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openAIKey)

	// Fazer a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Ler a resposta
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}
