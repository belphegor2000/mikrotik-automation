package i18n

import (
	"fmt"
	"sync"
)

const (
	LangVI = "vi"
	LangEN = "en"
)

var (
	mu          sync.RWMutex
	currentLang = LangVI
	subscribers []func()
)

// SetLang sets the current language ("vi" or "en").
func SetLang(lang string) {
	mu.Lock()
	if lang == LangEN {
		currentLang = LangEN
	} else {
		currentLang = LangVI
	}
	mu.Unlock()
}

// GetLang returns the current language.
func GetLang() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// ToggleLang switches between VI and EN and returns the new language.
func ToggleLang() string {
	mu.Lock()
	if currentLang == LangVI {
		currentLang = LangEN
	} else {
		currentLang = LangVI
	}
	newLang := currentLang
	mu.Unlock()
	Notify()
	return newLang
}

// Subscribe registers a listener callback invoked when language changes.
func Subscribe(cb func()) {
	mu.Lock()
	defer mu.Unlock()
	if cb != nil {
		subscribers = append(subscribers, cb)
	}
}

// Notify triggers all registered language change listeners.
func Notify() {
	mu.RLock()
	cbs := make([]func(), len(subscribers))
	copy(cbs, subscribers)
	mu.RUnlock()

	for _, cb := range cbs {
		cb()
	}
}

