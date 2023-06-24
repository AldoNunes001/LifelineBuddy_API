package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	GPT3URL    = "https://api.openai.com/v1/chat/completions"
	MAX_TOKENS = 500
	ROLE = "Você é um profissional especializado em evitar que pessoas cometam suicídios e me ajude a não cometer suicídio."
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GPT3Request struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	// MaxTokens int       `json:"max_tokens,omitempty"`
}

type GPT3Response struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Created int64 `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text          string      `json:"text"`
		Index         int         `json:"index"`
		Logprobs      interface{} `json:"logprobs"`
		FinishReason  string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens      int `json:"prompt_tokens"`
		CompletionTokens  int `json:"completion_tokens"`
		TotalTokens       int `json:"total_tokens"`
	} `json:"usage"`
}

func main() {
	http.HandleFunc("/buddy", buddyHandler)

	fmt.Println("Servidor iniciando na porta 8080")
	http.ListenAndServe(":8080", nil)
}

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

	// Preparar a requisição para o GPT-3
	request := GPT3Request{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "system", Content: ROLE},
			{Role: "user", Content: prompt},
		},
	}

	// Criar a requisição para o GPT-3
	reqBody, err := json.Marshal(request)
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
