package main

import (
	"fmt"
	"log"
	"sort"
)

type PnLResult struct {
	incomeTotal     float64
	expenseTotal    float64
	investmentTotal float64
	pnlAmount       float64
	pnlPercent      float64
}

// calculates the p&l for a specific month - does not include investments
func calculateMonthPnL(month, year string) (PnLResult, error) {
	var pnl PnLResult

	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return pnl, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	for txType, txList := range transactions[year][month] {
		if len(txList) == 0 {
			log.Printf("\nno transactions of type %s for %s %s\n", txType, month, year)
			continue
		}

		for _, tx := range txList {

			if txType == "income" {
				pnl.incomeTotal += tx.Amount
			}

			if txType == "expense" {
				pnl.expenseTotal += tx.Amount
			}

			if txType == "investment" {
				pnl.investmentTotal += tx.Amount
			}
		}
	}

	// calucalte the reulting P&L: income - expenses
	// in absolute value and in % savings i.e. if you receved 1000 and spent 400, this will be 60% savings rate (investments are not calculated as either income or expense, they are their separate category that does not factor into calculating the savings rate)
	if pnl.incomeTotal == 0 { // avoid division by zero
		pnl.pnlAmount = pnl.incomeTotal - pnl.expenseTotal
		pnl.pnlPercent = 0
	} else {
		pnl.pnlAmount = pnl.incomeTotal - pnl.expenseTotal
		pnl.pnlPercent = ((pnl.incomeTotal - pnl.expenseTotal) / pnl.incomeTotal) * 100
	}

	return pnl, nil
}

// calculates the p&l for a specific year - does not include investments
func calculateYearPnL(year string) (PnLResult, error) {
	var pnl PnLResult

	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return pnl, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	for month := range transactions[year] {
		for txType, txList := range transactions[year][month] {
			if len(txList) == 0 {
				log.Printf("\nno transactions of type %s for month %s\n", txType, month)
				continue
			}

			for _, tx := range txList {

				if txType == "income" {
					pnl.incomeTotal += tx.Amount
				}

				if txType == "expense" {
					pnl.expenseTotal += tx.Amount
				}

				if txType == "investment" {
					pnl.investmentTotal += tx.Amount
				}
			}
		}
	}

	// calucalte the reulting P&L: income - expenses
	// in absolute value and in % savings i.e. if you receved 1000 and spent 400, this will be 60% savings rate (investments are not calculated as either income or expense, they are their separate category that does not factor into calculating the savings rate)
	if pnl.incomeTotal == 0 { // avoid division by zero
		pnl.pnlAmount = pnl.incomeTotal - pnl.expenseTotal
		pnl.pnlPercent = 0
	} else {
		pnl.pnlAmount = pnl.incomeTotal - pnl.expenseTotal
		pnl.pnlPercent = ((pnl.incomeTotal - pnl.expenseTotal) / pnl.incomeTotal) * 100
	}

	return pnl, nil
}

// calculates the p&l for each month in a specific year
func calculateYearMonthlyPnL(year string) (map[string]PnLResult, error) {
	monthlyPnL := make(map[string]PnLResult)

	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return monthlyPnL, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	for month := range transactions[year] {
		pnl, err := calculateMonthPnL(month, year)
		if err != nil {
			return monthlyPnL, fmt.Errorf("unable to calculate pnl for %s %s: %w", month, year, err)
		}
		monthlyPnL[month] = pnl
	}

	return monthlyPnL, nil
}

// total spent on a single expense category and its share of all expenses in the period
type CategoryTotal struct {
	category string
	total    float64
	percent  float64 // % of total expenses for the period
}

// expense breakdown per category for a period (month or year)
type CategoryBreakdown struct {
	expenseTotal float64
	categories   []CategoryTotal // sorted by total, highest first
}

// builds a per-category expense breakdown for the given months of a year
// pure function over TransactionHistory so it can be reused for both month and year views
func buildExpenseCategoryBreakdown(transactions TransactionHistory, year string, months []string) CategoryBreakdown {
	var breakdown CategoryBreakdown
	totals := make(map[string]float64)

	for _, month := range months {
		for _, tx := range transactions[year][month]["expense"] {
			totals[tx.Category] += tx.Amount
			breakdown.expenseTotal += tx.Amount
		}
	}

	for category, total := range totals {
		ct := CategoryTotal{category: category, total: total}
		if breakdown.expenseTotal != 0 { // avoid division by zero
			ct.percent = (total / breakdown.expenseTotal) * 100
		}
		breakdown.categories = append(breakdown.categories, ct)
	}

	// highest spend first, alphabetical on ties so the order is stable between renders
	sort.Slice(breakdown.categories, func(i, j int) bool {
		if breakdown.categories[i].total != breakdown.categories[j].total {
			return breakdown.categories[i].total > breakdown.categories[j].total
		}
		return breakdown.categories[i].category < breakdown.categories[j].category
	})

	return breakdown
}

// calculates how much was spent on each expense category in a specific month
func calculateMonthCategoryBreakdown(month, year string) (CategoryBreakdown, error) {
	transactions, err := LoadTransactions()
	if err != nil {
		return CategoryBreakdown{}, fmt.Errorf("unable to load transactions file: %w", err)
	}

	return buildExpenseCategoryBreakdown(transactions, year, []string{month}), nil
}

// calculates how much was spent on each expense category across a whole year
func calculateYearCategoryBreakdown(year string) (CategoryBreakdown, error) {
	transactions, err := LoadTransactions()
	if err != nil {
		return CategoryBreakdown{}, fmt.Errorf("unable to load transactions file: %w", err)
	}

	months := make([]string, 0, len(transactions[year]))
	for month := range transactions[year] {
		months = append(months, month)
	}

	return buildExpenseCategoryBreakdown(transactions, year, months), nil
}
