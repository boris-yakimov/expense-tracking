package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// all transactions of a type in a month, straight from storage
func storedTransactions(t *testing.T, year, month, txType string) []Transaction {
	t.Helper()
	transactions, err := LoadTransactions()
	if err != nil {
		t.Fatalf("load transactions: %v", err)
	}
	return transactions[year][month][txType]
}

func historyTestData() TransactionHistory {
	return TransactionHistory{
		"2026": {
			"march": {
				"expense": {
					{Id: "aaaa1111", Amount: 10, Category: "food", Description: "lunch"},
					{Id: "bbbb2222", Amount: 50, Category: "car", Description: "fuel"},
				},
			},
		},
	}
}

func setupHistoryTest(t *testing.T) {
	t.Helper()
	setupTestStorage(t, StorageSQLite)
	if err := saveTransactionsToTestStorage(historyTestData()); err != nil {
		t.Fatalf("failed to set up test data: %v", err)
	}
}

func TestUndoRedoAdd(t *testing.T) {
	setupHistoryTest(t)

	err := handleAddTransaction(AddTransactionRequest{Type: "expense", Amount: "12.5", Category: "food", Description: "coffee", Month: "march", Year: "2026"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if got := len(storedTransactions(t, "2026", "march", "expense")); got != 3 {
		t.Fatalf("expected 3 expenses after add, got %d", got)
	}

	res, err := undoLastChange()
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	if res.label != "add expense €12.50 (coffee)" || res.month != "march" || res.year != "2026" || res.txType != "expense" {
		t.Errorf("unexpected undo result: %+v", res)
	}
	if got := len(storedTransactions(t, "2026", "march", "expense")); got != 2 {
		t.Fatalf("expected 2 expenses after undo, got %d", got)
	}

	if _, err := redoLastChange(); err != nil {
		t.Fatalf("redo: %v", err)
	}
	expenses := storedTransactions(t, "2026", "march", "expense")
	if len(expenses) != 3 {
		t.Fatalf("expected 3 expenses after redo, got %d", len(expenses))
	}
	var found bool
	for _, tx := range expenses {
		found = found || tx.Description == "coffee"
	}
	if !found {
		t.Errorf("redo did not bring back the added transaction: %+v", expenses)
	}
}

func TestUndoRedoDelete(t *testing.T) {
	setupHistoryTest(t)

	if err := handleDeleteTransaction("expense", "aaaa1111"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if got := len(storedTransactions(t, "2026", "march", "expense")); got != 1 {
		t.Fatalf("expected 1 expense after delete, got %d", got)
	}

	res, err := undoLastChange()
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	if res.label != "delete expense €10.00 (lunch)" {
		t.Errorf("unexpected undo label %q", res.label)
	}
	tx, err := getTransactionById("aaaa1111")
	if err != nil || tx.Description != "lunch" || tx.Amount != 10 {
		t.Errorf("undo did not restore the deleted transaction: %+v, err %v", tx, err)
	}

	if _, err := redoLastChange(); err != nil {
		t.Fatalf("redo: %v", err)
	}
	if got := len(storedTransactions(t, "2026", "march", "expense")); got != 1 {
		t.Errorf("expected 1 expense after redo, got %d", got)
	}
}

func TestUndoRedoUpdate(t *testing.T) {
	setupHistoryTest(t)

	err := handleUpdateTransaction(UpdateTransactionRequest{Type: "expense", Id: "bbbb2222", Amount: "75", Category: "car", Description: "tires"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	res, err := undoLastChange()
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	if res.label != "update expense €75.00 (tires)" || res.month != "march" || res.year != "2026" {
		t.Errorf("unexpected undo result: %+v", res)
	}
	if tx, _ := getTransactionById("bbbb2222"); tx.Amount != 50 || tx.Description != "fuel" {
		t.Errorf("undo did not restore the original values: %+v", tx)
	}

	if _, err := redoLastChange(); err != nil {
		t.Fatalf("redo: %v", err)
	}
	if tx, _ := getTransactionById("bbbb2222"); tx.Amount != 75 || tx.Description != "tires" {
		t.Errorf("redo did not re-apply the update: %+v", tx)
	}
}

// several changes are undone newest first, and redone in the original order
func TestUndoMultipleSteps(t *testing.T) {
	setupHistoryTest(t)

	if err := handleDeleteTransaction("expense", "aaaa1111"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := handleDeleteTransaction("expense", "bbbb2222"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	for want := 1; want <= 2; want++ {
		if _, err := undoLastChange(); err != nil {
			t.Fatalf("undo %d: %v", want, err)
		}
		if got := len(storedTransactions(t, "2026", "march", "expense")); got != want {
			t.Fatalf("after undo %d expected %d expenses, got %d", want, want, got)
		}
	}
	if _, err := undoLastChange(); !errors.Is(err, ErrNothingToUndo) {
		t.Errorf("expected ErrNothingToUndo once history is exhausted, got %v", err)
	}

	res, _ := redoLastChange()
	if res.label != "delete expense €10.00 (lunch)" {
		t.Errorf("first redo should re-apply the first delete, got %q", res.label)
	}
}

// making a new change after an undo drops the undone changes, like in any editor
func TestNewChangeClearsRedo(t *testing.T) {
	setupHistoryTest(t)

	if err := handleDeleteTransaction("expense", "aaaa1111"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := undoLastChange(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if err := handleDeleteTransaction("expense", "bbbb2222"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := redoLastChange(); !errors.Is(err, ErrNothingToRedo) {
		t.Errorf("expected ErrNothingToRedo after a new change, got %v", err)
	}
}

// failed actions don't save anything, so they must not leave an undo step behind
func TestFailedActionIsNotRecorded(t *testing.T) {
	setupHistoryTest(t)

	if err := handleAddTransaction(AddTransactionRequest{Type: "expense", Amount: "not a number", Category: "food", Month: "march", Year: "2026"}); err == nil {
		t.Fatalf("expected add with an invalid amount to fail")
	}
	if _, err := undoLastChange(); !errors.Is(err, ErrNothingToUndo) {
		t.Errorf("expected nothing to undo, got %v", err)
	}
}

func TestHistoryIsCapped(t *testing.T) {
	var stack []historyEntry
	for i := 0; i < maxHistoryEntries+5; i++ {
		stack = pushHistory(stack, historyEntry{label: string(rune('a' + i%26))})
	}
	if len(stack) != maxHistoryEntries {
		t.Errorf("expected history to be capped at %d, got %d", maxHistoryEntries, len(stack))
	}
}

// ctrl+z / ctrl+y on the transactions page undo and redo, and the header says what happened
func TestUndoRedoKeys(t *testing.T) {
	setupHistoryTest(t)
	screen := startSimulatedTui(t)

	if err := handleDeleteTransaction("expense", "aaaa1111"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	tui.QueueUpdateDraw(func() {
		if _, err := gridVisualizeTransactions("march", "2026", "expense", true); err != nil {
			t.Errorf("gridVisualizeTransactions: %v", err)
		}
	})
	waitFor(t, "grid drawn", func() bool { return strings.Contains(screenText(screen), "March 2026") })

	pressKey(screen, tcell.KeyCtrlZ, 0)
	waitFor(t, "undo applied", func() bool {
		tx, err := getTransactionById("aaaa1111")
		return err == nil && tx.Description == "lunch"
	})
	waitFor(t, "undo note shown", func() bool { return strings.Contains(screenText(screen), "Undo: delete expense €10.00 (lunch)") })

	pressKey(screen, tcell.KeyCtrlY, 0)
	waitFor(t, "redo applied", func() bool {
		_, err := getTransactionById("aaaa1111")
		return err != nil
	})
	waitFor(t, "redo note shown", func() bool { return strings.Contains(screenText(screen), "Redo: delete expense €10.00 (lunch)") })

	pressKey(screen, tcell.KeyCtrlY, 0)
	waitFor(t, "nothing to redo note", func() bool { return strings.Contains(screenText(screen), "Nothing to redo") })
}

// everything currently on the simulated screen as plain text
func screenText(screen tcell.SimulationScreen) string {
	cells, w, h := screen.GetContents()
	var sb strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if r := cells[y*w+x].Runes; len(r) > 0 {
				sb.WriteString(string(r))
			} else {
				sb.WriteRune(' ')
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}
