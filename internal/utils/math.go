package utils

import (
	"fmt"
	"math/big"
	"strings"
)

// TokenMath provides utilities for token amount calculations
type TokenMath struct{}

// NewTokenMath creates a new TokenMath instance
func NewTokenMath() *TokenMath {
	return &TokenMath{}
}

// AddTokenAmounts adds two token amounts represented as strings
func (tm *TokenMath) AddTokenAmounts(amount1, amount2 string) (string, error) {
	if amount1 == "" {
		amount1 = "0"
	}
	if amount2 == "" {
		amount2 = "0"
	}

	// Convert to big integers
	big1, ok1 := new(big.Int).SetString(amount1, 10)
	big2, ok2 := new(big.Int).SetString(amount2, 10)
	
	if !ok1 {
		return "0", fmt.Errorf("invalid amount1: %s", amount1)
	}
	if !ok2 {
		return "0", fmt.Errorf("invalid amount2: %s", amount2)
	}

	// Add and return as string
	result := new(big.Int).Add(big1, big2)
	return result.String(), nil
}

// SubtractTokenAmounts subtracts amount2 from amount1
func (tm *TokenMath) SubtractTokenAmounts(amount1, amount2 string) (string, error) {
	if amount1 == "" {
		amount1 = "0"
	}
	if amount2 == "" {
		amount2 = "0"
	}

	// Convert to big integers
	big1, ok1 := new(big.Int).SetString(amount1, 10)
	big2, ok2 := new(big.Int).SetString(amount2, 10)
	
	if !ok1 {
		return "0", fmt.Errorf("invalid amount1: %s", amount1)
	}
	if !ok2 {
		return "0", fmt.Errorf("invalid amount2: %s", amount2)
	}

	// Subtract and return as string
	result := new(big.Int).Sub(big1, big2)
	
	// Ensure we don't return negative values for token amounts
	if result.Sign() < 0 {
		return "0", nil
	}
	
	return result.String(), nil
}

// MultiplyTokenAmount multiplies a token amount by a percentage (0.0 - 1.0)
func (tm *TokenMath) MultiplyTokenAmount(amount string, percentage float64) (string, error) {
	if amount == "" {
		amount = "0"
	}

	// Convert amount to big integer
	bigAmount, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "0", fmt.Errorf("invalid amount: %s", amount)
	}

	// Convert percentage to big rational (for precision)
	percentageStr := fmt.Sprintf("%.18f", percentage)
	percentageBig, ok := new(big.Rat).SetString(percentageStr)
	if !ok {
		return "0", fmt.Errorf("invalid percentage: %f", percentage)
	}

	// Convert amount to rational
	amountRat := new(big.Rat).SetInt(bigAmount)
	
	// Multiply
	result := new(big.Rat).Mul(amountRat, percentageBig)
	
	// Convert back to integer (truncate decimals)
	resultInt := new(big.Int).Div(result.Num(), result.Denom())
	
	return resultInt.String(), nil
}

// CompareTokenAmounts compares two token amounts
// Returns: -1 if amount1 < amount2, 0 if equal, 1 if amount1 > amount2
func (tm *TokenMath) CompareTokenAmounts(amount1, amount2 string) (int, error) {
	if amount1 == "" {
		amount1 = "0"
	}
	if amount2 == "" {
		amount2 = "0"
	}

	// Convert to big integers
	big1, ok1 := new(big.Int).SetString(amount1, 10)
	big2, ok2 := new(big.Int).SetString(amount2, 10)
	
	if !ok1 {
		return 0, fmt.Errorf("invalid amount1: %s", amount1)
	}
	if !ok2 {
		return 0, fmt.Errorf("invalid amount2: %s", amount2)
	}

	return big1.Cmp(big2), nil
}

// IsZero checks if a token amount is zero
func (tm *TokenMath) IsZero(amount string) bool {
	if amount == "" || amount == "0" {
		return true
	}
	
	bigAmount, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return true // Invalid amounts treated as zero
	}
	
	return bigAmount.Sign() == 0
}

// IsPositive checks if a token amount is positive (> 0)
func (tm *TokenMath) IsPositive(amount string) bool {
	if amount == "" {
		return false
	}
	
	bigAmount, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return false
	}
	
	return bigAmount.Sign() > 0
}

// SumTokenAmounts sums an array of token amounts
func (tm *TokenMath) SumTokenAmounts(amounts []string) (string, error) {
	total := "0"
	
	for _, amount := range amounts {
		newTotal, err := tm.AddTokenAmounts(total, amount)
		if err != nil {
			return "0", err
		}
		total = newTotal
	}
	
	return total, nil
}

// CalculatePercentage calculates what percentage amount1 is of amount2
// Returns percentage as float64 (0.0 - 1.0)
func (tm *TokenMath) CalculatePercentage(amount1, amount2 string) (float64, error) {
	if amount1 == "" {
		amount1 = "0"
	}
	if amount2 == "" || amount2 == "0" {
		return 0.0, nil // Division by zero returns 0%
	}

	// Convert to big integers
	big1, ok1 := new(big.Int).SetString(amount1, 10)
	big2, ok2 := new(big.Int).SetString(amount2, 10)
	
	if !ok1 {
		return 0.0, fmt.Errorf("invalid amount1: %s", amount1)
	}
	if !ok2 {
		return 0.0, fmt.Errorf("invalid amount2: %s", amount2)
	}

	// Convert to rationals for precise division
	rat1 := new(big.Rat).SetInt(big1)
	rat2 := new(big.Rat).SetInt(big2)
	
	// Calculate percentage
	result := new(big.Rat).Quo(rat1, rat2)
	
	// Convert to float64
	percentage, _ := result.Float64()
	
	return percentage, nil
}

// FormatTokenAmount formats a token amount for display (adds commas, etc.)
func (tm *TokenMath) FormatTokenAmount(amount string) string {
	if amount == "" || amount == "0" {
		return "0"
	}
	
	// Remove any existing commas or spaces
	cleaned := strings.ReplaceAll(amount, ",", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	
	// Validate it's a valid number
	if _, ok := new(big.Int).SetString(cleaned, 10); !ok {
		return amount // Return original if invalid
	}
	
	// Add commas for readability
	return addCommas(cleaned)
}

// addCommas adds comma separators to a numeric string
func addCommas(s string) string {
	// Handle negative numbers
	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}
	
	// Split into groups of 3 from right to left
	result := ""
	for i, digit := range reverse(s) {
		if i > 0 && i%3 == 0 {
			result = "," + result
		}
		result = string(digit) + result
	}
	
	if negative {
		result = "-" + result
	}
	
	return result
}

// reverse reverses a string
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Global instance for convenience
var GlobalTokenMath = NewTokenMath()