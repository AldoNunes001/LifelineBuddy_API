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
	ROLE       = "Seu nome é Buddy, você é um profissional especializado em evitar que pessoas cometam suicídios e me ajude a não cometer suicídio."
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
	body, err := readRequestBody(c)
	if err != nil {
		handleError(c, err, http.StatusBadRequest)
		return
	}

	responseStr, err := makeGPT3Request(string(body))
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	response, err := parseResponse(responseStr)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	if len(response.Choices) == 0 {
		handleError(c, fmt.Errorf("no choices in the response"), http.StatusInternalServerError)
	}
	c.IndentedJSON(http.StatusOK, response.Choices[0].Message.Content)
}

func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("request body is empty")
	}

	return body, nil
}

func handleError(c *gin.Context, err error, status int) {
	c.JSON(status, gin.H{"error": err.Error()})
}

func parseResponse(responseStr string) (*GPT3Response, error) {
	var response GPT3Response
	err := json.Unmarshal([]byte(responseStr), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}


// ...outros códigos...

func makeGPT3Request(prompt string) (string, error) {
	openAIKey, err := getOpenAIKey()
	if err != nil {
		return "", err
	}

	payload, err := preparePayload(prompt)
	if err != nil {
		return "", err
	}

	req, err := createRequest(payload, openAIKey)
	if err != nil {
		return "", err
	}

	resp, err := executeRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return readResponse(resp)
}

func getOpenAIKey() (string, error) {
	openAIKey := os.Getenv("OPENAI_KEY")
	if openAIKey == "" {
		return "", fmt.Errorf("OPENAI_KEY undefined")
	}
	return openAIKey, nil
}

func preparePayload(prompt string) ([]byte, error) {
	request := GPT3Request{
		Model: MODEL,
		Messages: []Message{
			{Role: "system", Content: ROLE},
			{Role: "user", Content: prompt},
		},
	}

	return json.Marshal(request)
}

func createRequest(payload []byte, openAIKey string) (*http.Request, error) {
	req, err := http.NewRequest("POST", GPT3URL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openAIKey)

	return req, nil
}

func executeRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func readResponse(resp *http.Response) (string, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}
