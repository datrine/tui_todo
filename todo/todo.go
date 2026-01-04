package todo

import (
	"fmt"
	"time"
)

// Category represents a todo category
type Category struct {
	ID    uint
	Title string
	Items []Item
}

// Item represents a todo item
type Item struct {
	ID          uint
	Name        string
	Description string
	DueDate     time.Time
}

// TodoManager manages categories and items
type TodoManager struct {
	categories []Category
	nextCatID  uint
	nextItemID uint
}

// NewTodoManager creates a new TodoManager
func NewTodoManager() *TodoManager {
	return &TodoManager{
		categories: make([]Category, 0),
		nextCatID:  1,
		nextItemID: 1,
	}
}

// ==================== CATEGORY CRUD ====================

// AddCategory creates a new category
func (tm *TodoManager) AddCategory(title string) *Category {
	cat := Category{
		ID:    tm.nextCatID,
		Title: title,
		Items: make([]Item, 0),
	}
	tm.categories = append(tm.categories, cat)
	tm.nextCatID++
	return &tm.categories[len(tm.categories)-1]
}

// GetAllCategories returns all categories
func (tm *TodoManager) GetAllCategories() []Category {
	return tm.categories
}

// GetCategoryByID returns a category by ID
func (tm *TodoManager) GetCategoryByID(catID uint) (*Category, error) {
	for i := range tm.categories {
		if tm.categories[i].ID == catID {
			return &tm.categories[i], nil
		}
	}
	return nil, fmt.Errorf("category with id %d not found", catID)
}

// UpdateCategory updates a category's title
func (tm *TodoManager) UpdateCategory(catID uint, title string) error {
	cat, err := tm.GetCategoryByID(catID)
	if err != nil {
		return err
	}
	cat.Title = title
	return nil
}

// DeleteCategory removes a category
func (tm *TodoManager) DeleteCategory(catID uint) error {
	for i := range tm.categories {
		if tm.categories[i].ID == catID {
			tm.categories = append(tm.categories[:i], tm.categories[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("category with id %d not found", catID)
}

// AddItem adds an item to a category
func (tm *TodoManager) AddItem(catID uint, name, description string, dueDate time.Time) (*Item, error) {
	cat, err := tm.GetCategoryByID(catID)
	if err != nil {
		return nil, err
	}

	item := Item{
		ID:          tm.nextItemID,
		Name:        name,
		Description: description,
		DueDate:     dueDate,
	}
	cat.Items = append(cat.Items, item)
	tm.nextItemID++

	return &cat.Items[len(cat.Items)-1], nil
}

// GetItemByID returns an item from a category
func (tm *TodoManager) GetItemByID(catID, itemID uint) (*Item, error) {
	cat, err := tm.GetCategoryByID(catID)
	if err != nil {
		return nil, err
	}

	for i := range cat.Items {
		if cat.Items[i].ID == itemID {
			return &cat.Items[i], nil
		}
	}
	return nil, fmt.Errorf("item with id %d not found in category %d", itemID, catID)
}

// GetAllItems returns all items in a category
func (tm *TodoManager) GetAllItems(catID uint) ([]Item, error) {
	cat, err := tm.GetCategoryByID(catID)
	if err != nil {
		return nil, err
	}
	return cat.Items, nil
}

// UpdateItem updates an item's fields
func (tm *TodoManager) UpdateItem(catID, itemID uint, name, description *string, dueDate *time.Time) error {
	item, err := tm.GetItemByID(catID, itemID)
	if err != nil {
		return err
	}

	if name != nil {
		item.Name = *name
	}
	if description != nil {
		item.Description = *description
	}
	if dueDate != nil {
		item.DueDate = *dueDate
	}
	return nil
}

// DeleteItem removes an item from a category
func (tm *TodoManager) DeleteItem(catID, itemID uint) error {
	cat, err := tm.GetCategoryByID(catID)
	if err != nil {
		return err
	}

	for i := range cat.Items {
		if cat.Items[i].ID == itemID {
			cat.Items = append(cat.Items[:i], cat.Items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("item with id %d not found in category %d", itemID, catID)
}
