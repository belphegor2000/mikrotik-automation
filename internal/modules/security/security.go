package security

import (
	"fmt"
	"mikrotik-automation/internal/api"
)

const (
	AppComment              = "[MK-AUTO]"
	BlacklistAddressList    = "blacklist"
	BruteForceScriptName    = "block-brute-force"
	BruteForceSchedulerName = "run-block-brute-force"
)

const BruteForceScriptSource = `# Auto Block Brute Force Login Failures (MK-AUTO)
:local threshold 5;
:local blockTimeout 30d;
:local listName "blacklist";
:local checkWindow 6m;
:local currentTime [/system clock get time];

:local logList [:toarray ""];
:do {
  :set logList [:toarray [/log find time>($currentTime - $checkWindow) message~"login failure"]];
} on-error={
  :set logList [:toarray [/log find message~"login failure"]];
};

:if ([:len $logList] > 0) do={
  :foreach i in=$logList do={
    :local logMsg [/log get $i message];
    :local at [:find $logMsg " from "];
    :while ([:typeof [:find $logMsg " from " ($at + 1)]] = "num") do={
      :set at [:find $logMsg " from " ($at + 1)];
    };
    :if ([:typeof $at] = "num") do={
      :local rawRest [:pick $logMsg ($at + 6) [:len $logMsg]];
      
      # Strip quotes and control characters from string
      :local cleanStr "";
      :for k from=0 to=([:len $rawRest] - 1) do={
        :local ch [:pick $rawRest $k ($k + 1)];
        :if ($ch != "\"" && $ch != "'") do={
          :set cleanStr ($cleanStr . $ch);
        };
      };
      
      # Strip leading spaces
      :while ([:len $cleanStr] > 0 && [:pick $cleanStr 0 1] = " ") do={
        :set cleanStr [:pick $cleanStr 1 [:len $cleanStr]];
      };
      
      # Extract token up to first space or "via"
      :local viaPos [:find $cleanStr " via "];
      :local token "";
      :if ([:typeof $viaPos] = "num") do={
        :set token [:pick $cleanStr 0 $viaPos];
      } else={
        :local spPos [:find $cleanStr " "];
        :if ([:typeof $spPos] = "num") do={
          :set token [:pick $cleanStr 0 $spPos];
        } else={
          :set token $cleanStr;
        };
      };
      
      # Strip port if present (e.g. 192.168.1.7:54321)
      :local colonPos [:find $token ":"];
      :if ([:typeof $colonPos] = "num") do={
        :set token [:pick $token 0 $colonPos];
      };
      
      # Strictly validate as an IPv4 address
      :local parsedIP [:toip $token];
      :if ([:typeof $parsedIP] = "ip") do={
        :local ipStr [:tostr $parsedIP];
        :if ([:len [/ip firewall address-list find where list=$listName address=$parsedIP]] = 0) do={
          :local failureCount 0;
          :foreach j in=$logList do={
            :local jMsg [/log get $j message];
            :if ([:typeof [:find $jMsg $ipStr]] = "num") do={
              :set failureCount ($failureCount + 1);
            };
          };
          :if ($failureCount > $threshold) do={
            :do {
              :local cmt ("[MK-AUTO] Blocked: " . $failureCount . " failed logins");
              /ip firewall address-list add list=$listName address=$parsedIP timeout=$blockTimeout comment=$cmt;
              :log warning ("[MK-AUTO] Da them IP " . $ipStr . " vao " . $listName . " trong 30 ngay sau " . $failureCount . " lan dang nhap that bai");
            } on-error={};
          };
        };
      };
    };
  };
};`

var DefaultInsecureServices = []string{
	"api",
	"api-ssl",
	"ftp",
	"telnet",
	"www-ssl",
}

type Configurator struct {
	client api.MikrotikClient
	logFn  func(string)
}

func NewConfigurator(client api.MikrotikClient, logFn func(string)) *Configurator {
	return &Configurator{client: client, logFn: logFn}
}

func (s *Configurator) log(msg string) {
	if s.logFn != nil {
		s.logFn(msg)
	}
}

func (s *Configurator) DisableInsecureServices(services []string) (int, error) {
	if len(services) == 0 {
		services = DefaultInsecureServices
	}
	s.log("  ▶ Đang vô hiệu hoá các dịch vụ không an toàn...")
	allServices, err := s.client.Get("ip/service")
	if err != nil {
		return 0, err
	}

	targetSet := make(map[string]bool)
	for _, sv := range services {
		targetSet[sv] = true
	}

	count := 0
	for _, srv := range allServices {
		name, _ := srv["name"].(string)
		if targetSet[name] {
			id, _ := srv[".id"].(string)
			err := s.client.Set("ip/service", id, map[string]interface{}{
				"disabled": "yes",
			})
			if err != nil {
				s.log(fmt.Sprintf("    ⚠ Lỗi tắt service %s: %v", name, err))
			} else {
				s.log(fmt.Sprintf("    ✓ Đã tắt service '%s'", name))
				count++
			}
		}
	}
	return count, nil
}

