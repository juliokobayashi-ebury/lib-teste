package documentvalidator

import (
	"sync"

	"github.com/juliokobayashi-ebury/lib-teste/documentvalidator/cnpjvalidator"
	"github.com/juliokobayashi-ebury/lib-teste/documentvalidator/cpfvalidator"
	"github.com/juliokobayashi-ebury/lib-teste/documentvalidator/documenthelpers"
)

var (
	once                      sync.Once
	documentValidatorInstance *DocumentValidator
)

type DocumentValidator struct {
	cpfValidator  cpfvalidator.CPFValidator
	cnpjValidator cnpjvalidator.CNPJValidator
}

func NewDocumentValidator() *DocumentValidator {
	once.Do(func() {
		documentValidatorInstance = &DocumentValidator{
			cpfValidator:  cpfvalidator.CPFValidator{},
			cnpjValidator: cnpjvalidator.CNPJValidator{},
		}
	})
	return documentValidatorInstance
}

func (ref *DocumentValidator) IsValid(data string) bool {
	data = documenthelpers.SanitizeDocument(data)

	if len(data) == 11 {
		return ref.cpfValidator.Validate(data)
	}
	if len(data) == 14 {
		return ref.cnpjValidator.Validate(data)
	}
	return false
}
