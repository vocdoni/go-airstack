package airstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Constants
const (
	apiEndpointProd           = "https://api.airstack.xyz/gql"
	apiTimeout                = 60 * time.Second
	successStatusCode         = 200
	unprocessableEntityStatus = 422
)

// SendRequest handles HTTP requests to the Airstack API.
func SendRequest(ctx context.Context, method, url string, headers map[string]string, body []byte) (response []byte, statusCode int, err error) {
	client := &http.Client{Timeout: apiTimeout}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	response, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	statusCode = resp.StatusCode
	if statusCode != successStatusCode {
		err = json.Unmarshal(response, &map[string]interface{}{})
		if err != nil {
			// Handle JSON parse error
			return response, statusCode, err
		}
	}

	return response, statusCode, nil
}

// AirstackClient manages the API client for Airstack.
type AirstackClient struct {
	APIKey string
	URL    string
}

// NewAirstackClient initializes a new Airstack client.
func NewAirstackClient(apiKey string) *AirstackClient {
	return &AirstackClient{
		APIKey: apiKey,
		URL:    apiEndpointProd,
	}
}

// TokenBalance represents the structure of a token balance response.
type TokenBalance struct {
	Amount          string `json:"amount"`
	FormattedAmount string `json:"formattedAmount"`
	Blockchain      string `json:"blockchain"`
	TokenAddress    string `json:"tokenAddress"`
	TokenId         string `json:"tokenId"`
	// Include other fields as needed
}

// GetTokenBalances queries for token balances with given parameters.
func (client *AirstackClient) GetTokenBalances(ctx context.Context, variables map[string]interface{}) ([]TokenBalance, error) {
	query := `
	query GetTokensHeldByWalletAddress($identity: Identity, $tokenType: [TokenType!], $blockchain: TokenBlockchain!, $limit: Int) {
		TokenBalances(
			input: {filter: {owner: {_eq: $identity}, tokenType: {_in: $tokenType}}, blockchain: $blockchain, limit: $limit}
		) {
			TokenBalance {
				amount
				formattedAmount
				blockchain
				tokenAddress
				tokenId
				// Include other fields as needed
			}
		}
	}
	`

	resp, err := client.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	// Parsing part of the response into the structure we defined above.
	var respData struct {
		TokenBalances struct {
			TokenBalance []TokenBalance `json:"TokenBalance"`
		} `json:"TokenBalances"`
	}

	if err := json.Unmarshal(resp.Data, &respData); err != nil {
		return nil, err
	}

	return respData.TokenBalances.TokenBalance, nil
}

// QueryResponse holds the GraphQL query response structure.
type QueryResponse struct {
	Data         json.RawMessage
	StatusCode   int
	Error        string
	HasNextPage  bool
	HasPrevPage  bool
	NextPageFunc func() (*QueryResponse, error)
	PrevPageFunc func() (*QueryResponse, error)
}

// ExecuteQuery sends a GraphQL query to the Airstack API and returns the parsed response.
func (client *AirstackClient) ExecuteQuery(ctx context.Context, query string, variables map[string]interface{}) (*QueryResponse, error) {
	body, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": client.APIKey,
	}

	response, statusCode, err := SendRequest(ctx, "POST", client.URL, headers, body)
	if err != nil || statusCode != successStatusCode {
		return &QueryResponse{
			StatusCode: statusCode,
			Error:      fmt.Sprintf("HTTP error: %s, Status Code: %d", err, statusCode),
		}, nil
	}

	var respData map[string]json.RawMessage
	if err := json.Unmarshal(response, &respData); err != nil {
		return nil, err
	}

	// Check for "errors" field in response JSON
	if errorField, ok := respData["errors"]; ok {
		return &QueryResponse{
			Data:       nil,
			StatusCode: statusCode,
			Error:      string(errorField),
		}, nil
	}

	// Here you would handle pagination based on the response structure,
	// setting HasNextPage, HasPrevPage, NextPageFunc, and PrevPageFunc as needed.

	return &QueryResponse{
		Data:       respData["data"],
		StatusCode: statusCode,
	}, nil
}

// ExecutePaginatedQuery would be implemented here, focusing on handling pagination logic,
// including setting up NextPageFunc and PrevPageFunc callbacks.

// For simplicity, the detailed implementation of pagination handling is omitted,
// but you would typically parse the `pageInfo` from the response to set up these callbacks.
