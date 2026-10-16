package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/modules/ipsec"
)

type TabIPSec struct {
	state          *AppState
	container      *fyne.Container
	entryPeer      *widget.Entry
	chkPassive     *widget.Check
	entryPSK       *widget.Entry
	entryLocal     *widget.Entry
	entryRemote    *widget.Entry
	btnDetect      *widget.Button
	btnFromVLANs   *widget.Button
	btnCloudDDNS   *widget.Button
	btnApply       *widget.Button
	window         fyne.Window
	peerBox        *fyne.Container
	localBox       *fyne.Container
}

func NewTabIPSec(state *AppState) *TabIPSec {
	ti := &TabIPSec{state: state}

	ti.entryPeer = widget.NewEntry()
	ti.entryPeer.SetPlaceHolder("203.0.113.1 hoặc branch.sn.mynetname.net")

	ti.chkPassive = widget.NewCheck("Đầu xa dùng IP động (Chờ đầu xa kết nối - Passive Responder 0.0.0.0/0)", func(checked bool) {
		if checked {
			ti.entryPeer.SetText("0.0.0.0/0")
			ti.entryPeer.Disable()
		} else {
			if ti.entryPeer.Text == "0.0.0.0/0" {
				ti.entryPeer.SetText("")
			}
			ti.entryPeer.Enable()
		}
	})

	ti.btnCloudDDNS = widget.NewButton("🌐 Lấy Cloud DDNS Router Này", func() {
		ti.getCloudDDNS()
	})

	peerBox := container.NewVBox(
		ti.entryPeer,
		ti.chkPassive,
		ti.btnCloudDDNS,
	)
	ti.peerBox = peerBox

	ti.entryPSK = widget.NewPasswordEntry()
	ti.entryPSK.SetPlaceHolder("KhoaBiMat123!")

	ti.entryLocal = widget.NewEntry()
	ti.entryLocal.SetPlaceHolder("192.168.10.0/24, 192.168.20.0/24 (phân cách bằng dấu phẩy hoặc xuống dòng)")
	ti.entryLocal.SetText("192.168.10.0/24")

	ti.btnDetect = widget.NewButton("🔍 Quét VLAN từ Router", func() {
		ti.detectLocalSubnets()
	})

	ti.btnFromVLANs = widget.NewButton("📋 Lấy từ VLAN ứng dụng", func() {
		ti.loadFromAppVLANs()
	})

	localButtons := container.NewGridWithColumns(2, ti.btnDetect, ti.btnFromVLANs)
	localBox := container.NewVBox(
		ti.entryLocal,
		localButtons,
	)
	ti.localBox = localBox

	ti.entryRemote = widget.NewEntry()
	ti.entryRemote.SetPlaceHolder("172.16.1.0/24, 172.16.2.0/24 (phân cách bằng dấu phẩy hoặc xuống dòng)")

	ti.btnApply = widget.NewButton("🔐 Áp Dụng Cấu Hình IPSec VPN", func() {
		ti.apply()
	})
	ti.btnApply.Importance = widget.HighImportance

	return ti
}

func (ti *TabIPSec) BuildDesktopContainer() fyne.CanvasObject {
	form := widget.NewForm(
		widget.NewFormItem("IP Router Đầu Xa (Peer / DDNS):", ti.peerBox),
		widget.NewFormItem("Pre-Shared Key (Mật khẩu):", ti.entryPSK),
		widget.NewFormItem("Dải Mạng Nội Bộ (Local Subnets):", ti.localBox),
		widget.NewFormItem("Dải Mạng Đầu Xa (Remote Subnets):", ti.entryRemote),
	)

	desc := "Tạo đường hầm mã hoá VPN IKEv2 kết nối các chi nhánh văn phòng:\n" +
		"• Hỗ trợ kết nối đa dải mạng VLAN (1 hoặc nhiều VLAN nội bộ <--> 1 hoặc nhiều VLAN đầu xa).\n" +
		"• Tự động sinh ma trận đầy đủ các chính sách IPSec Policy và NAT Bypass rules.\n" +
		"• Hỗ trợ IP động qua Tên miền DDNS (FQDN) hoặc chế độ Passive Responder (0.0.0.0/0)."

	lblDesc := widget.NewLabel(desc)
	lblDesc.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		widget.NewLabelWithStyle("🔐 Cấu Hình IPSec VPN Site-to-Site (IKEv2 PSK & Đa VLAN)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		lblDesc,
		form,
		ti.btnApply,
	)
}

