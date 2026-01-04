package main

import (
	"strconv"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func asModel(t *testing.T, tm tea.Model) model {
	t.Helper()
	m, ok := tm.(model)
	if !ok {
		t.Fatalf("expected model type, got %T", tm)
	}
	return m
}

func pressKeyType(t *testing.T, m model, kt tea.KeyType) model {
	t.Helper()
	tm, _ := m.Update(tea.KeyMsg{Type: kt})
	return asModel(t, tm)
}

func pressRune(t *testing.T, m model, r rune) model {
	t.Helper()
	tm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return asModel(t, tm)
}

func TestInitModelDefaults(t *testing.T) {
	m := initModel()

	if m.currentView != categoryListView {
		t.Fatalf("expected currentView=categoryListView, got %v", m.currentView)
	}

	if got, want := m.textInput.Placeholder, "Enter category name..."; got != want {
		t.Errorf("placeholder mismatch: got %q want %q", got, want)
	}
	if got, want := m.textInput.CharLimit, 100; got != want {
		t.Errorf("char limit mismatch: got %d want %d", got, want)
	}
	if got, want := m.textInput.Width, 50; got != want {
		t.Errorf("width mismatch: got %d want %d", got, want)
	}

	cats := m.todoManager.GetAllCategories()
	if len(cats) != 2 {
		t.Fatalf("expected 2 initial categories, got %d", len(cats))
	}

	titles := map[string]bool{}
	for _, c := range cats {
		titles[c.Title] = true
	}
	if !titles["Work"] || !titles["Personal"] {
		t.Errorf("expected initial categories 'Work' and 'Personal', got %+v", titles)
	}

	if got, want := len(m.categoryTable.Rows()), 2; got != want {
		t.Errorf("category table rows mismatch: got %d want %d", got, want)
	}
}

func TestEnterOpensItemListAndPopulatesRows(t *testing.T) {
	m := initModel()

	if len(m.categoryTable.Rows()) == 0 {
		t.Fatal("expected at least one category row")
	}
	selected := m.categoryTable.SelectedRow()
	id, err := strconv.Atoi(selected[0])
	if err != nil {
		t.Fatalf("failed to parse selected category id: %v", err)
	}
	cat, err := m.todoManager.GetCategoryByID(uint(id))
	if err != nil {
		t.Fatalf("failed to get category by id: %v", err)
	}
	expectedItems := len(cat.Items)

	m = pressKeyType(t, m, tea.KeyEnter)

	if m.currentView != itemListView {
		t.Fatalf("expected currentView=itemListView, got %v", m.currentView)
	}
	if got := len(m.itemTable.Rows()); got != expectedItems {
		t.Errorf("item table rows mismatch: got %d want %d", got, expectedItems)
	}
}

func TestAddCategoryFlow(t *testing.T) {
	m := initModel()
	initialCount := len(m.todoManager.GetAllCategories())

	m = pressRune(t, m, 'n')
	if m.currentView != addCategoryView {
		t.Fatalf("expected addCategoryView after 'n', got %v", m.currentView)
	}
	if got, want := m.textInput.Placeholder, "Enter category name..."; got != want {
		t.Errorf("placeholder mismatch: got %q want %q", got, want)
	}

	m.textInput.SetValue("Chores")
	m = pressKeyType(t, m, tea.KeyEnter)

	if m.currentView != categoryListView {
		t.Fatalf("expected to return to categoryListView after adding, got %v", m.currentView)
	}
	if got, want := len(m.todoManager.GetAllCategories()), initialCount+1; got != want {
		t.Fatalf("category count mismatch: got %d want %d", got, want)
	}

	found := false
	for _, c := range m.todoManager.GetAllCategories() {
		if c.Title == "Chores" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find newly added category 'Chores'")
	}
}

func TestAddItemFlowSuccess(t *testing.T) {
	m := initModel()
	// enter first category
	m = pressKeyType(t, m, tea.KeyEnter)
	initialItems := len(m.itemTable.Rows())

	// start add item flow
	m = pressRune(t, m, 'n')
	if m.currentView != addItemView {
		t.Fatalf("expected addItemView after 'n', got %v", m.currentView)
	}
	if m.inputMode != "name" {
		t.Fatalf("expected inputMode 'name', got %q", m.inputMode)
	}

	m.textInput.SetValue("Write tests")
	m = pressKeyType(t, m, tea.KeyEnter)
	if m.inputMode != "description" {
		t.Fatalf("expected inputMode 'description', got %q", m.inputMode)
	}

	m.textInput.SetValue("Cover model Update")
	m = pressKeyType(t, m, tea.KeyEnter)
	if m.inputMode != "duedate" {
		t.Fatalf("expected inputMode 'duedate', got %q", m.inputMode)
	}

	due := time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04")
	m.textInput.SetValue(due)
	m = pressKeyType(t, m, tea.KeyEnter)

	if m.currentView != itemListView {
		t.Fatalf("expected to return to itemListView after adding, got %v", m.currentView)
	}
	if got, want := len(m.itemTable.Rows()), initialItems+1; got != want {
		t.Fatalf("item count mismatch: got %d want %d", got, want)
	}
	if m.errorMsg != "" {
		t.Fatalf("unexpected errorMsg: %q", m.errorMsg)
	}
}

func TestAddItemInvalidDateShowsError(t *testing.T) {
	m := initModel()
	// enter first category
	m = pressKeyType(t, m, tea.KeyEnter)

	// start add item flow
	m = pressRune(t, m, 'n')

	m.textInput.SetValue("Task")
	m = pressKeyType(t, m, tea.KeyEnter)

	m.textInput.SetValue("Desc")
	m = pressKeyType(t, m, tea.KeyEnter)

	m.textInput.SetValue("bad-date-format")
	m = pressKeyType(t, m, tea.KeyEnter)

	if m.currentView != addItemView {
		t.Fatalf("expected to remain in addItemView on invalid date, got %v", m.currentView)
	}
	if m.errorMsg == "" {
		t.Fatalf("expected errorMsg to be set for invalid date")
	}
}

func TestDeleteCategoryDeletesSelected(t *testing.T) {
	m := initModel()
	initial := len(m.todoManager.GetAllCategories())
	if initial == 0 {
		t.Fatal("expected initial categories > 0")
	}

	m = pressRune(t, m, 'd')

	if got, want := len(m.todoManager.GetAllCategories()), initial-1; got != want {
		t.Fatalf("category count after delete mismatch: got %d want %d", got, want)
	}
	if got := len(m.categoryTable.Rows()); got != initial-1 {
		t.Fatalf("category table rows after delete mismatch: got %d want %d", got, initial-1)
	}
	if m.errorMsg != "" {
		t.Fatalf("unexpected errorMsg after delete: %q", m.errorMsg)
	}
}

func TestDeleteItemDeletesSelected(t *testing.T) {
	m := initModel()
	// enter first category
	m = pressKeyType(t, m, tea.KeyEnter)

	initial := len(m.itemTable.Rows())
	if initial == 0 {
		t.Skip("no initial items to delete in selected category")
	}

	m = pressRune(t, m, 'd')

	if got, want := len(m.itemTable.Rows()), initial-1; got != want {
		t.Fatalf("item rows after delete mismatch: got %d want %d", got, want)
	}
	if m.errorMsg != "" {
		t.Fatalf("unexpected errorMsg after item delete: %q", m.errorMsg)
	}
}
