package client

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// CurrencyFormat has information about how to format a currency on the screen
type CurrencyFormat struct {
	IsoCode          string `json:"iso_code"`
	ExampleFormat    string `json:"example_format"`
	DecimalDigits    int    `json:"decimal_digits"`
	DecimalSeparator string `json:"decimal_separator"`
	SymbolFirst      bool   `json:"symbol_first"`
	GroupSeparator   string `json:"group_separator"`
	CurrencySymbol   string `json:"currency_symbol"`
	DisplaySymbol    bool   `json:"display_symbol"`
}

// Budget represents a YNAB budget
type Budget struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LastModifiedOn string `json:"last_modified_on"`
	FirstMonth     string `json:"first_month"`
	LastMonth      string `json:"last_month"`
	DateFormat     struct {
		Format string `json:"format"`
	} `json:"date_format"`
	CurrencyFormat CurrencyFormat `json:"currency_format"`
}

// Category represents a YNAB category
type Category struct {
	ID                      string `json:"id"`
	CategoryGroupID         string `json:"category_group_id"`
	Name                    string `json:"name"`
	Hidden                  bool   `json:"hidden"`
	OriginalCategoryGroupID string `json:"original_category_group_id"`
	Note                    string `json:"note"`
	Budgeted                int64  `json:"budgeted"`
	Activity                int64  `json:"activity"`
	Balance                 int64  `json:"balance"`
	GoalType                string `json:"goal_type"`
	GoalCreationMonth       string `json:"goal_creation_month"`
	GoalTarget              int64  `json:"goal_target"`
	GoalTargetMonth         string `json:"goal_target_month"`
	GoalPercentageComplete  int64  `json:"goal_percentage_complete"`
	Deleted                 bool   `json:"deleted"`
}

// CategoryGroup represents a group of categories in YNAB
type CategoryGroup struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Hidden     bool       `json:"hidden"`
	Deleted    bool       `json:"deleted"`
	Categories []Category `json:"categories"`
}

func copyCategories(dst, src *Category) {
	valueDst := reflect.ValueOf(dst)
	valueSrc := reflect.ValueOf(src)

	for i := 0; i < valueDst.Elem().NumField(); i++ {
		valueDst.Elem().Field(i).Set(valueSrc.Elem().Field(i))
	}
}

// Format formats currency microvalue `number` according to the currency `format`
func (format *CurrencyFormat) Format(number int64) string {
	var b strings.Builder

	millis := 3
	if number < 0 {
		millis = 4
	}

	if number < 0 {
		result := format.Format(-number)
		return fmt.Sprintf("-%s", result)
	}

	numberStr := strconv.FormatInt(number, 10)
	before := numberStr[:len(numberStr)-millis]
	after := numberStr[len(numberStr)-millis:]
	after = after[:format.DecimalDigits]

	groups := make([]string, 0)
	remainder := ""

	if len(before) == 0 {
		before = "0"
	} else if before[0] == '-' {
		remainder = "-"
		before = before[1:]
	}

	for true {
		if len(before) <= 3 {
			groups = append(groups, before)
			break
		} else {
			groups = append(groups, before[len(before)-3:])
			before = before[:len(before)-3]
		}
	}

	b.WriteString(remainder)
	for i := len(groups) - 1; i >= 0; i-- {
		b.WriteString(groups[i])

		if i != 0 {
			b.WriteString(format.GroupSeparator)
		}
	}

	b.WriteString(format.DecimalSeparator)
	b.WriteString(after)

	var result string

	if format.DisplaySymbol && !format.SymbolFirst {
		b.WriteString(" ")
		b.WriteString(format.CurrencySymbol)
		result = b.String()
	} else if format.DisplaySymbol {
		result = fmt.Sprintf("%s%s", format.CurrencySymbol, b.String())
	}

	return result
}
