package invoice

import (
	"context"
	"fmt"
	"math/big"
)

// DetermineGSTType determines whether a transaction is intra-state (CGST+SGST)
// or inter-state (IGST) based on seller and buyer state codes.
func DetermineGSTType(sellerStateCode, buyerStateCode string) GSTType {
	if sellerStateCode == buyerStateCode {
		return GSTTypeIntra
	}
	return GSTTypeInter
}

// CalculateLineTax computes the GST breakdown for a single line item.
// All calculations use big.Rat for precision (avoids float64).
func CalculateLineTax(taxableAmountStr string, hsnRate *HSNTaxRate, gstType GSTType) (*TaxBreakdown, error) {
	taxable, ok := new(big.Rat).SetString(taxableAmountStr)
	if !ok {
		return nil, fmt.Errorf("invalid taxable amount: %s", taxableAmountStr)
	}

	hundred := new(big.Rat).SetInt64(100)
	breakdown := &TaxBreakdown{
		TaxableAmount: taxableAmountStr,
	}

	if gstType == GSTTypeIntra {
		// CGST + SGST
		cgstRate, _ := new(big.Rat).SetString(hsnRate.CGSTRate)
		sgstRate, _ := new(big.Rat).SetString(hsnRate.SGSTRate)

		cgstAmt := new(big.Rat).Mul(taxable, cgstRate)
		cgstAmt.Quo(cgstAmt, hundred)

		sgstAmt := new(big.Rat).Mul(taxable, sgstRate)
		sgstAmt.Quo(sgstAmt, hundred)

		breakdown.CGSTRate = hsnRate.CGSTRate
		breakdown.CGSTAmount = formatDecimal(cgstAmt)
		breakdown.SGSTRate = hsnRate.SGSTRate
		breakdown.SGSTAmount = formatDecimal(sgstAmt)
		breakdown.IGSTRate = "0.00"
		breakdown.IGSTAmount = "0.00"

		totalTax := new(big.Rat).Add(cgstAmt, sgstAmt)
		breakdown.TotalTax = formatDecimal(totalTax)
	} else {
		// IGST
		igstRate, _ := new(big.Rat).SetString(hsnRate.IGSTRate)

		igstAmt := new(big.Rat).Mul(taxable, igstRate)
		igstAmt.Quo(igstAmt, hundred)

		breakdown.CGSTRate = "0.00"
		breakdown.CGSTAmount = "0.00"
		breakdown.SGSTRate = "0.00"
		breakdown.SGSTAmount = "0.00"
		breakdown.IGSTRate = hsnRate.IGSTRate
		breakdown.IGSTAmount = formatDecimal(igstAmt)

		breakdown.TotalTax = formatDecimal(igstAmt)
	}

	// Cess (applied regardless of GST type)
	cessRate, _ := new(big.Rat).SetString(hsnRate.CessRate)
	if cessRate != nil && cessRate.Sign() > 0 {
		cessAmt := new(big.Rat).Mul(taxable, cessRate)
		cessAmt.Quo(cessAmt, hundred)
		breakdown.CessRate = hsnRate.CessRate
		breakdown.CessAmount = formatDecimal(cessAmt)

		totalTax, _ := new(big.Rat).SetString(breakdown.TotalTax)
		totalTax.Add(totalTax, cessAmt)
		breakdown.TotalTax = formatDecimal(totalTax)
	} else {
		breakdown.CessRate = "0.00"
		breakdown.CessAmount = "0.00"
	}

	return breakdown, nil
}

// formatDecimal formats a big.Rat to a 2-decimal-place string.
func formatDecimal(r *big.Rat) string {
	f, _ := r.Float64()
	return fmt.Sprintf("%.2f", f)
}

type HSNStubService struct{}

func NewHSNStubService() *HSNStubService {
	return &HSNStubService{}
}

func (s *HSNStubService) GetHSNRate(ctx context.Context, hsnCode string) (*HSNTaxRate, error) {
	// Stub implementation returning generic 18% GST for anything
	return &HSNTaxRate{
		HSNCode: hsnCode,
		IGSTRate: "18.00",
		CGSTRate: "9.00",
		SGSTRate: "9.00",
		CessRate: "0.00",
	}, nil
}
