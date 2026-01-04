package todo

import (
	"testing"
	"time"
)

func TestAddCategoryAndGetAll(t *testing.T) {
	tm := NewTodoManager()

	if got := len(tm.GetAllCategories()); got != 0 {
		t.Fatalf("expected no categories initially, got %d", got)
	}

	c1 := tm.AddCategory("Work")
	if c1 == nil || c1.ID != 1 || c1.Title != "Work" {
		t.Fatalf("unexpected category after first add: %+v", c1)
	}

	c2 := tm.AddCategory("Personal")
	if c2 == nil || c2.ID != 2 || c2.Title != "Personal" {
		t.Fatalf("unexpected category after second add: %+v", c2)
	}

	cats := tm.GetAllCategories()
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
	if cats[0].ID != 1 || cats[1].ID != 2 {
		t.Errorf("expected IDs [1,2], got [%d,%d]", cats[0].ID, cats[1].ID)
	}
}

func TestGetCategoryByID_NotFound(t *testing.T) {
	tm := NewTodoManager()
	_, err := tm.GetCategoryByID(42)
	if err == nil {
		t.Fatal("expected error when category does not exist")
	}
}

func TestUpdateCategory_TitleChanges(t *testing.T) {
	tm := NewTodoManager()
	c := tm.AddCategory("Old")
	if err := tm.UpdateCategory(c.ID, "New"); err != nil {
		t.Fatalf("unexpected error updating category: %v", err)
	}
	got, err := tm.GetCategoryByID(c.ID)
	if err != nil {
		t.Fatalf("unexpected error getting category: %v", err)
	}
	if got.Title != "New" {
		t.Fatalf("expected title 'New', got %q", got.Title)
	}
}

func TestDeleteCategory_RemovesAndReturnsErrorOnMissing(t *testing.T) {
	tm := NewTodoManager()
	c := tm.AddCategory("Temp")

	if err := tm.DeleteCategory(c.ID); err != nil {
		t.Fatalf("unexpected error deleting existing category: %v", err)
	}
	if _, err := tm.GetCategoryByID(c.ID); err == nil {
		t.Fatal("expected error when getting deleted category")
	}
	if err := tm.DeleteCategory(c.ID); err == nil {
		t.Fatal("expected error when deleting already deleted category")
	}
}

func TestAddItemAndGetItemByID(t *testing.T) {
	tm := NewTodoManager()
	c := tm.AddCategory("Cat")
	due := time.Date(2025, 1, 2, 15, 4, 0, 0, time.UTC)

	it, err := tm.AddItem(c.ID, "Task", "Desc", due)
	if err != nil {
		t.Fatalf("unexpected error adding item: %v", err)
	}
	if it.ID != 1 || it.Name != "Task" || it.Description != "Desc" || !it.DueDate.Equal(due) {
		t.Fatalf("unexpected item after add: %+v", it)
	}

	got, err := tm.GetItemByID(c.ID, it.ID)
	if err != nil {
		t.Fatalf("unexpected error getting item: %v", err)
	}
	if got.ID != it.ID || got.Name != "Task" || got.Description != "Desc" || !got.DueDate.Equal(due) {
		t.Fatalf("mismatch on retrieved item: %+v", got)
	}
}

func TestGetItemByID_NotFound(t *testing.T) {
	tm := NewTodoManager()
	c := tm.AddCategory("Cat")
	_, err := tm.GetItemByID(c.ID, 999)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestGetAllItems_ErrorOnMissingCategory(t *testing.T) {
	tm := NewTodoManager()
	if _, err := tm.GetAllItems(123); err == nil {
		t.Fatal("expected error when getting items for missing category")
	}
}

func TestUpdateItem_PartialFields(t *testing.T) {
	tm := NewTodoManager()
	c := tm.AddCategory("Cat")
	due := time.Date(2025, 1, 2, 15, 4, 0, 0, time.UTC)
	it, err := tm.AddItem(c.ID, "Task", "Desc", due)
	if err != nil {
		t.Fatalf("unexpected error adding item: %v", err)
	}

	// Update only name
	newName := "NewTask"
	if err := tm.UpdateItem(c.ID, it.ID, &newName, nil, nil); err != nil {
		t.Fatalf("unexpected error updating item name: %v", err)
	}
	got, err := tm.GetItemByID(c.ID, it.ID)
	if err != nil {
		t.Fatalf("unexpected error getting item: %v", err)
	}
	if got.Name != "NewTask" || got.Description != "Desc" || !got.DueDate.Equal(due) {
		t.Fatalf("unexpected item after name update: %+v", got)
	}

	// Update description and due date
	newDesc := "NewDesc"
	newDue := due.Add(24 * time.Hour)
	if err := tm.UpdateItem(c.ID, it.ID, nil, &newDesc, &newDue); err != nil {
		t.Fatalf("unexpected error updating item desc/due: %v", err)
	}
	got, err = tm.GetItemByID(c.ID, it.ID)
	if err != nil {
		t.Fatalf("unexpected error getting item: %v", err)
	}
	if got.Name != "NewTask" || got.Description != "NewDesc" || !got.DueDate.Equal(newDue) {
		t.Fatalf("unexpected item after desc/due update: %+v", got)
	}
}

func TestDeleteItem_RemovesFromCategory(t *testing.T) {
	tm := NewTodoManager()
	c := tm.AddCategory("Cat")
	due := time.Now()
	it1, _ := tm.AddItem(c.ID, "A", "", due)
	_, _ = tm.AddItem(c.ID, "B", "", due)

	items, _ := tm.GetAllItems(c.ID)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if err := tm.DeleteItem(c.ID, it1.ID); err != nil {
		t.Fatalf("unexpected error deleting item: %v", err)
	}
	items, _ = tm.GetAllItems(c.ID)
	if len(items) != 1 || items[0].Name != "B" {
		t.Fatalf("expected remaining item 'B', got %+v", items)
	}

	// Deleting again should error
	if err := tm.DeleteItem(c.ID, it1.ID); err == nil {
		t.Fatal("expected error when deleting missing item")
	}
}
