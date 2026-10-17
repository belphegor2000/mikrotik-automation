package gui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
)

type logLayout struct {
	minHeight float32
}

func (l *logLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) >= 2 {
		topBar := objects[0]
		scroll := objects[1]
		tbMin := topBar.MinSize()
		topBar.Resize(fyne.NewSize(size.Width, tbMin.Height))
		topBar.Move(fyne.NewPos(0, 0))

		scrollY := tbMin.Height + 4
		scrollH := size.Height - scrollY
		if scrollH < 0 {
			scrollH = 0
		}
		scroll.Resize(fyne.NewSize(size.Width, scrollH))
		scroll.Move(fyne.NewPos(0, scrollY))
	}
}

func (l *logLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var w float32 = 400
	for _, o := range objects {
		if o.MinSize().Width > w {
			w = o.MinSize().Width
		}
	}
	return fyne.NewSize(w, l.minHeight)
}

// ReadOnlyLogEntry provides a high-contrast, read-only multi-line log viewer
// that prevents user editing without triggering Fyne's disabled-text dimming.
type ReadOnlyLogEntry struct {
	widget.Entry
}

func NewReadOnlyLogEntry() *ReadOnlyLogEntry {
	e := &ReadOnlyLogEntry{}
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapBreak
	e.TextStyle = fyne.TextStyle{Monospace: true}
	e.ExtendBaseWidget(e)
	return e
}

// TypedRune blocks keyboard character insertion.
func (e *ReadOnlyLogEntry) TypedRune(_ rune) {}

// TypedKey allows scrolling and cursor navigation but blocks editing keys.
func (e *ReadOnlyLogEntry) TypedKey(k *fyne.KeyEvent) {
	switch k.Name {
	case fyne.KeyUp, fyne.KeyDown, fyne.KeyLeft, fyne.KeyRight, fyne.KeyPageUp, fyne.KeyPageDown, fyne.KeyHome, fyne.KeyEnd:
		e.Entry.TypedKey(k)
	}
}

// TypedShortcut allows clipboard copy and select all, but blocks cut, paste, undo.
func (e *ReadOnlyLogEntry) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*fyne.ShortcutCopy); ok {
		e.Entry.TypedShortcut(s)
		return
	}
	if _, ok := s.(*fyne.ShortcutSelectAll); ok {
		e.Entry.TypedShortcut(s)
		return
	}
}

type LogPanel struct {
	window          fyne.Window
	entry           *ReadOnlyLogEntry
	scroll          *container.Scroll
	container       *fyne.Container
	headerLabel     *widget.Label
	btnExpand       *widget.Button
	btnCopy         *widget.Button
	btnClear        *widget.Button
	splitController func(mode int)
	expandedState   int // 0: normal (35%), 1: maximized (80%)
	mu              sync.Mutex
	logLines        []string
}

func NewLogPanel(w fyne.Window) *LogPanel {
	entry := NewReadOnlyLogEntry()
	scroll := container.NewVScroll(entry)

	lp := &LogPanel{
		window:        w,
		entry:         entry,
		scroll:        scroll,
		logLines:      make([]string, 0, 100),
		expandedState: 0,
	}

	lp.headerLabel = widget.NewLabelWithStyle(i18n.T("log_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	lp.btnCopy = widget.NewButtonWithIcon(i18n.T("log_copy"), theme.ContentCopyIcon(), func() {
		lp.mu.Lock()
		fullText := strings.Join(lp.logLines, "\n")
		lp.mu.Unlock()
		if lp.window != nil {
			lp.window.Clipboard().SetContent(fullText)
			lp.btnCopy.SetText(i18n.T("log_copied"))
			time.AfterFunc(2*time.Second, func() {
				fyne.Do(func() {
					lp.btnCopy.SetText(i18n.T("log_copy"))
				})
			})
		}
	})

	lp.btnClear = widget.NewButtonWithIcon(i18n.T("log_clear"), theme.DeleteIcon(), func() {
		lp.Clear()
	})

	lp.btnExpand = widget.NewButton(i18n.T("log_expand"), func() {
		lp.cycleExpand()
	})

	controls := container.NewHBox(lp.btnCopy, lp.btnClear, lp.btnExpand)
	topBar := container.NewBorder(nil, nil, lp.headerLabel, controls)

	lp.container = container.New(&logLayout{minHeight: 220}, topBar, scroll)
	return lp
}

func (l *LogPanel) SetWindow(w fyne.Window) {
	l.window = w
}

func (l *LogPanel) SetSplitController(sc func(mode int)) {
	l.splitController = sc
}

func (l *LogPanel) cycleExpand() {
	if l.splitController == nil {
		return
	}
	if l.expandedState == 0 {
		l.expandedState = 1
		l.btnExpand.SetText(i18n.T("log_default"))
		l.splitController(1) // Maximized
	} else {
		l.expandedState = 0
		l.btnExpand.SetText(i18n.T("log_expand"))
		l.splitController(0) // Default
	}
}

func (l *LogPanel) UpdateLanguage() {
	if l.headerLabel != nil {
		l.headerLabel.SetText(i18n.T("log_title"))
	}
	if l.btnCopy != nil {
		l.btnCopy.SetText(i18n.T("log_copy"))
	}
	if l.btnClear != nil {
		l.btnClear.SetText(i18n.T("log_clear"))
	}
	if l.btnExpand != nil {
		if l.expandedState == 1 {
			l.btnExpand.SetText(i18n.T("log_default"))
		} else {
			l.btnExpand.SetText(i18n.T("log_expand"))
		}
	}
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
	if len(l.logLines) > 1000 {
		l.logLines = l.logLines[len(l.logLines)-1000:]
	}

	fullText := strings.Join(l.logLines, "\n")
	fyne.Do(func() {
		l.entry.SetText(fullText)
		l.scroll.ScrollToBottom()
	})
}

func (l *LogPanel) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logLines = make([]string, 0, 100)
	fyne.Do(func() {
		l.entry.SetText("")
	})
}
