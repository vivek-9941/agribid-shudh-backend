// Package gstin provides GSTIN validation logic.
package gstin

import (
	"fmt"
	"regexp"
	"strings"
)

// GSTIN format: 2-digit state code + 10-char PAN + 1 entity number + 'Z' + 1 checksum
var gstinRegex = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[A-Z0-9]{1}[Z]{1}[A-Z0-9]{1}$`)

// Valid state codes (01-37).
var validStateCodes = map[string]string{
	"01": "Jammu & Kashmir", "02": "Himachal Pradesh", "03": "Punjab",
	"04": "Chandigarh", "05": "Uttarakhand", "06": "Haryana",
	"07": "Delhi", "08": "Rajasthan", "09": "Uttar Pradesh",
	"10": "Bihar", "11": "Sikkim", "12": "Arunachal Pradesh",
	"13": "Nagaland", "14": "Manipur", "15": "Mizoram",
	"16": "Tripura", "17": "Meghalaya", "18": "Assam",
	"19": "West Bengal", "20": "Jharkhand", "21": "Odisha",
	"22": "Chhattisgarh", "23": "Madhya Pradesh", "24": "Gujarat",
	"26": "Dadra & Nagar Haveli", "27": "Maharashtra", "29": "Karnataka",
	"30": "Goa", "31": "Lakshadweep", "32": "Kerala",
	"33": "Tamil Nadu", "34": "Puducherry", "35": "Andaman & Nicobar",
	"36": "Telangana", "37": "Andhra Pradesh",
}

// Validate checks if a GSTIN is valid.
func Validate(gstin string) error {
	gstin = strings.ToUpper(strings.TrimSpace(gstin))

	if len(gstin) != 15 {
		return fmt.Errorf("GSTIN must be exactly 15 characters")
	}

	if !gstinRegex.MatchString(gstin) {
		return fmt.Errorf("GSTIN format is invalid")
	}

	stateCode := gstin[:2]
	if _, ok := validStateCodes[stateCode]; !ok {
		return fmt.Errorf("invalid state code: %s", stateCode)
	}

	return nil
}

// ExtractStateCode returns the 2-digit state code from a GSTIN.
func ExtractStateCode(gstin string) string {
	if len(gstin) >= 2 {
		return gstin[:2]
	}
	return ""
}

// GetStateName returns the state name for a state code.
func GetStateName(stateCode string) string {
	return validStateCodes[stateCode]
}
