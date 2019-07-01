package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	ynab "github.com/rscarvalho/ynab-snapshot/client"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var SpecialGroupNames = map[string]bool{
	"Internal Master Category": true,
	"Hidden Categories":        true,
}

func main() {
	fmt.Printf("Running with args: %+v\n", os.Args)
	year, month, day := time.Now().Date()
	fileName := fmt.Sprintf("%02d-%02d-%02d_category_snapshot.csv", year, month, day)

	if len(os.Args) > 1 {
		fileName = path.Join(os.Args[1], fileName)
	}

	f, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(f)

	csvWriter := csv.NewWriter(w)
	defer f.Close()

	if err = csvWriter.Write([]string{
		"id", "budget_id", "group_id", "name", "budgeted", "balance",
	}); err != nil {
		panic(err)
	}

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
				if group.Hidden || group.Deleted || SpecialGroupNames[group.Name] {
					continue
				}
				printCategoryGroup(&budget, &group, csvWriter)
			}
		}
	}
	w.Flush()
}

func printCategoryGroup(budget *ynab.Budget, group *ynab.CategoryGroup, csvWriter *csv.Writer) {
	fmt.Printf("%s:\n", group.Name)

	for _, category := range group.Categories {

		if category.Hidden || category.Deleted || category.Budgeted == 0 && category.Balance == 0 {
			continue
		}
		err := writeRow(csvWriter, budget, &category)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error writing to csv: %v", err)
		}
		fmt.Printf("\t%s - %s (Budgeted %s)\n", category.Name, formatCurrency(category.Balance, budget.CurrencyFormat), formatCurrency(category.Budgeted, budget.CurrencyFormat))
	}
}

func writeRow(csvWriter *csv.Writer, budget *ynab.Budget, category *ynab.Category) error {
	record := []string{
		category.Id, budget.Id, category.CategoryGroupId, category.Name, strconv.FormatInt(category.Budgeted, 10), strconv.FormatInt(category.Balance, 10),
	}

	if err := csvWriter.Write(record); err != nil {
		return err
	}

	return nil
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
