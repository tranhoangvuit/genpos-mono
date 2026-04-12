package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance *validator.Validate
	once     sync.Once
)

// New returns a singleton validator instance.
func New() *validator.Validate {
	once.Do(func() {
		instance = validator.New()
	})
	return instance
}

// Validate validates a struct.
func Validate(i interface{}) error {
	return New().Struct(i)
}

// ValidateVar validates a single variable.
func ValidateVar(field interface{}, tag string) error {
	return New().Var(field, tag)
}
