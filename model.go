package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/datrine/tui-todo-app/todo"
)

type model struct {
	todoManager       *todo.TodoManager
	currentView       viewMode
	currentCategoryID uint
	currentItemID     uint
	textInput         textinput.Model
	infoText          string
	categoryTable     table.Model
	itemTable         table.Model
	cursorPos         int
	errorMsg          string
	inputMode         string // "name", "description", "duedate"
	formData          map[string]string
}

type viewMode int

const (
	categoryListView viewMode = iota
	itemListView
	addCategoryView
	addItemView
	editItemView
)

func initModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter category name..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	m := model{
		todoManager: todo.NewTodoManager(),
		currentView: categoryListView,
		textInput:   ti,
		formData:    make(map[string]string),
	}
	/*
		// Add some sample data
		cat1 := m.todoManager.AddCategory("Work")
		cat2 := m.todoManager.AddCategory("Personal")
		m.todoManager.AddItem(cat1.ID, "Review PR", "Review pull request #123", time.Now().Add(24*time.Hour))
		m.todoManager.AddItem(cat1.ID, "Meeting", "Team standup at 10am", time.Now().Add(2*time.Hour))
		m.todoManager.AddItem(cat2.ID, "Gym", "Leg day workout", time.Now().Add(5*time.Hour))
	*/
	m.updateCategoryTable()
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.errorMsg = ""

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			// Go back to previous view
			switch m.currentView {
			case itemListView:
				m.currentView = categoryListView
				m.updateCategoryTable()
			case addCategoryView, addItemView, editItemView:
				if m.currentCategoryID != 0 {
					m.currentView = itemListView
				} else {
					m.currentView = categoryListView
				}
				m.textInput.SetValue("")
				m.formData = make(map[string]string)
				m.inputMode = ""
			}
			return m, nil

		case "enter":
			switch m.currentView {
			case categoryListView:
				// Enter category to view items
				if len(m.categoryTable.Rows()) > 0 {
					selectedRow := m.categoryTable.SelectedRow()
					if len(selectedRow) > 0 {
						fmt.Sscanf(selectedRow[0], "%d", &m.currentCategoryID)
						m.currentView = itemListView
						m.updateItemTable()
					}
				}

			case addCategoryView:
				// Add category
				title := m.textInput.Value()
				if title != "" {
					m.todoManager.AddCategory(title)
					m.currentView = categoryListView
					m.textInput.SetValue("")
					m.updateCategoryTable()
				}

			case addItemView:
				// Handle multi-step form
				value := m.textInput.Value()
				if value == "" {
					m.errorMsg = "Field cannot be empty"
					return m, nil
				}

				switch m.inputMode {
				case "name":
					m.formData["name"] = value
					m.inputMode = "description"
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Enter description..."
				case "description":
					m.formData["description"] = value
					m.inputMode = "duedate"
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Enter due date (YYYY-MM-DD HH:MM)..."
				case "duedate":
					dueDate, err := time.Parse("2006-01-02 15:04", value)
					if err != nil {
						m.errorMsg = "Invalid date format. Use YYYY-MM-DD HH:MM"
						return m, nil
					}

					_, err = m.todoManager.AddItem(
						m.currentCategoryID,
						m.formData["name"],
						m.formData["description"],
						dueDate,
					)
					if err != nil {
						m.errorMsg = err.Error()
					} else {
						m.currentView = itemListView
						m.updateItemTable()
						m.textInput.SetValue("")
						m.formData = make(map[string]string)
						m.inputMode = ""
					}
				}
			}

		case "n":
			// Add new
			switch m.currentView {
			case categoryListView:
				m.currentView = addCategoryView
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Enter category name..."
				m.textInput.Focus()
			case itemListView:
				m.currentView = addItemView
				m.inputMode = "name"
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Enter item name..."
				m.textInput.Focus()
			}

		case "d":
			// Delete
			switch m.currentView {
			case categoryListView:
				if len(m.categoryTable.Rows()) > 0 {
					selectedRow := m.categoryTable.SelectedRow()
					if len(selectedRow) > 0 {
						var catID uint
						fmt.Sscanf(selectedRow[0], "%d", &catID)
						err := m.todoManager.DeleteCategory(catID)
						if err != nil {
							m.errorMsg = err.Error()
						}
						m.updateCategoryTable()
					}
				}
			case itemListView:
				if len(m.itemTable.Rows()) > 0 {
					selectedRow := m.itemTable.SelectedRow()
					if len(selectedRow) > 0 {
						var itemID uint
						fmt.Sscanf(selectedRow[0], "%d", &itemID)
						err := m.todoManager.DeleteItem(m.currentCategoryID, itemID)
						if err != nil {
							m.errorMsg = err.Error()
						}
						m.updateItemTable()
					}
				}
			}
		}
	}

	// Update components
	switch m.currentView {
	case categoryListView:
		m.categoryTable, cmd = m.categoryTable.Update(msg)
	case itemListView:
		m.itemTable, cmd = m.itemTable.Update(msg)
	case addCategoryView, addItemView, editItemView:
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(1, 0)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	var content string
	var help string

	switch m.currentView {
	case categoryListView:
		if len(m.todoManager.GetAllCategories()) == 0 {
			promptStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))
			styledText := promptStyle.Render("Please add a category")
			content = styledText
		} else {
			content = titleStyle.Render("📋 Categories") + "\n\n" +
				m.categoryTable.View()
		}
		help = "n: new category | enter: view items | d: delete | q: quit"

	case itemListView:
		cat, _ := m.todoManager.GetCategoryByID(m.currentCategoryID)
		title := "Items"
		if cat != nil {
			title = fmt.Sprintf("Items in '%s'", cat.Title)
		}
		content = titleStyle.Render("📝 "+title) + "\n\n" +
			m.itemTable.View()
		help = "n: new item | d: delete | esc: back | q: quit"

	case addCategoryView:
		content = titleStyle.Render("➕ Add New Category") + "\n\n" +
			m.textInput.View()
		help = "enter: save | esc: cancel"

	case addItemView:
		var prompt string
		switch m.inputMode {
		case "name":
			prompt = "Step 1/3: Item Name"
		case "description":
			prompt = "Step 2/3: Description"
		case "duedate":
			prompt = "Step 3/3: Due Date (YYYY-MM-DD HH:MM)"
		default:
			m.inputMode = "name"
			prompt = "Step 1/3: Item Name"
		}
		content = titleStyle.Render("➕ Add New Item") + "\n\n" +
			prompt + "\n\n" +
			m.textInput.View()
		help = "enter: next/save | esc: cancel"
	}

	errorMsg := ""
	if m.errorMsg != "" {
		errorMsg = "\n" + errorStyle.Render("⚠️  "+m.errorMsg)
	}

	return fmt.Sprintf(
		"%s%s\n\n%s\n",
		content,
		errorMsg,
		helpStyle.Render(help),
	)
}

func (m *model) updateCategoryTable() {
	categories := m.todoManager.GetAllCategories()

	if len(categories) == 0 {
		m.infoText = "Add a category"
		return
	}

	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Title", Width: 30},
		{Title: "Items", Width: 10},
	}

	var rows []table.Row
	for _, cat := range categories {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", cat.ID),
			cat.Title,
			fmt.Sprintf("%d", len(cat.Items)),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m.categoryTable = t
}

func (m *model) updateItemTable() {
	items, err := m.todoManager.GetAllItems(m.currentCategoryID)
	if err != nil {
		m.errorMsg = err.Error()
		return
	}

	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 20},
		{Title: "Description", Width: 30},
		{Title: "Due Date", Width: 20},
	}

	var rows []table.Row
	for _, item := range items {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", item.ID),
			item.Name,
			item.Description,
			item.DueDate.Format("2006-01-02 15:04"),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m.itemTable = t
}
