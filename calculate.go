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

// spend on a category in one month next to its average monthly spend over the year
type CategoryComparison struct {
	CategoryTotal         // this month's total and share of this month's expenses
	monthlyAvg    float64 // average monthly spend on the category across the year's months
}

// a month's category breakdown compared against the monthly averages of its year
type MonthComparison struct {
	month           CategoryBreakdown    // the month itself, categories sorted by spend (drives the pie chart)
	rows            []CategoryComparison // month categories first (by spend), then categories only spent in other months (by average)
	monthsInYear    int                  // number of months with transactions that the averages are based on
	avgExpenseTotal float64              // average total monthly expenses across the year
}

// compares each category's spend in the month with its average monthly spend in the same year
// the average is over months that have any transactions, so a partial year isn't diluted by empty months
func buildMonthComparison(transactions TransactionHistory, month, year string) MonthComparison {
	months := make([]string, 0, len(transactions[year]))
	for m := range transactions[year] {
		months = append(months, m)
	}

	cmp := MonthComparison{
		month:        buildExpenseCategoryBreakdown(transactions, year, []string{month}),
		monthsInYear: len(months),
	}
	yearBreakdown := buildExpenseCategoryBreakdown(transactions, year, months)
	if cmp.monthsInYear == 0 {
		return cmp
	}
	n := float64(cmp.monthsInYear)
	cmp.avgExpenseTotal = yearBreakdown.expenseTotal / n

	avg := make(map[string]float64, len(yearBreakdown.categories))
	for _, ct := range yearBreakdown.categories {
		avg[ct.category] = ct.total / n
	}

	inMonth := make(map[string]bool, len(cmp.month.categories))
	for _, ct := range cmp.month.categories {
		inMonth[ct.category] = true
		cmp.rows = append(cmp.rows, CategoryComparison{CategoryTotal: ct, monthlyAvg: avg[ct.category]})
	}
	// year categories are already sorted by total, so these come out ordered by average too
	for _, ct := range yearBreakdown.categories {
		if !inMonth[ct.category] {
			cmp.rows = append(cmp.rows, CategoryComparison{CategoryTotal: CategoryTotal{category: ct.category}, monthlyAvg: avg[ct.category]})
		}
	}

	return cmp
}

// loads transactions and compares the given month against its year's monthly averages
func calculateMonthComparison(month, year string) (MonthComparison, error) {
	transactions, err := LoadTransactions()
	if err != nil {
		return MonthComparison{}, fmt.Errorf("unable to load transactions file: %w", err)
	}
	return buildMonthComparison(transactions, month, year), nil
}
