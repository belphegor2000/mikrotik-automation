package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
	"mikrotik-automation/internal/modules/backup"
)

type TabBackup struct {
	state          *AppState
	window         fyne.Window
	container      *fyne.Container
	lblMainTitle   *widget.Label
	lblRSCTitle    *widget.Label
	lblRSCDesc     *widget.Label
	btnExportRSC   *widget.Button
	lblBKTitle     *widget.Label
	lblBKDesc      *widget.Label
	lblBKNameTitle *widget.Label
	entryBkName    *widget.Entry
	btnCreateBK    *widget.Button
	lblFilesTitle  *widget.Label
	btnRefreshList *widget.Button
	filesList      *widget.List
	filesData      []backup.BackupItem
}

func NewTabBackup(state *AppState, w fyne.Window) *TabBackup {
	tb := &TabBackup{state: state, window: w}

	tb.lblMainTitle = widget.NewLabelWithStyle(i18n.T("bk_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// 1. Export script
	tb.lblRSCTitle = widget.NewLabelWithStyle(i18n.T("bk_export_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tb.lblRSCDesc = widget.NewLabel("Xuất cấu hình dạng lệnh CLI text để xem lại hoặc nạp sang thiết bị khác.")
	tb.lblRSCDesc.Wrapping = fyne.TextWrapWord

	tb.btnExportRSC = widget.NewButton(i18n.T("bk_export_btn"), func() {
		tb.exportRSC()
	})
	tb.btnExportRSC.Importance = widget.HighImportance

	// 2. Binary backup
	tb.lblBKTitle = widget.NewLabelWithStyle(i18n.T("bk_create_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tb.lblBKDesc = widget.NewLabel("Lưu trọn vẹn trạng thái hệ thống (Users, Mật khẩu, MAC, Certificates).")
	tb.lblBKDesc.Wrapping = fyne.TextWrapWord

	tb.lblBKNameTitle = widget.NewLabelWithStyle(i18n.T("bk_name_label"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tb.entryBkName = widget.NewEntry()
	tb.entryBkName.SetText("backup_auto")

	tb.btnCreateBK = widget.NewButton(i18n.T("bk_create_btn"), func() {
		tb.createBackup()
	})
	tb.btnCreateBK.Importance = widget.HighImportance

	// 3. Router files list (Desktop 3-column table)
	tb.lblFilesTitle = widget.NewLabelWithStyle(i18n.T("bk_list_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tb.filesData = make([]backup.BackupItem, 0)
	tb.filesList = widget.NewList(
		func() int {
			return len(tb.filesData)
		},
		func() fyne.CanvasObject {
			btn := widget.NewButton(i18n.T("bk_restore_btn"), nil)
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
			btn.SetText(i18n.T("bk_restore_btn"))
			btn.OnTapped = func() {
				tb.confirmRestore(item.Name)
			}
		},
	)

	tb.btnRefreshList = widget.NewButton(i18n.T("bk_refresh_btn"), func() {
		tb.refreshFiles()
	})

	return tb
}

func (tb *TabBackup) BuildDesktopContainer() fyne.CanvasObject {
	cardRSC := container.NewVBox(
		tb.lblRSCTitle,
		tb.lblRSCDesc,
		tb.btnExportRSC,
	)

	cardBK := container.NewVBox(
		tb.lblBKTitle,
		tb.lblBKDesc,
		container.NewBorder(nil, nil, tb.lblBKNameTitle, nil, tb.entryBkName),
		tb.btnCreateBK,
	)

	cardFiles := container.NewBorder(
		container.NewBorder(nil, nil, tb.lblFilesTitle, tb.btnRefreshList),
		nil, nil, nil,
		tb.filesList,
	)

	return container.NewBorder(
		container.NewVBox(
			tb.lblMainTitle,
			container.NewGridWithColumns(2, cardRSC, cardBK),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		cardFiles,
	)
}

func (tb *TabBackup) Container() fyne.CanvasObject {
	return tb.BuildDesktopContainer()
}

func (tb *TabBackup) UpdateLanguage() {
	if tb.lblMainTitle != nil {
		tb.lblMainTitle.SetText(i18n.T("bk_title"))
	}
	if tb.lblRSCTitle != nil {
		tb.lblRSCTitle.SetText(i18n.T("bk_export_title"))
	}
	if tb.btnExportRSC != nil {
		tb.btnExportRSC.SetText(i18n.T("bk_export_btn"))
	}
	if tb.lblBKTitle != nil {
		tb.lblBKTitle.SetText(i18n.T("bk_create_title"))
	}
	if tb.lblBKNameTitle != nil {
		tb.lblBKNameTitle.SetText(i18n.T("bk_name_label"))
	}
	if tb.btnCreateBK != nil {
		tb.btnCreateBK.SetText(i18n.T("bk_create_btn"))
	}
	if tb.lblFilesTitle != nil {
		tb.lblFilesTitle.SetText(i18n.T("bk_list_title"))
	}
	if tb.btnRefreshList != nil {
		tb.btnRefreshList.SetText(i18n.T("bk_refresh_btn"))
	}
	if tb.filesList != nil {
		tb.filesList.Refresh()
	}
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

	msg := i18n.T("bk_restore_confirm_msg", fileName)
	dialog.ShowConfirm(i18n.T("bk_restore_confirm_title"), msg, func(ok bool) {
		if !ok {
			return
		}
		go func() {
			svc := backup.NewService(c, tb.state.Log)
			_ = svc.RestoreSystemBackup(fileName)
		}()
	}, tb.window)
}
