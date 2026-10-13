package backup

import (
	"fmt"
	"os"
	"strings"

	"mikrotik-automation/internal/api"
)

type BackupItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size string `json:"size"`
}

type Service struct {
	client api.MikrotikClient
	logFn  func(string)
}

func NewService(client api.MikrotikClient, logFn func(string)) *Service {
	return &Service{client: client, logFn: logFn}
}

func (b *Service) log(msg string) {
	if b.logFn != nil {
		b.logFn(msg)
	}
}

func (b *Service) ExportRSCScript() (string, error) {
	out, err := b.client.ExecRaw("/export compact")
	if err == nil && strings.TrimSpace(out) != "" {
		return out, nil
	}

	// Try without compact
	out, err = b.client.ExecRaw("/export")
	if err == nil && strings.TrimSpace(out) != "" {
		return out, nil
	}

	// Fallback REST API
	b.log("  ▶ Đang xuất cấu hình trên router qua REST API...")
	_, err = b.client.Run("export", map[string]interface{}{"file": "mk_auto_export"})
	if err != nil {
		return "", fmt.Errorf("không thể export qua REST API: %w", err)
	}
	return "# RouterOS 7 Configuration Export\n# File đã tạo trên Router: mk_auto_export.rsc\n", nil
}

func (b *Service) SaveRSCToFile(localPath string) error {
	content, err := b.ExportRSCScript()
	if err != nil {
		return err
	}
	err = os.WriteFile(localPath, []byte(content), 0644)
	if err != nil {
		return err
	}
	b.log(fmt.Sprintf("✓ Đã lưu file cấu hình thành công: %s", localPath))
	return nil
}

func (b *Service) CreateSystemBackup(name string) (string, error) {
	cleanName := strings.ReplaceAll(name, " ", "_")
	if cleanName == "" {
		cleanName = "backup_auto"
	}
	b.log(fmt.Sprintf("  ▶ Đang tạo bản sao lưu '%s.backup' trên router...", cleanName))
	_, err := b.client.ExecRaw(fmt.Sprintf(`/system backup save name="%s"`, cleanName))
	if err != nil {
		_, err = b.client.Run("system/backup/save", map[string]interface{}{"name": cleanName})
	}
	if err != nil {
		return "", err
	}
	b.log(fmt.Sprintf("    ✓ Bản sao lưu '%s.backup' đã được lưu thành công.", cleanName))
	return cleanName + ".backup", nil
}

func (b *Service) ListRouterFiles() ([]BackupItem, error) {
	files, err := b.client.Get("file")
	if err != nil {
		return nil, err
	}
	var list []BackupItem
	for _, f := range files {
		name, _ := f["name"].(string)
		if strings.HasSuffix(name, ".backup") || strings.HasSuffix(name, ".rsc") {
			fType := "backup"
			if strings.HasSuffix(name, ".rsc") {
				fType = "script"
			}
			sizeStr := fmt.Sprintf("%v", f["size"])
			list = append(list, BackupItem{
				Name: name,
				Type: fType,
				Size: sizeStr,
			})
		}
	}
	return list, nil
}

func (b *Service) RestoreSystemBackup(name string) error {
	cleanName := strings.TrimSuffix(name, ".backup")
	b.log(fmt.Sprintf("⚠️ Đang yêu cầu phục hồi từ bản sao lưu '%s'...", cleanName))
	_, err := b.client.ExecRaw(fmt.Sprintf(`:execute "/system backup load name=\"%s\""`, cleanName))
	if err != nil {
		_, err = b.client.Run("system/backup/load", map[string]interface{}{"name": cleanName})
	}
	if err != nil {
		return err
	}
	b.log(fmt.Sprintf("✓ Đã gửi lệnh khôi phục bản sao lưu '%s'. Router sẽ tự khởi động lại!", cleanName))
	return nil
}
