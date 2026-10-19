package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
)

// LogTray wraps LogPanel into a sleek collapsible bottom bar with smart auto-expand.
type LogTray struct {
	panel       *LogPanel
	window      fyne.Window
	isExpanded  bool
	barLayout   *fyne.Container
	fullLayout  *fyne.Container
	container   *fyne.Container
	lblLatest   *widget.Label
	lblCount    *widget.Label
	btnToggle   *widget.Button
	btnCopy     *widget.Button
	btnClear    *widget.Button
	OnHeightChg func(isExpanded bool)
}

func NewLogTray(panel *LogPanel, w fyne.Window) *LogTray {
	lt := &LogTray{
		panel:      panel,
		window:     w,
		isExpanded: false,
	}

	lt.lblLatest = widget.NewLabelWithStyle("📝 [Hệ thống] Sẵn sàng hoạt động", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	lt.lblLatest.Truncation = fyne.TextTruncateEllipsis

	lt.lblCount = widget.NewLabelWithStyle("[0]", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	lt.btnToggle = widget.NewButton("▲  Mở Rộng", func() {
		lt.Toggle()
	})

	lt.btnCopy = widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		lt.panel.mu.Lock()
		fullText := strings.Join(lt.panel.logLines, "\n")
		lt.panel.mu.Unlock()
		if lt.window != nil {
			lt.window.Clipboard().SetContent(fullText)
			lt.btnCopy.SetIcon(theme.ConfirmIcon())
			time.AfterFunc(1500*time.Millisecond, func() {
				fyne.Do(func() {
					lt.btnCopy.SetIcon(theme.ContentCopyIcon())
				})
			})
		}
	})

	lt.btnClear = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		lt.panel.Clear()
		lt.lblLatest.SetText("📝 [Hệ thống] Đã dọn sạch nhật ký.")
		lt.lblCount.SetText("[0]")
	})

	rightControls := container.NewHBox(lt.lblCount, lt.btnCopy, lt.btnClear, lt.btnToggle)
	topBar := container.NewBorder(nil, nil, nil, rightControls, lt.lblLatest)

	lt.barLayout = container.NewVBox(
		widget.NewSeparator(),
		topBar,
	)

	lt.fullLayout = container.NewBorder(
		lt.barLayout,
		nil,
		nil,
		nil,
		lt.panel.scroll,
	)

	lt.container = container.NewStack(lt.barLayout)

	lt.panel.OnLog = func(msg string) {
		lt.onNewLog(msg)
	}

	return lt
}

func (lt *LogTray) onNewLog(msg string) {
	ts := time.Now().Format("15:04:05")
	cleanMsg := strings.TrimSpace(msg)
	preview := fmt.Sprintf("📝 [%s] %s", ts, cleanMsg)

	// Smart auto-expand on major deployment steps or errors
	autoExpand := strings.Contains(msg, "[BƯỚC") ||
		strings.Contains(msg, "TRIỂN KHAI") ||
		strings.Contains(msg, "DEPLOY") ||
		strings.Contains(msg, "❌") ||
		strings.Contains(msg, "✗") ||
		strings.Contains(msg, "Lỗi")

	fyne.Do(func() {
		lt.lblLatest.SetText(preview)
		lt.lblCount.SetText(fmt.Sprintf("[%d]", len(lt.panel.logLines)))
		if autoExpand && !lt.isExpanded {
			lt.SetExpanded(true)
		}
	})
}

func (lt *LogTray) Container() fyne.CanvasObject {
	return lt.container
}

func (lt *LogTray) SetWindow(w fyne.Window) {
	lt.window = w
	lt.panel.SetWindow(w)
}

func (lt *LogTray) Toggle() {
	lt.SetExpanded(!lt.isExpanded)
}

func (lt *LogTray) SetExpanded(expanded bool) {
	lt.isExpanded = expanded
	if lt.isExpanded {
		lt.btnToggle.SetText("▼  Thu Gọn")
		lt.container.Objects = []fyne.CanvasObject{lt.fullLayout}
	} else {
		lt.btnToggle.SetText("▲  Mở Rộng")
		lt.container.Objects = []fyne.CanvasObject{lt.barLayout}
	}
	lt.container.Refresh()
	if lt.OnHeightChg != nil {
		lt.OnHeightChg(lt.isExpanded)
	}
}

func (lt *LogTray) IsExpanded() bool {
	return lt.isExpanded
}

// Log logs a message into the wrapped panel.
func (lt *LogTray) Log(msg string) {
	lt.panel.Log(msg)
}

func (lt *LogTray) UpdateLanguage() {
	if lt.isExpanded {
		lt.btnToggle.SetText("▼  " + i18n.T("log_default"))
	} else {
		lt.btnToggle.SetText("▲  " + i18n.T("log_expand"))
	}
}
