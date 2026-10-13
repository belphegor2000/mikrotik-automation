package gui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LogPanel struct {
	entry     *widget.Entry
	scroll    *container.Scroll
	container *fyne.Container
	toggleBtn *widget.Button
	collapsed bool
	mu        sync.Mutex
	logLines  []string
}

func NewLogPanel() *LogPanel {
	entry := widget.NewMultiLineEntry()
	entry.Wrapping = fyne.TextWrapBreak
	entry.Disable() // Read-only

	entry.SetMinRowsVisible(6)
	scroll := container.NewVScroll(entry)

	lp := &LogPanel{
		entry:  entry,
		scroll: scroll,
	}

	headerLabel := widget.NewLabelWithStyle("📝 Nhật Ký Hoạt Động (Logs)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	lp.toggleBtn = widget.NewButton("▼ Thu gọn", func() {
		lp.Toggle()
	})

	topBar := container.NewBorder(nil, nil, headerLabel, lp.toggleBtn)
	lp.container = container.NewBorder(topBar, nil, nil, nil, scroll)
	return lp
}

func (l *LogPanel) Container() fyne.CanvasObject {
	return l.container
}

func (l *LogPanel) Log(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ts := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s] %s", ts, msg)
	l.logLines = append(l.logLines, line)
	if len(l.logLines) > 500 {
		l.logLines = l.logLines[len(l.logLines)-500:]
	}

	fullText := strings.Join(l.logLines, "\n")
	fyne.Do(func() {
		l.entry.SetText(fullText)
		l.scroll.ScrollToBottom()
	})
}

func (l *LogPanel) Toggle() {
	l.collapsed = !l.collapsed
	if l.collapsed {
		l.scroll.Hide()
		l.toggleBtn.SetText("▲ Mở rộng")
	} else {
		l.scroll.Show()
		l.toggleBtn.SetText("▼ Thu gọn")
	}
	l.container.Refresh()
}
