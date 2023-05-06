package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/imroc/req"
)

const (
	OPENAI_COMPLETIONS_API_URL_BETA = "https://api.openai.com/v1/chat/completions"
	OPENAI_MODELS_API_URL           = "https://api.openai.com/v1/models"
)

type ListResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

type Model struct {
	Id         string      `json:"id"`
	Object     string      `json:"object"`
	Created    int         `json:"created"`
	OwnedBy    string      `json:"owned_by"`
	Permission []ModelPerm `json:"permission"`
	Root       string      `json:"root"`
	Parent     *Model      `json:"parent,omitempty"`
}

type ModelPerm struct {
	Id                 string `json:"id"`
	Object             string `json:"object"`
	Created            int    `json:"created"`
	AllowCreateEngine  bool   `json:"allow_create_engine"`
	AllowSampling      bool   `json:"allow_sampling"`
	AllowLogprobs      bool   `json:"allow_logprobs"`
	AllowSearchIndices bool   `json:"allow_search_indices"`
	AllowView          bool   `json:"allow_view"`
	AllowFineTuning    bool   `json:"allow_fine_tuning"`
	Organization       string `json:"organization"`
	Group              string `json:"group,omitempty"`
	IsBlocking         bool   `json:"is_blocking"`
}

type ChatCompletion struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

func ChatGPTChatCompletion(pastMessages []map[string]string, query string) (ChatCompletion, error) {
	apiKey := chatGPTAPITOKEN
	if apiKey == "" {
		return ChatCompletion{}, fmt.Errorf("no OPENAI_API_KEY provided")
	}

	header := req.Header{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + apiKey,
	}

	messages := append(pastMessages, map[string]string{"role": "user", "content": query})
	messagesTokens := CalculateTokens(messages)
	maxTokens := MAX_COMPLETION_CHARS - messagesTokens

	if maxTokens <= 20 {
		return ChatCompletion{}, fmt.Errorf("MAX TOKENS SURPASSED: " + strconv.Itoa(messagesTokens))
	}

	body := req.BodyJSON(map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"max_tokens":  maxTokens,
		"temperature": 0,
		"messages":    messages,
	})

	// fmt.Println("Request body:", body)
	r, err := req.Post(OPENAI_COMPLETIONS_API_URL_BETA, header, body)
	if err != nil {
		return ChatCompletion{}, err
	}

	// fmt.Println("Response body:", r)
	var chatCompletionError ErrorResponse
	if err := json.Unmarshal(r.Bytes(), &chatCompletionError); err != nil {
		return ChatCompletion{}, fmt.Errorf("error parsing ChatGPT API error response: %s", err)
	}
	if chatCompletionError.Error.Message != "" {
		return ChatCompletion{}, fmt.Errorf(chatCompletionError.Error.Message)
	}

	var chatCompletion ChatCompletion
	if err := json.Unmarshal(r.Bytes(), &chatCompletion); err != nil {
		return ChatCompletion{}, fmt.Errorf("error parsing ChatGPT API response: %s", err)
	}

	// fmt.Println("Response body:", chatCompletion)

	return chatCompletion, nil
}

func ChatGPTQuery(query string) (ChatCompletion, error) {
	apiKey := chatGPTAPITOKEN
	if apiKey == "" {
		return ChatCompletion{}, fmt.Errorf("no OPENAI_API_KEY provided")
	}

	header := req.Header{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + apiKey,
	}

	messages := []map[string]string{
		{"role": "user", "content": query},
	}

	body := req.BodyJSON(map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"max_tokens":  4096,
		"temperature": 0,
		"messages":    messages,
	})
	fmt.Println("Request body:", body)
	r, err := req.Post(OPENAI_COMPLETIONS_API_URL_BETA, header, body)
	if err != nil {
		return ChatCompletion{}, err
	}

	// Read the response body
	var chatCompletion ChatCompletion
	if err := json.Unmarshal(r.Bytes(), &chatCompletion); err != nil {
		return ChatCompletion{}, fmt.Errorf("error parsing ChatGPT API response: %s", err)
	}

	// fmt.Println("Response body:", chatCompletion)

	return chatCompletion, nil
}

func CalculateTokens(messages []map[string]string) int {
	// Now it is aproximated to words, but tokens are evaluated differently
	// depending on the model
	return calculateWords(messages)
}

func ListModels() ([]Model, error) {
	apiKey := chatGPTAPITOKEN
	if apiKey == "" {
		return []Model{}, fmt.Errorf("no OPENAI_API_KEY provided to list models")
	}

	header := req.Header{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + apiKey,
	}

	r, err := req.Get(OPENAI_MODELS_API_URL, header)
	if err != nil {
		return []Model{}, fmt.Errorf("error requesting models.")
	}
	var modelsResponse ListResponse
	err = json.Unmarshal(r.Bytes(), &modelsResponse)
	if err != nil {
		log.Fatal(err)
		return []Model{}, fmt.Errorf("error marshalling response from list models.")
	}
	return modelsResponse.Data, nil
}
