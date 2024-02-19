package main

import (
	"context"
	"fmt"

	"github.com/vocdoni/go-airstack/airstack"
)

func main() {
	client := airstack.NewAirstackClient("your_api_key_here")

	variables := map[string]interface{}{
		"identity":   "wallet_address_here",
		"tokenType":  []string{"ERC20", "ERC721"}, // Example token types
		"blockchain": "ethereum",
		"limit":      10,
	}

	balances, err := client.GetTokenBalances(context.Background(), variables)
	if err != nil {
		fmt.Println("Error fetching token balances:", err)
		return
	}

	for _, balance := range balances {
		fmt.Printf("Token Address: %s, Amount: %s\n", balance.TokenAddress, balance.Amount)
	}
}
