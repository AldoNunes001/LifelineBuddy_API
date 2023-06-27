package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	GPT3URL    = "https://api.openai.com/v1/chat/completions"
	MAX_TOKENS = 500
	MODEL      = "gpt-3.5-turbo"
	ROLE       = "Seu nome é Buddy, você é um profissional especializado em evitar que pessoas cometam suicídios e me ajude a não cometer suicídio, não me mande procurar ajudar."
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GPT3Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	// MaxTokens int       `json:"max_tokens,omitempty"`
}

type GPT3Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func main() {
	router := gin.Default()

	router.POST("/buddy", askToBuddy)

	router.Run("localhost:8080")
}

func askToBuddy(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt := string(body)
	responseStr, err := makeGPT3Request(prompt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var response GPT3Response
	err = json.Unmarshal([]byte(responseStr), &response)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(responseStr)
	c.IndentedJSON(http.StatusOK, response.Choices[0].Message.Content)
}

func makeGPT3Request(prompt string) (string, error) {
	openAIKey := os.Getenv("OPENAI_KEY")
	if openAIKey == "" {
		return "", fmt.Errorf("OPENAI_KEY undefined")
	}

	request := GPT3Request{
		Model: MODEL,
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
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}
