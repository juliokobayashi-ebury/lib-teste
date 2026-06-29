package documentvalidator

import (
	"bexstech/spb-str/internal/platform/documentvalidator/cnpjvalidator"
	"bexstech/spb-str/internal/platform/documentvalidator/cpfvalidator"
	"bexstech/spb-str/internal/platform/documentvalidator/documenthelpers"
)

type DocumentValidator struct{
	cpfValidator cpfvalidator.CPFValidator
	cnpjValidator cnpjvalidator.CNPJValidator
}

func (ref DocumentValidator) IsValid(data string) bool {
	data = documenthelpers.SanitizeDocument(data)
	
	if len(data) == 11 {
		return ref.cpfValidator.Validate(data)
	} 
	if len(data) == 14 {
		return ref.cnpjValidator.Validate(data)
	}
	return false
}


