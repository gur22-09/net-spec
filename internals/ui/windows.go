package ui

import (
	"fmt"
	"path/filepath"

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
	window.Resize(fyne.NewSize(1440, 900))
	numOfClomuns := 9

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
					label.SetText("Process")
				case 3:
					label.SetText("Local Address")
				case 4:
					label.SetText("Local Port")
				case 5:
					label.SetText("Remote Address")
				case 6:
					label.SetText("Remote Port")
				case 7:
					label.SetText("State")
				case 8:
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
				label.SetText(filepath.Base(conn.Process))
			case 3:
				label.SetText(conn.LocalAddress)
			case 4:
				label.SetText(fmt.Sprintf("%d", conn.LocalPort))
			case 5:
				label.SetText(conn.RemoteAddress)
			case 6:
				label.SetText(fmt.Sprintf("%d", conn.RemotePort))
			case 7:
				label.SetText(conn.State)
			case 8:
				label.SetText("TODO") // TODO - add duration after snapshot logic
			}
		},
	)

	// Set column widths
	table.SetColumnWidth(0, 60)
	table.SetColumnWidth(1, 60)
	table.SetColumnWidth(2, 300)
	table.SetColumnWidth(3, 400)
	table.SetColumnWidth(4, 100)
	table.SetColumnWidth(5, 400)
	table.SetColumnWidth(6, 100)
	table.SetColumnWidth(7, 120)
	table.SetColumnWidth(8, 60)

	window.SetContent(table)
	window.CenterOnScreen()

	return &MainWindow{
		Window: window,
	}
}