// T retrieves translated string by key with optional format arguments.
func T(key string, args ...interface{}) string {
	mu.RLock()
	lang := currentLang
	mu.RUnlock()

	var dict map[string]string
	if lang == LangEN {
		dict = dictEN
	} else {
		dict = dictVI
	}

	val, found := dict[key]
	if !found {
		// Fallback to Vietnamese
		val, found = dictVI[key]
		if !found {
			val = key
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

// Dictionary definitions
var dictVI = map[string]string{
	// Common & Navigation
	"app_title":       "Mikrotik RouterOS 7 — Automation Tool",
	"lang_toggle_btn": "🌐 English",
	"theme_light":     "☀️ Sáng",
	"theme_dark":      "🌙 Tối",
	"btn_apply":       "Áp dụng",
	"btn_cancel":      "Huỷ",
	"btn_delete":      "🗑️ Xoá",
	"status_success":  "Thành công",
	"status_error":    "Lỗi",

	// Login & Auth
	"login_header":        "🔐 MIKROTIK ROUTEROS 7",
	"login_sub":           "Automation & Network Controller",
	"login_user":          "Tài khoản quản trị",
	"login_user_label":    "Tài khoản:",
	"login_pass":          "Mật khẩu",
	"login_pass_label":    "Mật khẩu:",
	"login_btn":           "✓ ĐĂNG NHẬP HỆ THỐNG",
	"login_err_empty":     "⚠️ Vui lòng nhập đầy đủ tài khoản và mật khẩu!",
	"login_err_fail":      "❌ Tài khoản hoặc mật khẩu không chính xác!",
	"login_hint":          "💡 Mặc định: admin / Loveisdie1@",
	"login_chg_pass_btn":  "🔑 Đổi Mật Khẩu",
	"chg_pass_title":      "Đổi Mật Khẩu Quản Trị",
	"chg_pass_old":        "Mật khẩu hiện tại:",
	"chg_pass_new":        "Mật khẩu mới (≥ 6 ký tự):",
	"chg_pass_confirm":    "Xác nhận mật khẩu mới:",
	"chg_pass_btn":        "Lưu Mật Khẩu Mới",
	"chg_pass_err_match":  "Mật khẩu xác nhận không khớp!",
	"chg_pass_success":    "✓ Đổi mật khẩu thành công! Hãy đăng nhập lại bằng mật khẩu mới.",

	// Connection Bar
	"conn_proto":      "Phương thức:",
	"conn_ip":         "Địa chỉ IP:",
	"conn_port":       "Cổng:",
	"conn_user":       "Tài khoản:",
	"conn_pass":       "Mật khẩu:",
	"conn_btn":        "🔌 Kết nối",
	"conn_btn_ing":    "⏳ Đang kết nối...",
	"conn_btn_done":   "✓ Đã kết nối",
	"conn_disconn":    "❌ Ngắt kết nối",
	"conn_not":        "⚪ Chưa kết nối",
	"conn_trying":     "🟡 Đang kết nối...",
	"conn_failed":     "🔴 Kết nối thất bại",
	"conn_off":        "⚪ Đã ngắt kết nối",

	// Main Tabs
	"tab_dashboard": "📊 Dashboard",
	"tab_network":   "🌐 Cấu Hình Mạng",
	"tab_queue":     "⚡ Băng Thông",
	"tab_security":  "🔒 Bảo Mật",
	"tab_ipsec":     "🔐 IPSec VPN",
	"tab_backup":    "💾 Sao Lưu & Phục Hồi",
	"tab_clear":     "🗑️ Xoá Cấu Hình",

	// Log Panel
	"log_title":      "📝 Nhật Ký Hoạt Động (Logs)",
	"log_copy":       "📋 Sao chép",
	"log_copied":     "✓ Đã chép!",
	"log_clear":      "🗑️ Xoá log",
	"log_expand":     "▲ Mở rộng",
	"log_default":    "▼ Mặc định",

	// Dashboard
	"dash_title":        "📊 Tổng Quan Hệ Thống RouterOS",
	"dash_refresh":      "🔄 Làm Mới Thông Tin",
	"dash_loading":      "⏳ Đang tải...",
	"dash_card_res":     "Thông Số Phần Cứng",
	"dash_cpu":          "Tải CPU:",
	"dash_ram":          "Bộ nhớ RAM:",
	"dash_hdd":          "Ổ đĩa HDD:",
	"dash_storage":      "Ổ Cứng:",
	"dash_uptime":       "Thời gian chạy (Uptime):",
	"dash_card_dev":     "Thông Tin Thiết Bị",
	"dash_model":        "Model Router:",
	"dash_version":      "Phiên bản RouterOS:",
	"dash_arch":         "Kiến trúc:",
	"dash_col_name":     "Tên Cổng",
	"dash_col_type":     "Loại",
	"dash_col_status":   "Trạng Thái",
	"dash_col_mac":      "Địa Chỉ MAC",
	"dash_col_comment":  "Ghi Chú",
	"dash_link_up":      "🟢 Link Up",
	"dash_link_down":    "🔴 Link Down",

	// Tab Network
	"net_title":             "🌐 Cấu Hình Mạng Toàn Diện",
	"net_sub_pppoe":         "1. PPPoE WAN",
	"net_sub_bridge":        "2. Bridge LAN",
	"net_sub_vlan":          "3. VLANs & Subnets",
	"net_pppoe_title":       "Cấu Hình Đường Truyền PPPoE WAN:",
	"net_pppoe_1wan":        "1 đường (ether1)",
	"net_pppoe_2wan":        "2 đường (ether1 + ether2) + PCC",
	"net_pppoe_wan1":        "WAN 1 (ether1)",
	"net_pppoe_wan2":        "WAN 2 (ether2)",
	"net_pppoe_user":        "Tài khoản:",
	"net_pppoe_pass":        "Mật khẩu:",
	"net_pppoe_hint":        "Khi chọn 2 đường: Hệ thống tự động tạo Mangle PCC, Dual Default Routes & NAT masquerade.",
	"net_bridge_title":      "Cấu Hình Bridge LAN (bridgeVLAN):",
	"net_bridge_desc":       "• Tạo bridge trung tâm 'bridgeVLAN' và gán các cổng Ethernet còn lại.\n• Hỗ trợ mọi dải IP tùy biến (10.x, 172.x, 192.168.x...). Tự tính Gateway và DHCP Pool.\n• Nếu đổi IP khác IP hiện tại, hệ thống sẽ tự động chuyển kết nối sang IP mới.",
	"net_bridge_ip_label":   "Địa chỉ IP Bridge (CIDR):",
	"net_bridge_calc_label": "Thông số tính toán:",
	"net_bridge_dhcp_label": "Dịch vụ DHCP Server:",
	"net_bridge_dhcp_chk":   "Tự động cấu hình DHCP Server cho Bridge",
	"net_bridge_dns_label":  "DNS Servers:",
	"net_bridge_check_port": "🔍 Kiểm tra danh sách Port",
	"net_bridge_ports_init": "Nhấn 'Kiểm tra danh sách Port' để xem các cổng sẽ gán vào Bridge.",
	"net_vlan_title":        "Quản Lý Phân Đoạn Mạng VLANs (5 dòng mặc định, tự tính Gateway):",
	"net_vlan_col_id":       "VLAN ID",
	"net_vlan_col_name":     "Tên VLAN",
	"net_vlan_col_cidr":     "Network (CIDR)",
	"net_vlan_col_gw":       "Gateway (Tự tính)",
	"net_vlan_col_action":   "Thao tác",
	"net_vlan_add_row":      "➕ Thêm dòng VLAN",
	"net_vlan_clear_all":    "🗑️ Xoá tất cả",
	"net_btn_apply_step":    "⚙️ Áp dụng mục hiện tại",
	"net_btn_deploy_all":    "🚀 Triển Khai Toàn Diện (WAN ➔ Bridge ➔ VLAN)",

	// Tab Security
	"sec_title":          "🔒 Thiết Lập Bảo Mật & Tường Lửa RouterOS 7",
	"sec_desc":           "Tối ưu hoá an ninh hệ thống chống dò quét mật khẩu và tấn công mạng.",
	"sec_srv_title":      "1. Vô Hiệu Hoá Dịch Vụ Không Dùng",
	"sec_srv_desc":       "Tắt các dịch vụ insecure (telnet, ftp, www-ssl, api, api-ssl). Chỉ giữ lại Webfig và SSH an toàn.",
	"sec_srv_btn":        "🚫 Tắt Dịch Vụ Không Dùng",
	"sec_fw_title":       "2. Tường Lửa Stateful Firewall Filter",
	"sec_fw_desc":        "Thiết lập bộ luật chuẩn bảo vệ Router (chấp nhận established/related, chặn invalid, chặn truy cập từ WAN).",
	"sec_fw_btn":         "🛡️ Kích Hoạt Firewall Filter",
	"sec_bf_title":       "3. Chống Dò Mật Khẩu (Brute Force Protection)",
	"sec_bf_desc":        "Quét nhật ký RouterOS, phát hiện đăng nhập sai quá 5 lần trong 6 phút và tự động đưa IP vào Blacklist trong 30 ngày.",
	"sec_bf_btn":         "⚡ Kích Hoạt Chống Brute Force",
	"sec_btn_all":        "🛡️ ÁP DỤNG TOÀN DIỆN BẢO MẬT",

	// Tab IPSec
	"ipsec_title":        "🔐 Cấu Hình IPSec VPN Site-to-Site (IKEv2 PSK & Đa VLAN)",
	"ipsec_desc":         "Tạo đường hầm mã hoá VPN IKEv2 kết nối các chi nhánh văn phòng:\n• Hỗ trợ kết nối đa dải mạng VLAN (1 hoặc nhiều VLAN nội bộ <--> 1 hoặc nhiều VLAN đầu xa).\n• Tự động sinh ma trận đầy đủ các chính sách IPSec Policy và NAT Bypass rules.\n• Hỗ trợ IP động qua Tên miền DDNS (FQDN) hoặc chế độ Passive Responder (0.0.0.0/0).",
	"ipsec_peer_label":   "IP Router Đầu Xa (Peer / DDNS):",
	"ipsec_passive_chk":  "Đầu xa dùng IP động (Chờ đầu xa kết nối - Passive Responder 0.0.0.0/0)",
	"ipsec_local_addr":   "Local Address: Tự động lấy IP PPPoE ether1 (hoặc tự điền)",
	"ipsec_cloud_ddns":   "🌐 Lấy Cloud DDNS Router Này",
	"ipsec_psk_label":    "Pre-Shared Key (Mật khẩu):",
	"ipsec_local_label":  "Dải Mạng Nội Bộ (Local Subnets):",
	"ipsec_scan_btn":     "🔍 Quét VLAN từ Router",
	"ipsec_from_vlan":    "📋 Lấy từ VLAN ứng dụng",
	"ipsec_remote_label": "Dải Mạng Đầu Xa (Remote Subnets):",
	"ipsec_apply_btn":    "🔐 Áp Dụng Cấu Hình IPSec VPN",

	// Tab Backup
	"bk_title":         "💾 Sao Lưu & Phục Hồi Cấu Hình Router",
	"bk_desc":          "Tạo các bản sao lưu an toàn trước khi thực hiện thay đổi cấu hình mạng.",
	"bk_create_title":   "Tạo Bản Sao Lưu Mới (.backup)",
	"bk_name_label":    "Tên file backup:",
	"bk_pass_label":    "Mật khẩu mã hoá (tuỳ chọn):",
	"bk_create_btn":    "💾 Tạo File Backup",
	"bk_export_title":  "Xuất Cấu Hình Script (.rsc)",
	"bk_export_btn":    "📜 Xuất File Script Cấu Hình (.rsc)",
	"bk_list_title":    "Danh Sách File Trên Router:",
	"bk_refresh_btn":   "🔄 Tải Lại Danh Sách File",
	"bk_col_name":      "Tên File",
	"bk_col_size":      "Dung Lượng",
	"bk_col_date":      "Ngày Tạo",
	"bk_col_action":    "Thao Tác",
	"bk_restore_btn":   "Phục Hồi",

	"dash_reboot":               "⚠️ Khởi Động Lại Router (Reboot)",
	"dash_iface_title":          "Danh Sách Cổng Mạng (Interfaces):",
	"dash_reboot_confirm_title": "Khởi Động Lại",
	"dash_reboot_confirm_msg":   "Bạn có chắc chắn muốn khởi động lại Router không? Mạng sẽ tạm thời gián đoạn.",

	// Tab Clear
	"clr_title":         "🗑️ Dọn Dẹp & Xoá Cấu Hình RouterOS 7",
	"clr_warn_title":    "⚠️ CẢNH BÁO NGUY HIỂM",
	"clr_warn_desc":     "⚠️ Chức năng này sẽ dọn dẹp và xoá an toàn tất cả cấu hình do công cụ tạo ra theo 15 bước tuần tự:\n  1. DHCP Leases & Networks\n  2. DHCP Servers & IP Pools\n  3. IP Addresses của VLAN & Bridge\n  4. Bridge VLAN table & VLAN interfaces\n  5. Bridge bridgeVLAN và các port gán kèm\n  6. PPPoE WAN clients (pppoe-wan1, pppoe-wan2)\n  7. Firewall Filter, Mangle và NAT rules có tag [MK-AUTO]\n  8. Routing Tables và Custom Routes\n  9. IPSec VPN Policies, Identities, Peers, Profiles\n\nCác cấu hình mặc định khác của router sẽ được giữ nguyên an toàn.",
	"clr_confirm_chk":   "Tôi hiểu rõ và xác nhận muốn xoá toàn bộ cấu hình tự động",
	"clr_btn":           "🗑️ Xoá Cấu Hình Tự Động (15 Bước)",
	"clr_btn_ing":       "⏳ Đang dọn dẹp...",
	"clr_confirm_title": "Xác Nhận Xoá Cấu Hình",
	"clr_confirm_msg":   "Bạn có chắc chắn muốn chạy quy trình xoá 15 bước không?",

	// Backup prompts
	"bk_restore_confirm_title": "Khôi Phục Cấu Hình",
	"bk_restore_confirm_msg":   "Bạn có chắc chắn muốn phục hồi cấu hình từ '%s' không?\nRouter sẽ tự động khởi động lại!",

	// Tab Queue
	"queue_title":                "⚡ Quản Lý Băng Thông (Simple Queues)",
	"queue_form_title":           "Thêm Giới Hạn Băng Thông Mới:",
	"queue_name_label":           "Tên Queue:",
	"queue_target_label":         "Địa chỉ Target (IP / Subnet):",
	"queue_upload_label":         "Tốc độ Upload tối đa:",
	"queue_download_label":       "Tốc độ Download tối đa:",
	"queue_from_vlan_btn":        "📋 Chọn Subnet từ VLAN",
	"queue_add_btn":              "➕ Thêm Giới Hạn Băng Thông",
	"queue_list_title":           "Danh Sách Simple Queues Đang Chạy:",
	"queue_refresh_btn":          "🔄 Tải lại danh sách",
	"queue_col_name":             "Tên Queue",
	"queue_col_target":           "Target",
	"queue_col_upload":           "Upload Limit",
	"queue_col_download":         "Download Limit",
	"queue_col_action":           "Thao Tác",
	"queue_delete_btn":           "Xoá",
	"queue_delete_confirm_title": "Xác Nhận Xoá Queue",
	"queue_delete_confirm_msg":   "Bạn có chắc chắn muốn xoá Simple Queue '%s' trên Router không?",
}

var dictEN = map[string]string{
	// Common & Navigation
	"app_title":       "Mikrotik RouterOS 7 — Automation Tool",
	"lang_toggle_btn": "🌐 Tiếng Việt",
	"theme_light":     "☀️ Light",
	"theme_dark":      "🌙 Dark",
	"btn_apply":       "Apply",
	"btn_cancel":      "Cancel",
	"btn_delete":      "🗑️ Delete",
	"status_success":  "Success",
	"status_error":    "Error",

	// Login & Auth
	"login_header":        "🔐 MIKROTIK ROUTEROS 7",
	"login_sub":           "Automation & Network Controller",
	"login_user":          "Admin Username",
	"login_user_label":    "Username:",
	"login_pass":          "Password",
	"login_pass_label":    "Password:",
	"login_btn":           "✓ LOG IN SYSTEM",
	"login_err_empty":     "⚠️ Please enter both username and password!",
	"login_err_fail":      "❌ Incorrect username or password!",
	"login_hint":          "💡 Default: admin / Loveisdie1@",
	"login_chg_pass_btn":  "🔑 Change Password",
	"chg_pass_title":      "Change Admin Password",
	"chg_pass_old":        "Current Password:",
	"chg_pass_new":        "New Password (≥ 6 chars):",
	"chg_pass_confirm":    "Confirm New Password:",
	"chg_pass_btn":        "Save New Password",
	"chg_pass_err_match":  "New passwords do not match!",
	"chg_pass_success":    "✓ Password changed successfully! Please log in again.",

	// Connection Bar
	"conn_proto":      "Protocol:",
	"conn_ip":         "IP Address:",
	"conn_port":       "Port:",
	"conn_user":       "Username:",
	"conn_pass":       "Password:",
	"conn_btn":        "🔌 Connect",
	"conn_btn_ing":    "⏳ Connecting...",
	"conn_btn_done":   "✓ Connected",
	"conn_disconn":    "❌ Disconnect",
	"conn_not":        "⚪ Disconnected",
	"conn_trying":     "🟡 Connecting...",
	"conn_failed":     "🔴 Connection Failed",
	"conn_off":        "⚪ Disconnected",

	// Main Tabs
	"tab_dashboard": "📊 Dashboard",
	"tab_network":   "🌐 Network Setup",
	"tab_queue":     "⚡ Bandwidth",
	"tab_security":  "🔒 Security",
	"tab_ipsec":     "🔐 IPSec VPN",
	"tab_backup":    "💾 Backup & Restore",
	"tab_clear":     "🗑️ Reset Config",

	// Log Panel
	"log_title":      "📝 Activity Logs",
	"log_copy":       "📋 Copy",
	"log_copied":     "✓ Copied!",
	"log_clear":      "🗑️ Clear Logs",
	"log_expand":     "▲ Maximize",
	"log_default":    "▼ Default",

	// Dashboard
	"dash_title":        "📊 RouterOS System Overview",
	"dash_refresh":      "🔄 Refresh System Info",
	"dash_loading":      "⏳ Loading...",
	"dash_card_res":     "Hardware Resources",
	"dash_cpu":          "CPU Usage:",
	"dash_ram":          "RAM Memory:",
	"dash_hdd":          "HDD Storage:",
	"dash_storage":      "Storage:",
	"dash_uptime":       "System Uptime:",
	"dash_card_dev":     "Device Information",
	"dash_model":        "Router Model:",
	"dash_version":      "RouterOS Version:",
	"dash_arch":         "Architecture:",
	"dash_col_name":     "Name",
	"dash_col_type":     "Type",
	"dash_col_status":   "Status",
	"dash_col_mac":      "MAC Address",
	"dash_col_comment":  "Comment",
	"dash_link_up":      "🟢 Link Up",
	"dash_link_down":    "🔴 Link Down",

	// Tab Network
	"net_title":             "🌐 Comprehensive Network Setup",
	"net_sub_pppoe":         "1. PPPoE WAN",
	"net_sub_bridge":        "2. Bridge LAN",
	"net_sub_vlan":          "3. VLANs & Subnets",
	"net_pppoe_title":       "PPPoE WAN Uplink Configuration:",
	"net_pppoe_1wan":        "1 Line (ether1)",
	"net_pppoe_2wan":        "2 Lines (ether1 + ether2) + PCC",
	"net_pppoe_wan1":        "WAN 1 (ether1)",
	"net_pppoe_wan2":        "WAN 2 (ether2)",
	"net_pppoe_user":        "Username:",
	"net_pppoe_pass":        "Password:",
	"net_pppoe_hint":        "When 2 lines selected: Automatically creates Mangle PCC, Dual Default Routes & NAT masquerade.",
	"net_bridge_title":      "Bridge LAN Configuration (bridgeVLAN):",
	"net_bridge_desc":       "• Creates central bridge 'bridgeVLAN' and assigns available Ethernet ports.\n• Supports any custom IP subnet (10.x, 172.x, 192.168.x...). Computes Gateway and DHCP Pool.\n• If IP changed from current, system will automatically migrate and reconnect.",
	"net_bridge_ip_label":   "Bridge IP Address (CIDR):",
	"net_bridge_calc_label": "Calculated Parameters:",
	"net_bridge_dhcp_label": "DHCP Server Service:",
	"net_bridge_dhcp_chk":   "Automatically configure DHCP Server for Bridge",
	"net_bridge_dns_label":  "DNS Servers:",
	"net_bridge_check_port": "🔍 Check Available Ports",
	"net_bridge_ports_init": "Click 'Check Available Ports' to see interfaces to be added to Bridge.",
	"net_vlan_title":        "VLAN Segments Management (5 default rows, auto Gateway):",
	"net_vlan_col_id":       "VLAN ID",
	"net_vlan_col_name":     "VLAN Name",
	"net_vlan_col_cidr":     "Network (CIDR)",
	"net_vlan_col_gw":       "Gateway (Auto)",
	"net_vlan_col_action":   "Actions",
	"net_vlan_add_row":      "➕ Add VLAN Row",
	"net_vlan_clear_all":    "🗑️ Clear All",
	"net_btn_apply_step":    "⚙️ Apply Current Step",
	"net_btn_deploy_all":    "🚀 Full Deploy (WAN ➔ Bridge ➔ VLAN)",

	// Tab Security
	"sec_title":          "🔒 RouterOS 7 Security Hardening & Firewall",
	"sec_desc":           "System security optimizations against password brute-force and network threats.",
	"sec_srv_title":      "1. Disable Unused Insecure Services",
	"sec_srv_desc":       "Disable insecure services (telnet, ftp, www-ssl, api, api-ssl). Retains only secure Webfig and SSH.",
	"sec_srv_btn":        "🚫 Disable Insecure Services",
	"sec_fw_title":       "2. Stateful Firewall Filter Rules",
	"sec_fw_desc":        "Standard router protection rules (accept established/related, drop invalid, drop WAN incoming).",
	"sec_fw_btn":         "🛡️ Activate Firewall Filter",
	"sec_bf_title":       "3. Brute Force Login Protection",
	"sec_bf_desc":        "Scans RouterOS logs, detects >5 failed logins in 6m, and automatically blacklists the IP for 30 days.",
	"sec_bf_btn":         "⚡ Activate Brute Force Defense",
	"sec_btn_all":        "🛡️ APPLY COMPREHENSIVE SECURITY",

	// Tab IPSec
	"ipsec_title":        "🔐 IPSec Site-to-Site VPN (IKEv2 PSK & Multi-VLAN)",
	"ipsec_desc":         "Encrypted IKEv2 VPN tunnel connecting branch offices:\n• Supports multi-VLAN interconnect (1 or more local VLANs <--> 1 or more remote VLANs).\n• Generates complete matrix of IPSec Policies and NAT Bypass rules.\n• Supports dynamic IP via MikroTik Cloud DDNS (FQDN) or Passive Responder (0.0.0.0/0).",
	"ipsec_peer_label":   "Remote Router IP (Peer / DDNS):",
	"ipsec_passive_chk":  "Remote uses Dynamic IP (Wait for remote - Passive Responder 0.0.0.0/0)",
	"ipsec_local_addr":   "Local Address: Auto-detected from PPPoE ether1 (or custom IP)",
	"ipsec_cloud_ddns":   "🌐 Get Router Cloud DDNS",
	"ipsec_psk_label":    "Pre-Shared Key (Secret):",
	"ipsec_local_label":  "Local Subnets:",
	"ipsec_scan_btn":     "🔍 Scan VLANs from Router",
	"ipsec_from_vlan":    "📋 Load from App VLANs",
	"ipsec_remote_label": "Remote Subnets:",
	"ipsec_apply_btn":    "🔐 Apply IPSec VPN Configuration",

	// Tab Backup
	"bk_title":         "💾 Router Backup & Configuration Restore",
	"bk_desc":          "Create safe system backups before applying network changes.",
	"bk_create_title":   "Create New Binary Backup (.backup)",
	"bk_name_label":    "Backup file name:",
	"bk_pass_label":    "Encryption password (optional):",
	"bk_create_btn":    "💾 Create Backup File",
	"bk_export_title":  "Export Configuration Script (.rsc)",
	"bk_export_btn":    "📜 Export Configuration Script (.rsc)",
	"bk_list_title":    "Files Stored on Router:",
	"bk_refresh_btn":   "🔄 Refresh File List",
	"bk_col_name":      "File Name",
	"bk_col_size":      "File Size",
	"bk_col_date":      "Creation Date",
	"bk_col_action":    "Actions",
	"bk_restore_btn":   "Restore",

	"dash_reboot":               "⚠️ Reboot Router",
	"dash_iface_title":          "Network Interfaces List:",
	"dash_reboot_confirm_title": "Reboot Router",
	"dash_reboot_confirm_msg":   "Are you sure you want to reboot the Router? Network connection will be briefly interrupted.",

	// Tab Clear
	"clr_title":         "🗑️ Reset RouterOS 7 Configuration",
	"clr_warn_title":    "⚠️ CRITICAL WARNING",
	"clr_warn_desc":     "⚠️ This feature will safely clean and remove all configuration created by this tool across 15 sequential steps:\n  1. DHCP Leases & Networks\n  2. DHCP Servers & IP Pools\n  3. VLAN & Bridge IP Addresses\n  4. Bridge VLAN table & VLAN interfaces\n  5. Bridge bridgeVLAN and associated member ports\n  6. PPPoE WAN clients (pppoe-wan1, pppoe-wan2)\n  7. Firewall Filter, Mangle and NAT rules with tag [MK-AUTO]\n  8. Routing Tables and Custom Routes\n  9. IPSec VPN Policies, Identities, Peers, Profiles\n\nOther default router configurations remain completely safe.",
	"clr_confirm_chk":   "I understand and confirm that I want to remove all automated configuration",
	"clr_btn":           "🗑️ Run 15-Step Automated Reset",
	"clr_btn_ing":       "⏳ Resetting configuration...",
	"clr_confirm_title": "Confirm Configuration Reset",
	"clr_confirm_msg":   "Are you sure you want to execute the 15-step configuration cleanup?",

	// Backup prompts
	"bk_restore_confirm_title": "Restore Configuration",
	"bk_restore_confirm_msg":   "Are you sure you want to restore configuration from '%s'?\nRouter will automatically reboot!",

	// Tab Queue
	"queue_title":                "⚡ Bandwidth Management (Simple Queues)",
	"queue_form_title":           "Add New Bandwidth Limit:",
	"queue_name_label":           "Queue Name:",
	"queue_target_label":         "Target Address (IP / Subnet):",
	"queue_upload_label":         "Max Upload Limit:",
	"queue_download_label":       "Max Download Limit:",
	"queue_from_vlan_btn":        "📋 Pick Subnet from VLAN",
	"queue_add_btn":              "➕ Add Bandwidth Queue",
	"queue_list_title":           "Active Simple Queues List:",
	"queue_refresh_btn":          "🔄 Refresh List",
	"queue_col_name":             "Queue Name",
	"queue_col_target":           "Target",
	"queue_col_upload":           "Upload Limit",
	"queue_col_download":         "Download Limit",
	"queue_col_action":           "Actions",
	"queue_delete_btn":           "Delete",
	"queue_delete_confirm_title": "Confirm Queue Deletion",
	"queue_delete_confirm_msg":   "Are you sure you want to delete Simple Queue '%s' from the Router?",
}
