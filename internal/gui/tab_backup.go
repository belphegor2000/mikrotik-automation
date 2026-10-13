package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/modules/backup"
)

type TabBackup struct {
	state          *AppState
	window         fyne.Window
	container      *fyne.Container
	entryBkName    *widget.Entry
	btnExportRSC   *widget.Button
	btnCreateBK    *widget.Button
	btnRefreshList  *widget.Button
	filesList       *widget.List
	filesListMobile *widget.List
	filesData       []backup.BackupItem
}

func NewTabBackup(state *AppState, w fyne.Window) *TabBackup {
	tb := &TabBackup{state: state, window: w}

	// 1. Export script
	tb.btnExportRSC = widget.NewButton("💾 Tải File Cấu Hình (.rsc) Về Máy Tính", func() {
		tb.exportRSC()
	})
	tb.btnExportRSC.Importance = widget.HighImportance

	// 2. Binary backup
	tb.entryBkName = widget.NewEntry()
	tb.entryBkName.SetText("backup_auto")

	tb.btnCreateBK = widget.NewButton("📦 Tạo Bản Sao Lưu (.backup) Trên Router", func() {
		tb.createBackup()
	})
	tb.btnCreateBK.Importance = widget.HighImportance

	// 3. Router files list (Desktop 3-column table)
	tb.filesData = make([]backup.BackupItem, 0)
	tb.filesList = widget.NewList(
		func() int {
			return len(tb.filesData)
		},
		func() fyne.CanvasObject {
			btn := widget.NewButton("Phục Hồi", nil)
			btn.Importance = widget.WarningImportance
			return container.NewGridWithColumns(3,
				widget.NewLabel("File Name"),
				widget.NewLabel("Type"),
				btn,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(tb.filesData) {
				return
			}
			item := tb.filesData[id]
			box := obj.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(item.Name)
			box.Objects[1].(*widget.Label).SetText(item.Type)

			btn := box.Objects[2].(*widget.Button)
			btn.OnTapped = func() {
				tb.confirmRestore(item.Name)
			}
		},
	)

	// Mobile 2-line Card
	tb.filesListMobile = widget.NewList(
		func() int {
			return len(tb.filesData)
		},
		func() fyne.CanvasObject {
			lblName := widget.NewLabelWithStyle("File Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			lblType := widget.NewLabel("Type")
			btn := widget.NewButton("Phục Hồi", nil)
			btn.Importance = widget.WarningImportance
			topRow := container.NewBorder(nil, nil, lblName, btn)
			return container.NewVBox(topRow, lblType)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(tb.filesData) {
				return
			}
			item := tb.filesData[id]
			vbox := obj.(*fyne.Container)
			topRow := vbox.Objects[0].(*fyne.Container)
			topRow.Objects[0].(*widget.Label).SetText(item.Name)
			btn := topRow.Objects[1].(*widget.Button)
			btn.OnTapped = func() {
				tb.confirmRestore(item.Name)
			}
			vbox.Objects[1].(*widget.Label).SetText(item.Type)
		},
	)

	tb.btnRefreshList = widget.NewButton("🔄 Làm mới danh sách file", func() {
		tb.refreshFiles()
	})

	return tb
}

func (tb *TabBackup) BuildDesktopContainer() fyne.CanvasObject {
	cardRSC := container.NewVBox(
		widget.NewLabelWithStyle("1. Xuất Cấu Hình Script (.rsc)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Xuất cấu hình dạng lệnh CLI text để xem lại hoặc nạp sang thiết bị khác."),
		tb.btnExportRSC,
	)

	cardBK := container.NewVBox(
		widget.NewLabelWithStyle("2. Bản Sao Lưu Toàn Diện (.backup)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Lưu trọn vẹn trạng thái hệ thống (Users, Mật khẩu, MAC, Certificates)."),
		container.NewBorder(nil, nil, widget.NewLabel("Tên file:"), nil, tb.entryBkName),
		tb.btnCreateBK,
	)

	cardFiles := container.NewBorder(
		container.NewBorder(nil, nil, widget.NewLabelWithStyle("3. Danh Sách File Sao Lưu Trên Router:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), tb.btnRefreshList),
		nil, nil, nil,
		tb.filesList,
	)

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("💾 Sao Lưu & Phục Hồi Cấu Hình", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewGridWithColumns(2, cardRSC, cardBK),
		),
		nil, nil, nil,
		cardFiles,
	)
}

func (tb *TabBackup) BuildMobileContainer() fyne.CanvasObject {
	cardRSC := container.NewVBox(
		widget.NewLabelWithStyle("1. Xuất Cấu Hình Script (.rsc)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Xuất cấu hình dạng text để xem lại."),
		container.NewPadded(tb.btnExportRSC),
	)

	cardBK := container.NewVBox(
		widget.NewLabelWithStyle("2. Bản Sao Lưu Toàn Diện (.backup)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Lưu toàn bộ users, mật khẩu, certs."),
		container.NewBorder(nil, nil, widget.NewLabelWithStyle("Tên:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, tb.entryBkName),
		container.NewPadded(tb.btnCreateBK),
	)

	cardFiles := container.NewVBox(
		widget.NewLabelWithStyle("3. File Sao Lưu Trên Router:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tb.btnRefreshList,
		container.NewGridWrap(fyne.NewSize(360, 240), tb.filesListMobile),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("💾 Sao Lưu & Phục Hồi", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		cardRSC,
		widget.NewSeparator(),
		cardBK,
		widget.NewSeparator(),
		cardFiles,
	)
}

func (tb *TabBackup) Container() fyne.CanvasObject {
	return tb.BuildDesktopContainer()
}

func (tb *TabBackup) exportRSC() {
	c := tb.state.GetClient()
	if !c.IsConnected() {
		tb.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		path := uc.URI().Path()
		_ = uc.Close()

		tb.btnExportRSC.Disable()
		go func() {
			defer fyne.Do(func() { tb.btnExportRSC.Enable() })
			svc := backup.NewService(c, tb.state.Log)
			_ = svc.SaveRSCToFile(path)
		}()
	}, tb.window)
}

func (tb *TabBackup) createBackup() {
	c := tb.state.GetClient()
	if !c.IsConnected() {
		tb.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	name := tb.entryBkName.Text
	tb.btnCreateBK.Disable()
	go func() {
		defer fyne.Do(func() { tb.btnCreateBK.Enable() })
		svc := backup.NewService(c, tb.state.Log)
		_, _ = svc.CreateSystemBackup(name)
		tb.refreshFiles()
	}()
}

func (tb *TabBackup) refreshFiles() {
	c := tb.state.GetClient()
	if !c.IsConnected() {
		tb.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	tb.btnRefreshList.Disable()
	go func() {
		svc := backup.NewService(c, tb.state.Log)
		files, err := svc.ListRouterFiles()
		fyne.Do(func() {
			tb.btnRefreshList.Enable()
			if err == nil {
				tb.filesData = files
				tb.filesList.Refresh()
				if tb.filesListMobile != nil {
					tb.filesListMobile.Refresh()
				}
				tb.state.Log(fmt.Sprintf("✓ Đã tìm thấy %d bản sao lưu trên Router.", len(files)))
			}
		})
	}()
}

func (tb *TabBackup) confirmRestore(fileName string) {
	c := tb.state.GetClient()
	if !c.IsConnected() {
		tb.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	msg := fmt.Sprintf("Bạn có chắc chắn muốn phục hồi cấu hình từ '%s' không?\nRouter sẽ tự động khởi động lại!", fileName)
	dialog.ShowConfirm("Khôi Phục Cấu Hình", msg, func(ok bool) {
		if !ok {
			return
		}
		go func() {
			svc := backup.NewService(c, tb.state.Log)
			_ = svc.RestoreSystemBackup(fileName)
		}()
	}, tb.window)
}
