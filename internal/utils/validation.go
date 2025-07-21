package utils

import (
	"regexp"
	"strings"
)

// IsValidEthereumAddress validates if a string is a valid Ethereum address
func IsValidEthereumAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	
	// Check if all characters after 0x are valid hex characters
	hexRegex := regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)
	return hexRegex.MatchString(address)
}

// IsValidTransactionHash validates if a string is a valid Ethereum transaction hash
func IsValidTransactionHash(hash string) bool {
	if len(hash) != 66 {
		return false
	}
	
	if !strings.HasPrefix(hash, "0x") {
		return false
	}
	
	// Check if all characters after 0x are valid hex characters
	hexRegex := regexp.MustCompile(`^0x[a-fA-F0-9]{64}$`)
	return hexRegex.MatchString(hash)
}

// NormalizeAddress converts an Ethereum address to lowercase
func NormalizeAddress(address string) string {
	return strings.ToLower(address)
}