func (s *Configurator) ConfigureFirewallFilters(wanInterfaces []string) error {
	if len(wanInterfaces) == 0 {
		wanInterfaces = []string{"pppoe-wan1", "pppoe-wan2"}
	}
	s.log("  ▶ Đang thiết lập các Firewall Filter Rules bảo vệ router...")

	// 0. Drop blacklisted brute-force IPs on input chain
	_, _ = s.client.Add("ip/firewall/filter", map[string]interface{}{
		"chain":            "input",
		"action":           "drop",
		"src-address-list": BlacklistAddressList,
		"comment":          AppComment + " Drop Blacklist (Brute Force)",
	})

	// 1. Accept established, related, untracked
	_, _ = s.client.Add("ip/firewall/filter", map[string]interface{}{
		"chain":            "input",
		"action":           "accept",
		"connection-state": "established,related,untracked",
		"comment":          AppComment,
	})

	// 2. Drop invalid input
	_, _ = s.client.Add("ip/firewall/filter", map[string]interface{}{
		"chain":            "input",
		"action":           "drop",
		"connection-state": "invalid",
		"comment":          AppComment,
	})

	// 3. Accept ICMP
	_, _ = s.client.Add("ip/firewall/filter", map[string]interface{}{
		"chain":    "input",
		"action":   "accept",
		"protocol": "icmp",
		"comment":  AppComment,
	})

	// 4. Drop all from WAN interfaces on input chain
	for _, wan := range wanInterfaces {
		_, _ = s.client.Add("ip/firewall/filter", map[string]interface{}{
			"chain":        "input",
			"action":       "drop",
			"in-interface": wan,
			"comment":      AppComment,
		})
	}

	// 5. Accept established, related forward
	_, _ = s.client.Add("ip/firewall/filter", map[string]interface{}{
		"chain":            "forward",
		"action":           "accept",
		"connection-state": "established,related,untracked",
		"comment":          AppComment,
	})

	// 6. Drop invalid forward
	_, _ = s.client.Add("ip/firewall/filter", map[string]interface{}{
		"chain":            "forward",
		"action":           "drop",
		"connection-state": "invalid",
		"comment":          AppComment,
	})

	// 7. Drop WAN to LAN forward if not dst-natted
	for _, wan := range wanInterfaces {
		_, _ = s.client.Add("ip/firewall/filter", map[string]interface{}{
			"chain":                "forward",
			"action":               "drop",
			"connection-nat-state": "!dstnat",
			"in-interface":         wan,
			"comment":              AppComment,
		})
	}

	s.log("✓ Hoàn thành thiết lập Firewall Filter Rules.")
	return nil
}

// ConfigureBruteForceProtection sets up script and scheduler to detect failed logins and blacklist IPs for 30 days
func (s *Configurator) ConfigureBruteForceProtection() error {
	s.log("  ▶ [1/3] Cấu hình RouterOS Script kiểm tra đăng nhập thất bại...")
	existingScripts, _ := s.client.Get("system/script")
	for _, sc := range existingScripts {
		if name, _ := sc["name"].(string); name == BruteForceScriptName {
			if id, ok := sc[".id"].(string); ok {
				_ = s.client.Remove("system/script", id)
			}
		}
	}

	_, err := s.client.Add("system/script", map[string]interface{}{
		"name":    BruteForceScriptName,
		"source":  BruteForceScriptSource,
		"comment": AppComment,
	})
	if err != nil {
		s.log(fmt.Sprintf("    ⚠ Lỗi tạo script '%s': %v", BruteForceScriptName, err))
		return err
	}
	s.log(fmt.Sprintf("    ✓ Đã tạo script '%s' thành công.", BruteForceScriptName))

	s.log("  ▶ [2/3] Cấu hình RouterOS Scheduler (chu kỳ mỗi 5 phút)...")
	existingSchedulers, _ := s.client.Get("system/scheduler")
	for _, sc := range existingSchedulers {
		if name, _ := sc["name"].(string); name == BruteForceSchedulerName {
			if id, ok := sc[".id"].(string); ok {
				_ = s.client.Remove("system/scheduler", id)
			}
		}
	}

	_, err = s.client.Add("system/scheduler", map[string]interface{}{
		"name":       BruteForceSchedulerName,
		"interval":   "5m",
		"on-event":   BruteForceScriptName,
		"start-time": "startup",
		"comment":    AppComment,
	})
	if err != nil {
		s.log(fmt.Sprintf("    ⚠ Lỗi tạo scheduler '%s': %v", BruteForceSchedulerName, err))
		return err
	}
	s.log(fmt.Sprintf("    ✓ Đã kích hoạt scheduler '%s' (mỗi 5 phút).", BruteForceSchedulerName))

	s.log("  ▶ [3/3] Thêm Firewall Drop rule cho danh sách blacklist...")
	_, err = s.client.Add("ip/firewall/filter", map[string]interface{}{
		"chain":            "input",
		"action":           "drop",
		"src-address-list": BlacklistAddressList,
		"comment":          AppComment + " Drop Blacklist (Brute Force)",
	})
	if err != nil {
		s.log(fmt.Sprintf("    ⚠ Lỗi tạo firewall drop rule: %v", err))
		return err
	}
	s.log(fmt.Sprintf("    ✓ Đã thêm Firewall Drop rule cho address-list '%s'.", BlacklistAddressList))
	s.log("✓ Hoàn tất cấu hình tự động chặn Brute Force (Blacklist 30 ngày)!")
	return nil
}
