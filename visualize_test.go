package main

import (
	"fmt"
	"testing"
)

func TestListTransactions(t *testing.T) {
	// make sure to use the actual data file for those tests
	transactionsFilePath = "data.json"

	fmt.Println("\nList Transactions without arguments:")
	if _, err := listAllTransactions(); err != nil {
		t.Errorf("list transactions without args failed: %s", err)
	}

	fmt.Println("\nList Transactions with \"expenses\" argument:")
	if _, err := listTransactionsByMonth("expenses", "july", "2025"); err != nil {
		t.Errorf("list expeses failed: %s", err)
	}

	fmt.Println("\nList Transactions with the \"investments\" argument:")
	if _, err := listTransactionsByMonth("investments", "july", "2025"); err != nil {
		t.Errorf("list investments failed: %s", err)
	}

	fmt.Println("\nList Transactions with the \"income\" argument:")
	if _, err := listTransactionsByMonth("income", "july", "2025"); err != nil {
		t.Errorf("list income failed: %s", err)
	}

	fmt.Println("\nList Transactions with invalid type")
	if _, err := listTransactionsByMonth("invalidtype", "july", "2025"); err == nil {
		t.Error("expected error for invalid transaction type, got nil")
	}
}

func TestShowTotal(t *testing.T) {
	fmt.Println("\nShow Total Without Arguments:")
	if _, err := visualizeTransactions([]string{}); err != nil {
		t.Errorf("show total failed: %v", err)
	}
}

func TestShowAllowedCategories(t *testing.T) {
	fmt.Println("\nShow Allowed Expense Categories:")
	if err := showAllowedCategories("expenses"); err != nil {
		t.Errorf("show allowed expense categories failed: %s", err)
	}

	fmt.Println("\nShow Allowed Investment Categories:")
	if err := showAllowedCategories("investments"); err != nil {
		t.Errorf("show allowed investment categories failed: %s", err)
	}

	fmt.Println("\nShow Allowed Income Categories:")
	if err := showAllowedCategories("income"); err != nil {
		t.Errorf("show allowed income categories failed: %s", err)
	}
}
