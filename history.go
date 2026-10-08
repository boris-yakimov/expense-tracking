package main

import (
	"errors"
	"fmt"
)

// how many changes can be undone; older ones are dropped
const maxHistoryEntries = 100

var (
	ErrNothingToUndo = errors.New("nothing to undo")
	ErrNothingToRedo = errors.New("nothing to redo")
)

// one change that can be undone or redone
// every save rewrites the whole transaction history, so the state from before (or after) the change is kept as a full snapshot
type historyEntry struct {
	label    string // what the change did, e.g. "add expense €12.50 (lunch)"
	month    string // where the change happened, so undo/redo can show it
	year     string
	txType   string
	snapshot TransactionHistory // state to restore
}

// undo/redo stacks, kept in memory for the current session only
var (
	undoStack []historyEntry
	redoStack []historyEntry
)

// result of an undo or redo, used to tell the user what happened and where
type historyResult struct {
	label  string
	month  string
	year   string
	txType string
}

// saves transactions and records the state from before the save so the change can be undone
// a new change makes the previously undone changes impossible to redo
func saveTransactionsWithHistory(label, month, year, txType string, transactions TransactionHistory) error {
	before, err := LoadTransactions()
	if err != nil {
		return fmt.Errorf("unable to load transactions for undo history: %w", err)
	}

	if err := SaveTransactions(transactions); err != nil {
		return err
	}

	undoStack = pushHistory(undoStack, historyEntry{label: label, month: month, year: year, txType: txType, snapshot: before})
	redoStack = nil
	return nil
}

// reverts the most recent change
func undoLastChange() (historyResult, error) {
	return moveHistory(&undoStack, &redoStack, ErrNothingToUndo)
}

// re-applies the most recently undone change
func redoLastChange() (historyResult, error) {
	return moveHistory(&redoStack, &undoStack, ErrNothingToRedo)
}

// restores the top snapshot of `from` and pushes the current state onto `to` so the step can be reversed again
// stacks are only touched once the restore has been saved, so a failed save leaves undo/redo as it was
func moveHistory(from, to *[]historyEntry, emptyErr error) (historyResult, error) {
	if len(*from) == 0 {
		return historyResult{}, emptyErr
	}
	entry := (*from)[len(*from)-1]

	current, err := LoadTransactions()
	if err != nil {
		return historyResult{}, fmt.Errorf("unable to load transactions: %w", err)
	}
	if err := SaveTransactions(entry.snapshot); err != nil {
		return historyResult{}, fmt.Errorf("unable to restore transactions: %w", err)
	}

	*from = (*from)[:len(*from)-1]
	*to = pushHistory(*to, historyEntry{label: entry.label, month: entry.month, year: entry.year, txType: entry.txType, snapshot: current})
	return historyResult{label: entry.label, month: entry.month, year: entry.year, txType: entry.txType}, nil
}

// appends to a history stack, dropping the oldest entries past maxHistoryEntries
func pushHistory(stack []historyEntry, entry historyEntry) []historyEntry {
	stack = append(stack, entry)
	if len(stack) > maxHistoryEntries {
		stack = stack[len(stack)-maxHistoryEntries:]
	}
	return stack
}

// forgets all undo/redo history
func clearHistory() {
	undoStack = nil
	redoStack = nil
}

// short description of a transaction for undo/redo messages, e.g. "expense €12.50 (lunch)"
func describeTransaction(txType string, tx Transaction) string {
	if tx.Description == "" {
		return fmt.Sprintf("%s €%.2f", txType, tx.Amount)
	}
	return fmt.Sprintf("%s €%.2f (%s)", txType, tx.Amount, tx.Description)
}
