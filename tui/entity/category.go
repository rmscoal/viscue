package entity

import (
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Category struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

// FilterValue implements list.Item
func (category Category) FilterValue() string { return category.Name }

// Title implements list.DefaultItem
func (category Category) Title() string { return category.Name }

// Description implements list.DefaultItem
func (category Category) Description() string { return "" }

func (category Category) Validate() error {
	err := validation.ValidateStruct(&category,
		validation.Field(&category.Name, validation.Required),
	)
	if err != nil {
		msg := strings.Split(err.Error(), "; ")
		return fmt.Errorf(strings.Join(msg, " and "))
	}

	return nil
}
