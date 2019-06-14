package main

import (
	"fmt"
	ynab "github.com/rscarvalho/YNABSnapshot/client"
	"os"
	"strconv"
	"strings"
)

func main() {
	token, ok := os.LookupEnv("YNAB_TOKEN")

	if !ok {
		panic(fmt.Errorf("could not find api token in $YNAB_TOKEN"))
	}

	ynabClient := ynab.NewClient(token)
	budgets, err := ynabClient.GetBudgets()
	if err != nil {
		panic(err)
	}

	fmt.Println("Budgets:")
	for _, budget := range budgets {
		msg := fmt.Sprintf("Budget: %s (decimal digits=%d)", budget.Name, budget.CurrencyFormat.DecimalDigits)
		fmt.Printf("%s\n%s\n\n", msg, strings.Repeat("-", len(msg)))

		groups, err := ynabClient.GetCategories(budget.Id)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing response: %v", err)
		} else {
			for _, group := range groups {
				printCategoryGroup(&budget, &group)
			}
		}
	}
}

func printCategoryGroup(budget *ynab.Budget, group *ynab.CategoryGroup) {
	fmt.Printf("%s:\n", group.Name)

	for _, category := range group.Categories {
		if category.Budgeted == 0 && category.Balance == 0 {
			continue
		}
		fmt.Printf("\t%s - %s (Budgeted %s)\n", category.Name, formatCurrency(category.Balance, budget.CurrencyFormat), formatCurrency(category.Budgeted, budget.CurrencyFormat))
	}
}

func formatCurrency(number int64, format ynab.CurrencyFormat) string {
	var b strings.Builder

	millis := 3
	if number < 0 {
		millis = 4
	}

	if format.DisplaySymbol && format.SymbolFirst {
		b.WriteString(format.CurrencySymbol)
	}

	if number > 0 {
		numberStr := strconv.FormatInt(number, 10)
		before := numberStr[:len(numberStr)-millis]
		after := numberStr[len(numberStr)-millis:]
		after = after[:format.DecimalDigits]

		groups := make([]string, 0)
		remainder := ""
		if before[0] == '-' {
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
	} else {
		b.WriteString(fmt.Sprintf("0%s%s", format.DecimalSeparator, strings.Repeat("0", format.DecimalDigits)))
	}

	if format.DisplaySymbol && !format.SymbolFirst {
		b.WriteString(" ")
		b.WriteString(format.CurrencySymbol)
	}

	return b.String()
}
