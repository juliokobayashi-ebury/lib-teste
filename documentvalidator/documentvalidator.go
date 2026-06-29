package documentvalidator

import (
	"github.com/juliokobayashi-ebury/lib-teste/documentvalidator/cnpjvalidator"
	"github.com/juliokobayashi-ebury/lib-teste/documentvalidator/cpfvalidator"
	"github.com/juliokobayashi-ebury/lib-teste/documentvalidator/documenthelpers"
)

type DocumentValidator struct {
	cpfValidator  cpfvalidator.CPFValidator
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