func (ti *TabIPSec) Container() fyne.CanvasObject {
	return ti.BuildDesktopContainer()
}

func (ti *TabIPSec) detectLocalSubnets() {
	c := ti.state.GetClient()
	if !c.IsConnected() {
		ti.state.Log("✗ Chưa kết nối đến router!")
		return
	}
	go func() {
		cfg := ipsec.NewConfigurator(c, ti.state.Log)
		subnets, err := cfg.DetectLocalSubnets()
		fyne.Do(func() {
			if err == nil && len(subnets) > 0 {
				ti.entryLocal.SetText(strings.Join(subnets, ", "))
				ti.state.Log(fmt.Sprintf("✓ Đã phát hiện %d dải mạng VLAN từ Router: %s", len(subnets), strings.Join(subnets, ", ")))
			} else {
				ti.state.Log("⚠ Không phát hiện được dải mạng VLAN nào từ Router.")
			}
		})
	}()
}

func (ti *TabIPSec) loadFromAppVLANs() {
	if len(ti.state.VLANs) == 0 {
		ti.state.Log("⚠ Danh sách VLAN ứng dụng (Tab Network) đang trống.")
		return
	}
	var subnets []string
	for _, v := range ti.state.VLANs {
		if v.CIDR != "" {
			subnets = append(subnets, v.CIDR)
		}
	}
	normalized, err := ipsec.ParseAndNormalizeSubnets(strings.Join(subnets, ", "))
	if err == nil && len(normalized) > 0 {
		ti.entryLocal.SetText(strings.Join(normalized, ", "))
		ti.state.Log(fmt.Sprintf("✓ Đã lấy %d dải mạng VLAN từ ứng dụng: %s", len(normalized), strings.Join(normalized, ", ")))
	}
}

func (ti *TabIPSec) getCloudDDNS() {
	c := ti.state.GetClient()
	if !c.IsConnected() {
		ti.state.Log("✗ Chưa kết nối đến router!")
		return
	}
	go func() {
		cfg := ipsec.NewConfigurator(c, ti.state.Log)
		dns, err := cfg.GetCloudDDNS()
		fyne.Do(func() {
			if err == nil && dns != "" {
				ti.state.Log(fmt.Sprintf("🌐 MikroTik Cloud DDNS của Router: %s", dns))
			} else {
				ti.state.Log(fmt.Sprintf("⚠ Lỗi lấy Cloud DDNS: %v", err))
			}
		})
	}()
}

func (ti *TabIPSec) apply() {
	c := ti.state.GetClient()
	if !c.IsConnected() {
		ti.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	peer := strings.TrimSpace(ti.entryPeer.Text)
	isPassive := ti.chkPassive.Checked
	if isPassive {
		peer = "0.0.0.0/0"
	}
	psk := ti.entryPSK.Text

	if peer == "" {
		ti.state.Log("✗ Vui lòng điền IP hoặc Tên miền DDNS của Router đầu xa (hoặc chọn IP động)!")
		return
	}
	if psk == "" {
		ti.state.Log("✗ Vui lòng điền Pre-Shared Key (Mật khẩu)!")
		return
	}

	localSubnets, err := ipsec.ParseAndNormalizeSubnets(ti.entryLocal.Text)
	if err != nil {
		ti.state.Log(fmt.Sprintf("✗ Lỗi dải mạng nội bộ (Local): %v", err))
		return
	}

	remoteSubnets, err := ipsec.ParseAndNormalizeSubnets(ti.entryRemote.Text)
	if err != nil {
		ti.state.Log(fmt.Sprintf("✗ Lỗi dải mạng đầu xa (Remote): %v", err))
		return
	}

	ti.btnApply.Disable()
	go func() {
		defer fyne.Do(func() { ti.btnApply.Enable() })
		cfg := ipsec.NewConfigurator(c, ti.state.Log)
		_ = cfg.Configure(ipsec.IPSecConfig{
			PeerAddress:   peer,
			PSK:           psk,
			LocalSubnets:  localSubnets,
			RemoteSubnets: remoteSubnets,
			IsPassivePeer: isPassive,
		})
	}()
}
