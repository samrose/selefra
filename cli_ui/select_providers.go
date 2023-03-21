package cli_ui

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
)

const DefaultSelectProvidersTitle = "[ Use arrows to move, Space to select, Enter to complete the selection ]"

// SelectProviders Give a list of providers and let the user select some of them
// Does the installation sequence have to be consistent with the selected sequence? Temporarily, I think it can be inconsistent
func SelectProviders(providers []string, title ...string) map[string]struct{} {

	if len(title) == 0 {
		title = append(title, DefaultSelectProvidersTitle)
	}

	selectProviders := make(map[string]struct{})

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	l := newList(title[0], listForShow(providers, selectProviders))
	ui.Render(l)

	previousKey := ""
	uiEvents := ui.PollEvents()

	for {
		e := <-uiEvents
		switch e.ID {
		case "j", "<Down>":
			if len(l.Rows) == 0 {
				continue
			}
			l.ScrollDown()
		case "k", "<Up>":
			if len(l.Rows) == 0 {
				continue
			}
			l.ScrollUp()
		case "<C-d>":
			if len(l.Rows) == 0 {
				continue
			}
			l.ScrollHalfPageDown()
		case "<C-c>":
			return nil
		case "<C-u>":
			if len(l.Rows) == 0 {
				continue
			}
			l.ScrollHalfPageUp()
		case "<C-f>":
			if len(l.Rows) == 0 {
				continue
			}
			l.ScrollPageDown()
		case "<C-b>":
			if len(l.Rows) == 0 {
				continue
			}
			l.ScrollPageUp()
		case "g":
			if len(l.Rows) == 0 {
				continue
			}
			if previousKey == "g" {
				l.ScrollTop()
			}
		case "<Enter>":
			return selectProviders
		case "<Space>":

			if len(l.Rows) == 0 {
				continue
			}

			// Select or deselect provider
			operateProviderName := providers[l.SelectedRow]
			if _, exists := selectProviders[operateProviderName]; exists {
				delete(selectProviders, operateProviderName)
			} else {
				selectProviders[operateProviderName] = struct{}{}
			}
			l.Rows = listForShow(providers, selectProviders)

		case "<Home>":
			if len(l.Rows) == 0 {
				continue
			}
			l.ScrollTop()
		case "G", "<End>":
			if len(l.Rows) == 0 {
				continue
			}
			l.ScrollBottom()
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(l)
	}
}

// Create a list widgets for select providers
func newList(title string, lines []string) *widgets.List {
	l := widgets.NewList()
	l.Rows = lines
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.Title = title
	l.BorderLeft = false
	l.BorderRight = false
	l.BorderTop = false
	l.BorderBottom = false
	l.SelectedRowStyle = ui.NewStyle(ui.ColorRed)
	l.SetRect(0, 0, 800, 30)
	return l
}

// Shows all the providers and which ones are currently selected
func listForShow(providers []string, selectedProviders map[string]struct{}) []string {
	var listProviders []string
	for _, provider := range providers {
		if _, exists := selectedProviders[provider]; exists {
			// Putting the checkbox first avoids the provider name alignment problem
			listProviders = append(listProviders, " [✔] "+provider)
		} else {
			listProviders = append(listProviders, " [ ] "+provider)
		}
	}
	return listProviders
}
