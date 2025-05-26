package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/gur22-09/net-spec/internals/constants"
	"github.com/gur22-09/net-spec/internals/network"
)

type MainWindow struct {
	Window fyne.Window
}

func NewMainWindow(app fyne.App, connections []network.Connection) *MainWindow {
	window := app.NewWindow(constants.AppName)
	window.Resize(fyne.NewSize(800, 600))
	numOfClomuns := 6
	// Create the table widget
	table := widget.NewTable(
		func() (int, int) {
			return len(connections) + 1, numOfClomuns // +1 for extra header row
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				switch id.Col {
				case 0:
					label.SetText("Proto")
				case 1:
					label.SetText("PID")
				case 2:
					label.SetText("Local Address")
				case 3:
					label.SetText("Remote Address")
				case 4:
					label.SetText("State")
				case 5:
					label.SetText("Duration")
				}
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			// Data rows
			conn := connections[id.Row-1]
			label.TextStyle = fyne.TextStyle{}

			switch id.Col {
			case 0:
				label.SetText(conn.Protocol)
			case 1:
				label.SetText(fmt.Sprintf("%d", conn.PID))
			case 2:
				label.SetText(conn.LocalAddress)
			case 3:
				label.SetText(conn.RemoteAddress)
			case 4:
				label.SetText(conn.State)
			case 5:
				label.SetText("TODO") // TODO - add duration after snapshot logic
			}
		},
	)

	// Set column widths
	table.SetColumnWidth(0, 60)
	table.SetColumnWidth(1, 60)
	table.SetColumnWidth(2, 200)
	table.SetColumnWidth(3, 200)
	table.SetColumnWidth(4, 100)
	table.SetColumnWidth(5, 100)

	window.SetContent(table)
	window.CenterOnScreen()

	return &MainWindow{
		Window: window,
	}
}
