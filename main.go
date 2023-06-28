package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"LifelineBuddy_API/openai"

	"github.com/gin-gonic/gin"
)

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

	responseStr, err := openai.MakeGPT3Request(string(body))
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
	// c.IndentedJSON(http.StatusOK, response.Choices[0].Message)
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
