package components

import "fmt"

func CostMeter(tokens, costUSD float64) string {
	return fmt.Sprintf("Tokens: %.0f | Cost: $%.4f", tokens, costUSD)
}
