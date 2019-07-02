package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	ynab "github.com/rscarvalho/ynab-snapshot/client"
)

var specialGroupNames = map[string]bool{
	"Internal Master Category": true,
	"Hidden Categories":        true,
}

func main() {
	_ = godotenv.Load()

	year, month, day := time.Now().Date()
	snapshotMonth := ynab.CURRENT_MONTH

	token, ok := os.LookupEnv("YNAB_TOKEN")
	if !ok {
		token = ""
	}
	tokenPtr := flag.String("Token", token, "The ynab API token. Default: $YNAB_TOKEN")
	datePtr := flag.String("Date", snapshotMonth, "The month to take the snapshot in format YYYY-MM. Default: today's month")
	targetPath := flag.String("Path", "", "The target path for the snapshot. Default: The current path")

	flag.Parse()
	if len(*tokenPtr) == 0 {
		panic(fmt.Errorf("could not find api token in $YNAB_TOKEN "))
	}
	token = *tokenPtr

	snapshotMonth = *datePtr

	var fileName string
	if snapshotMonth == ynab.CURRENT_MONTH {
		fileName = fmt.Sprintf("%02d-%02d-%02d_CURRENT_category_snapshot.csv", year, month, day)
	} else {
		fileName = fmt.Sprintf("%02d-%02d-%02d_%s_category_snapshot.csv", year, month, day, snapshotMonth)
	}

	if len(*targetPath) > 0 {
		fileName = path.Join(*targetPath, fileName)
	}

	f, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(f)

	csvWriter := csv.NewWriter(w)
	defer func() {
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}()

	if err = csvWriter.Write([]string{
		"id", "budget_id", "group_id", "name", "budgeted", "balance",
	}); err != nil {
		panic(err)
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

		groups, err := ynabClient.GetCategories(budget.ID, snapshotMonth)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing response: %v", err)
		} else {
			for _, group := range groups {
				if group.Hidden || group.Deleted || specialGroupNames[group.Name] {
					continue
				}
				printCategoryGroup(&budget, &group, csvWriter)
			}
		}
	}
	_ = w.Flush()
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
		fmt.Printf("\t%s - %s (Budgeted %s)\n", category.Name, budget.CurrencyFormat.Format(category.Balance), budget.CurrencyFormat.Format(category.Budgeted))
	}
}

func writeRow(csvWriter *csv.Writer, budget *ynab.Budget, category *ynab.Category) error {
	record := []string{
		category.ID, budget.ID, category.CategoryGroupID, category.Name, strconv.FormatInt(category.Budgeted, 10), strconv.FormatInt(category.Balance, 10),
	}

	if err := csvWriter.Write(record); err != nil {
		return err
	}

	return nil
}
