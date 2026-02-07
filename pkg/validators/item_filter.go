package validators

import (
	"github.com/go-playground/validator/v10"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ValidItemFilter returns whether the given item filter is valid
func ValidItemFilter(fl validator.FieldLevel) bool {
	if value, ok := fl.Field().Interface().(string); ok {
		if value == "" {
			return true
		}

		if value == models.TransactionNoItemFilterValue {
			return true
		}

		_, err := models.ParseTransactionItemFilter(value)

		return err == nil
	}

	return false
}
