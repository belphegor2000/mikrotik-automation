# MikroTik RouterOS 7 Automation Tool 🚀

Ứng dụng tự động hóa cấu hình và quản trị thiết bị mạng **MikroTik RouterOS v7**, hỗ trợ đa nền tảng **Windows Desktop** và **Android Mobile** (tối ưu hóa cho Samsung Galaxy A56).

Được xây dựng hoàn toàn bằng ngôn ngữ **Golang** với framework giao diện đồ họa **Fyne v2** native, hỗ trợ kết nối qua **REST API** (RouterOS 7) và **SSH CLI**.

---

## 🌟 Tính Năng Nổi Bật

- 📊 **Dashboard Giám Sát**: Xem thông số phần cứng CPU, RAM, ổ cứng, phiên bản RouterOS, uptime và trạng thái các cổng mạng thời gian thực.
- 🌐 **Cấu Hình Mạng Toàn Diện**:
  - Tự động quay PPPoE Client (1 đường WAN hoặc 2 đường WAN gộp tải PCC).
  - Tự động tạo Bridge VLAN (`bridgeVLAN`) và cấu hình PVID Access Port.
  - Phân đoạn mạng VLAN với DHCP Server, DHCP Pool, DNS tự động.
- 🔒 **Bảo Mật & Tường Lửa Chuẩn Doanh Nghiệp**:
  - Bảo vệ tấn công SSH / WinBox Brute Force.
  - Chống quét cổng Port Scanning, Drop Invalid Packets.
  - Bảo vệ ICMP Flood / Smurf / Ping of Death.
  - Tắt các dịch vụ không an toàn (FTP, Telnet, Bandwidth Test...).
- 🔐 **IPSec Site-to-Site VPN Đa VLAN**:
  - Hỗ trợ nhiều dải VLAN nội bộ và nhiều dải VLAN đầu xa.
  - Tự động gom địa chỉ IP vào Address List chuẩn: `LAN-Address` và `Remote-Address`.
  - Hỗ trợ kết nối đầu xa dùng IP động (Dynamic IP / Cloud DDNS / Passive Responder `0.0.0.0/0`).
  - Tự động cấu hình NAT Bypass (Accept) và Mangle Bypass PCC (Prerouting 2 chiều).
- 💾 **Sao Lưu & Phục Hồi An Toàn**: Tạo bản sao lưu `.backup` và xuất script cấu hình `.rsc` chỉ với 1 cú click.
- 🗑️ **Xoá Cấu Hình Thông Minh**: Quét và gỡ bỏ cấu hình tự động một cách an toàn mà không làm mất kết nối quản trị.
- 📱 **Giao Diện Mobile Thông Minh**:
  - Thiết kế dọc tỷ lệ 19.5:9 dành riêng cho điện thoại Android (Samsung Galaxy A56).
  - Thanh kết nối dạng Header Card thu gọn tiết kiệm diện tích.
  - Tab Nhật Ký chuyên biệt toàn màn hình.
  - Tích hợp nút xem trước `[📱 Chế độ Mobile / 💻 Chế độ Desktop]` ngay trên máy tính.

---

## 💻 Hướng Dẫn Sử Dụng Trên Windows

### Chạy ứng dụng:
1. Nhấp đúp chuột vào file `mikrotik-automation.exe` (hoặc chạy qua `.\run.bat` / `.\run.ps1`).
2. Đăng nhập hệ thống:
   - **Tài khoản**: `admin`
   - **Mật khẩu**: `Loveisdie1@`
3. Nhập địa chỉ IP Router MikroTik (mặc định: `192.168.88.1`), chọn phương thức REST API hoặc SSH rồi nhấn **🔌 Kết nối**.

### Tự biên dịch từ mã nguồn:
Yêu cầu: Go 1.22+ và trình biên dịch C/C++ GCC (MinGW-w64).
```powershell
.\build.ps1
```

---

## 📱 Hướng Dẫn Build & Cài Đặt Android APK (Samsung Galaxy A56)

### Cách 1: Tự động qua GitHub Actions (Khuyên dùng)
Dự án đã tích hợp sẵn GitHub Actions tại `.github/workflows/build-apk.yml`.
Bất cứ khi nào bạn đẩy code lên repository, GitHub sẽ tự động:
1. Nạp môi trường Java 17, Android NDK r26b và Fyne CLI.
2. Biên dịch mã máy native 64-bit `arm64-v8a` (chip Exynos 1580 của Samsung A56).
3. Đính kèm file APK vào tab **Actions > Artifacts** để bạn tải thẳng về điện thoại!

### Cách 2: Biên dịch trực tiếp trên máy tính
Yêu cầu: Java JDK 17, Android NDK và Fyne CLI.
```powershell
.\build-apk.bat
```

### Cài đặt lên điện thoại:
1. Chép file `.apk` vào thư mục `Download` trên Samsung Galaxy A56.
2. Mở ứng dụng **"File của bạn" (My Files)** và chọn file để Cài đặt.

---

## 🛡️ Giấy Phép (License)
Phát hành theo giấy phép MIT.
