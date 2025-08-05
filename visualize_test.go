package main

import (
	"fmt"
	"testing"
)

func TestListTransactions(t *testing.T) {
	fmt.Println("\nList Transactions without arguments:")
	// TODO: Fix
	if _, err := listAllTransactions([]string{}); err != nil {
		t.Errorf("list transactions without args failed: %v", err)
	}

	fmt.Println("\nList Transactions with \"expenses\" argument:")
	if _, err := listTransactionsByMonth("expenses", "July", "2025"); err != nil {
		t.Errorf("list expeses failed: %v", err)
	}

	fmt.Println("\nList Transactions with the \"investments\" argument:")
	if _, err := listTransactionsByMonth("investments", "July", "2025"); err != nil {
		t.Errorf("list investments failed: %v", err)
	}

	fmt.Println("\nList Transactions with the \"income\" argument:")
	if _, err := listTransactionsByMonth("income", "July", "2025"); err != nil {
		t.Errorf("list income failed: %v", err)
	}
}

func TestShowTotal(t *testing.T) {
	fmt.Println("\nShow Total Without Arguments:")
	if _, err := showTotal([]string{}); err != nil {
		t.Errorf("show total failed: %v", err)
	}
}

func TestShowAllowedCategories(t *testing.T) {
	fmt.Println("\nShow Allowed Expense Categories:")
	if err := showAllowedCategories("expenses"); err != nil {
		t.Errorf("show allowed expense categories failed: %v", err)
	}

	fmt.Println("\nShow Allowed Investment Categories:")
	if err := showAllowedCategories("investments"); err != nil {
		t.Errorf("show allowed investment categories failed: %v", err)
	}

	fmt.Println("\nShow Allowed Income Categories:")
	if err := showAllowedCategories("income"); err != nil {
		t.Errorf("show allowed income categories failed: %v", err)
	}
}
