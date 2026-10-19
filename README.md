# MikroTik RouterOS 7 — Automation Tool & Enterprise Controller 🚀

[![Build Windows Desktop](https://github.com/belphegor2000/mikrotik-automation/actions/workflows/build-desktop.yml/badge.svg)](https://github.com/belphegor2000/mikrotik-automation/actions/workflows/build-desktop.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](https://golang.org)
[![Fyne Version](https://img.shields.io/badge/GUI-Fyne%20v2.8-blue)](https://fyne.io)
[![RouterOS](https://img.shields.io/badge/MikroTik-RouterOS%20v7-red?logo=mikrotik)](https://mikrotik.com)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Phần mềm máy tính chuyên nghiệp (**Windows Desktop Native**) xây dựng hoàn toàn bằng **Golang** và framework đồ họa **Fyne v2**, được thiết kế để tự động hóa toàn diện quy trình triển khai, tối ưu hóa băng thông, an ninh mạng và quản trị hệ thống thiết bị mạng **MikroTik RouterOS 7**.

Ứng dụng hỗ trợ giao tiếp chuẩn qua **REST API** (RouterOS 7 native) và **SSH CLI**, tương thích với mọi dòng router MikroTik từ văn phòng nhỏ (hEX, hAP, RB750Gr3, RB5009) đến trung tâm dữ liệu (CCR Series).

---

## 🌟 Tính Năng Nổi Bật

### 1. 📊 System Overview (Dashboard Giám Sát Thời Gian Thực)
- **Tài nguyên phần cứng:** Đo lường tải CPU (%), dung lượng RAM khả dụng, ổ cứng HDD, thời gian hoạt động liên tục (System Uptime).
- **Thông tin thiết bị:** Model router, phiên bản RouterOS 7, kiến trúc vi xử lý (ARM, ARM64, TILE, MIPSBE,...).
- **Trạng thái cổng mạng:** Bảng kiểm tra danh sách cổng Ethernet/SFP với địa chỉ MAC, loại cổng, ghi chú và trạng thái Link Up/Down trực quan.

### 2. 🌐 Cấu Hình Mạng Toàn Diện (Network Setup)
- **PPPoE Client:** Hỗ trợ cấu hình 1 đường WAN hoặc 2 đường WAN gộp tải cân bằng (**PCC Load Balancing**) tự động sinh Mangle rules, Default Routes và NAT Masquerade.
- **Smart Failover (Recursive Routing):** Tự động phát hiện mất mạng tầng ISP và chuyển hướng dự phòng thông minh qua DNS kiểm tra (`8.8.8.8` / `1.1.1.1`).
- **Bridge LAN (`bridgeVLAN`):** Tạo bridge trung tâm, gán các cổng Ethernet khả dụng, hỗ trợ mọi dải IP tùy biến (10.x, 172.x, 192.168.x). Tự tính toán Gateway và DHCP Pool. Tự động chuyển kết nối sang IP mới nếu đổi IP router.
- **Quản lý VLANs:** Phân đoạn mạng linh hoạt với bảng dữ liệu động, tự tính toán Gateway theo CIDR, cấp DHCP Server độc lập cho từng VLAN.
- **Định Tuyến Chính Sách (PBR - Policy Based Routing):**
  - Chỉ định luồng truy cập theo **VLAN / Subnet** hoặc theo **Tên miền / IP dịch vụ** bắt buộc đi qua đường WAN mong muốn (`to_wan1` hoặc `to_wan2`).
- **Mẫu Cấu Hình Chuẩn (JSON Templates):** Xuất toàn bộ cấu hình mạng ra file template `.json` hoặc nạp triển khai tức thì cho Router mới với **1-Click Deploy**.

### 3. ⚡ Quản Lý Băng Thông & Smart CAKE QoS (Bandwidth)
- **Simple Queue:** Giới hạn tốc độ Upload/Download theo từng IP hoặc Subnet với nút **Chọn nhanh từ VLAN**.
- **Smart CAKE QoS 1-Click (Triệt tiêu Bufferbloat):**
  - Thuật toán CAKE tiên tiến thế hệ mới trong RouterOS 7: tự động tối ưu RTT (100ms), phân loại luồng ưu tiên `diffserv4` (Gaming/VoIP/DNS luôn mượt mà), cách ly công bằng `triple-isolate` chống 1 máy chiếm dụng băng thông.
  - Tự động duy trì trần 95% line-rate loại bỏ triệt để hiện tượng giật lag, spike ping.
  - **Hỗ trợ Dual WAN:** Tự động tạo 2 hàng đợi CAKE riêng biệt cho từng interface WAN (`pppoe-wan1` và `pppoe-wan2`), đảm bảo mỗi đường truyền ISP được giới hạn độc lập ở mức 95% ngay cả khi tải không đều.

### 4. 🔒 Bảo Mật & Tường Lửa Chuẩn Doanh Nghiệp (Security)
- **Vô hiệu hoá dịch vụ không an toàn:** 1-Click tắt các service rủi ro cao (`telnet`, `ftp`, `www-ssl`, `api`, `api-ssl`), chỉ giữ lại SSH và Webfig an toàn.
- **Stateful Firewall Filter:** Thiết lập bộ luật tiêu chuẩn bảo vệ router: chấp nhận established/related/untracked, drop invalid packets, chặn truy cập trái phép từ cổng WAN.
- **Chống Dò Mật Khẩu (Brute Force Protection):** Tự động quét nhật ký đăng nhập thất bại quá 5 lần trong 6 phút và khóa địa chỉ IP vào Blacklist trong 30 ngày.
- **Port Forwarding (DST-NAT) & Auto Hairpin NAT:** Mở cổng dịch vụ ra ngoài Internet và tự động cấu hình Hairpin NAT toàn mạng nội bộ, cho phép máy tính trong mạng LAN truy cập vào máy chủ cục bộ bằng IP WAN / DDNS.

### 5. 🔐 IPSec Site-to-Site VPN Đa VLAN
- Tạo đường hầm mã hóa bảo mật IKEv2 PSK kết nối các văn phòng chi nhánh.
- Hỗ trợ kết nối đa dải mạng VLAN (1 hoặc nhiều VLAN nội bộ $\leftrightarrow$ 1 hoặc nhiều VLAN đầu xa).
- Hỗ trợ kết nối đầu xa dùng IP động qua **Tên miền DDNS (FQDN)** hoặc chế độ **Passive Responder (0.0.0.0/0)**.
- Tự động sinh ma trận chính sách IPSec Policy và NAT Bypass rules.

### 6. 💾 Sao Lưu, Phục Hồi & Xoá Cấu Hình (Backup & Cleaner)
- **Sao lưu:** Tạo bản sao lưu nhị phân an toàn `.backup` hoặc xuất toàn bộ lệnh cấu hình dạng script `.rsc`.
- **Phục hồi:** Nạp bản sao lưu và tự động khởi động lại thiết bị an toàn.
- **Xoá cấu hình an toàn (15 Bước):** Gỡ bỏ tuần tự cấu hình (VLAN, Bridge, DHCP, Firewall, Queue, Routing) mà không làm mất kết nối quản trị đang hoạt động.

---

## 🎨 Trải Nghiệm Giao Diện Người Dùng (Enterprise UI/UX)

- **Header Bar 48px hiện đại:**
  - Chip hiển thị trạng thái kết nối thời gian thực: `[ ⚪ Chưa kết nối ]` / `[ 🟢 Tên Router (Giao thức) ]`.
  - Nút chuyển nhanh ngôn ngữ: **Tiếng Việt (VI)** $\leftrightarrow$ **English (EN)** tức thì không cần khởi động lại.
  - Chuyển đổi giao diện: **Light Mode (Sáng)** $\leftrightarrow$ **Dark Mode (Tối)** với màu sắc bán trong suốt chống chói mắt.
  - Quản lý danh bạ Router (Address Book): Lưu trữ hồ sơ thiết bị, đăng nhập nhanh chỉ với 1 click.
- **Sidebar Navigation:** Menu điều hướng cố định bên trái phong cách Enterprise, chuyển đổi tab mượt mà.
- **Collapsible Log Tray:** Bảng nhật ký hoạt động cố định ở cạnh dưới, hỗ trợ thu gọn tiết kiệm diện tích (8%) hoặc mở rộng chi tiết (32%) kèm các nút **Sao chép** và **Xoá log**.

---

## 💻 Hướng Dẫn Cài Đặt & Sử Dụng

### Yêu Cầu Hệ Thống
- Hệ điều hành: Windows 10 / Windows 11 (64-bit).
- Thiết bị mạng: MikroTik RouterOS v7.1 trở lên (khuyên dùng v7.12+).
- Dịch vụ trên Router: Bật dịch vụ **REST API (www)** hoặc **SSH**.

### Chạy Ứng Dụng Nhanh (Không Cần Cài Đặt)
1. Tải về file [`mikrotik-automation.exe`](file:///E:/projects/mikrotik-automation/mikrotik-automation.exe) từ mục [Releases](https://github.com/belphegor2000/mikrotik-automation/releases) hoặc trong thư mục dự án.
2. Nhấp đúp chuột để khởi động ứng dụng trực tiếp.
3. Đăng nhập phần mềm:
   - **Tài khoản mặc định:** `admin`
   - **Mật khẩu mặc định:** `Loveisdie1@` *(có thể đổi mật khẩu sau khi đăng nhập).*
4. Nhấn **🔌 Kết Nối Router** trên thanh Header, nhập thông tin IP Router (mặc định `192.168.88.1`) và tài khoản RouterOS rồi bắt đầu cấu hình.

---

## 🛠️ Tự Biên Dịch Từ Mã Nguồn (Build from Source)

### Yêu Cầu Môi Trường
- [Go](https://go.dev/dl/) phiên bản 1.24 trở lên.
- Trình biên dịch C/C++ GCC 64-bit (ví dụ: [MinGW-w64](https://winlibs.com/)).

### Quy Trình Biên Dịch
Dự án đã tích hợp sẵn kịch bản biên dịch tối ưu hoá sản phẩm (tự động chạy unit test, nhúng icon Windows và ẩn console):

```powershell
# Cách 1: Chạy trực tiếp qua PowerShell (Khuyên dùng)
powershell -ExecutionPolicy Bypass -File .\build.ps1

# Cách 2: Nhấp đúp chuột hoặc chạy file Batch
.\build.bat
```

Sau khi hoàn tất, tệp thực thi độc lập **`mikrotik-automation.exe`** (~25.5 MB) sẽ được tạo ra sẵn sàng sử dụng.

---

## 📁 Cấu Trúc Mã Nguồn

```text
mikrotik-automation/
├── .github/workflows/          # CI/CD tự động kiểm thử và build bản Windows
│   └── build-desktop.yml
├── internal/
│   ├── api/                    # Driver giao tiếp MikroTik (REST API & SSH)
│   ├── auth/                   # Hệ thống xác thực đăng nhập người dùng (bcrypt)
│   ├── gui/                    # Giao diện đồ họa Fyne v2 (Theme, Sidebar, Header, Tabs)
│   │   ├── header_bar.go       # Thanh Header mỏng, chip trạng thái & quick actions
│   │   ├── sidebar.go          # Menu điều hướng bên trái Enterprise
│   │   ├── log_tray.go         # Bảng log co giãn phía dưới
│   │   ├── theme.go            # Hệ thống Semantic Theme Light & Dark mode
│   │   ├── tab_dashboard.go    # Màn hình Dashboard phần cứng & cổng mạng
│   │   ├── tab_network.go      # Cấu hình PPPoE, Bridge, VLAN, PBR & Templates
│   │   ├── tab_queue.go        # Simple Queue & Smart CAKE QoS 1-Click
│   │   ├── tab_security.go     # Firewall, Service Hardening, Brute Force, Hairpin NAT
│   │   ├── tab_ipsec.go        # IPSec Site-to-Site VPN đa VLAN
│   │   ├── tab_backup.go       # Sao lưu .backup & xuất cấu hình .rsc
│   │   └── tab_clear.go        # Quy trình dọn dẹp reset 15 bước
│   ├── i18n/                   # Hệ thống đa ngôn ngữ động (Tiếng Việt & English)
│   ├── modules/                # Các mô-đun nghiệp vụ RouterOS
│   │   ├── bridge/             # Tự động hóa Bridge LAN & DHCP
│   │   ├── nat/                # Port Forwarding & Hairpin NAT
│   │   ├── pppoe/              # PPPoE Client, PCC Mangle & Recursive Failover
│   │   ├── queue/              # Simple Queue & Smart CAKE QoS Engine
│   │   ├── routing/            # Policy Based Routing (PBR) Engine
│   │   ├── security/           # Stateful Firewall & Brute Force Monitor
│   │   └── vlan/               # VLAN Segmentation Engine
│   ├── orchestrator/           # Điều phối triển khai mạng đồng bộ
│   └── profile/                # Quản lý danh bạ Router & Network Templates
├── build.ps1                   # Kịch bản biên dịch Production Release
├── build.bat                   # File thực thi build tiện dụng
├── run.bat                     # File khởi chạy ứng dụng nhanh
├── FyneApp.toml                # Metadata cấu hình ứng dụng Fyne
├── icon.png                    # Biểu tượng ứng dụng (512x512)
├── icon.ico                    # Biểu tượng định dạng Windows Icon
└── main.go                     # Điểm khởi chạy chính của phần mềm
```

---

## 🛡️ Giấy Phép (License)

Dự án được phân phối theo giấy phép mã nguồn mở **MIT License**. Bạn được tự do sử dụng, chỉnh sửa và phân phối cho mục đích cá nhân lẫn thương mại.
