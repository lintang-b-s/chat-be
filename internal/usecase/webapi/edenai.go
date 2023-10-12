package webapi

import (
	"io"
	"net/http"
	"strings"
)

// EdenAIAPI -.
type EdenAIAPI struct {
	apiKey string
}

// New -.
func NewEdenAIAPI(apiKey string) *EdenAIAPI {
	return &EdenAIAPI{
		apiKey: apiKey,
	}
}

func (api *EdenAIAPI) GenerateText(text string) string {
	url := "https://api.edenai.run/v2/text/generation"

	payload := strings.NewReader("{\"response_as_dict\":true,\"attributes_as_list\":false,\"show_original_response\":false,\"temperature\":0,\"max_tokens\":1000,\"providers\":\"ai21labs\",\"text\":" + "\"" + text + "\"" + "}")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Bearer "+api.apiKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	resText := string(body)

	return resText
}