package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// --- RENK VE BİÇİMLENDİRME SABİTLERİ ---
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Bold   = "\033[1m"
)

// --- 🌟 SENİN GERÇEK CONFIG.YAML FORMATINA %100 UYUMLU MODEL ---
type Config struct {
	Agent struct {
		HostID          string `yaml:"host_id"`
		IntervalSeconds int    `yaml:"interval_seconds"`
		SecretKey       string `yaml:"secret_key"`
	} `yaml:"agent"`
	Collector struct {
		URL string `yaml:"url"`
	} `yaml:"collector"`
	Checks struct {
		CPU struct {
			Enabled          bool    `yaml:"enabled"`
			ThresholdWarning float64 `yaml:"threshold_warning"`
			ThresholdError   float64 `yaml:"threshold_error"`
		} `yaml:"cpu"`
		RAM struct {
			Enabled          bool    `yaml:"enabled"`
			ThresholdWarning float64 `yaml:"threshold_warning"`
			ThresholdError   float64 `yaml:"threshold_error"`
		} `yaml:"ram"`
		Disk struct {
			Enabled          bool    `yaml:"enabled"`
			Path             string  `yaml:"path"`
			ThresholdWarning float64 `yaml:"threshold_warning"`
			ThresholdError   float64 `yaml:"threshold_error"`
		} `yaml:"disk"`
	} `yaml:"checks"`
}

type Event struct {
	ID          uint      `json:"id"`
	AgentID     string    `json:"agent_id"`
	CheckType   string    `json:"check_type"`
	MetricValue float64   `json:"metric_value"`
	Level       string    `json:"level"`
	Timestamp   time.Time `json:"timestamp"`
}

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit("Hata: Hiçbir alt komut belirtilmedi.", 1)
	}

	command := os.Args[1]

	switch command {
	case "status":
		handleStatus()
	case "config":
		if len(os.Args) < 3 || os.Args[2] != "validate" {
			printUsageAndExit("Hata: 'config' için 'validate' alt komutu gereklidir. Örn: pulsectl config validate", 1)
		}
		// Varsayılan olarak agent/config.yaml dosyasına bakar
		configPath := "agent/config.yaml"
		if len(os.Args) >= 4 {
			configPath = os.Args[3]
		}
		handleConfigValidate(configPath)
	case "report":
		handleReport()
	case "help", "--help", "-h":
		printUsageAndExit("", 0)
	default:
		printUsageAndExit(fmt.Sprintf("Hata: Bilinmeyen komut '%s'", command), 1)
	}
}

// --- 1. PULSECTL STATUS ---
func handleStatus() {
	cfg, err := loadLocalConfig("agent/config.yaml")
	if err != nil {
		fatalError("Ajan yapılandırma dosyası (agent/config.yaml) okunamadı", err)
	}

	collectorURL := strings.TrimRight(cfg.Collector.URL, "/")
	if collectorURL == "" {
		collectorURL = "http://localhost:8081"
	}

	url := fmt.Sprintf("%s/api/events?agent_id=%s", collectorURL, cfg.Agent.HostID)
	events, err := fetchEvents(url)
	if err != nil {
		fatalError("Collector ile iletişim kurulamadı. Sunucu açık mı?", err)
	}

	fmt.Printf("\n%s=== PULSEGUARD AJAN DURUMU ===%s\n", Bold, Reset)
	fmt.Printf("📦 Ajan (Host) ID : %s%s%s\n", Blue, cfg.Agent.HostID, Reset)
	fmt.Printf("🌐 Collector URL  : %s\n", collectorURL)

	if len(events) == 0 {
		fmt.Printf("📡 Durum          : %s ÇEVRİMDIŞI (Hiç veri gönderilmemiş)%s\n\n", Red, Reset)
		os.Exit(2)
	}

	// En son gönderilen logun saatini bulalım
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	latest := events[0]
	durationSince := time.Since(latest.Timestamp)

	// 5 dakikadan önce veri gelmişse ONLINE sayalım
	if durationSince < 5*time.Minute {
		fmt.Printf("📡 Durum          : %s● ÇEVRİMİÇİ (Aktif)%s\n", Green, Reset)
	} else {
		fmt.Printf("📡 Durum          : %s● ÇEVRİMDIŞI (Bağlantı Kopuk)%s\n", Red, Reset)
	}

	fmt.Printf("⏱️  Son Gönderim   : %s (%s önce)\n\n",
		latest.Timestamp.Local().Format("02.01.2006 15:04:05"),
		formatDuration(durationSince))

	if durationSince >= 5*time.Minute {
		os.Exit(2) // Çevrimdışıysa scriptler için hata kodu dön
	}
	os.Exit(0)
}

// --- 2. PULSECTL CONFIG VALIDATE ---
func handleConfigValidate(path string) {
	fmt.Printf("\n🔍 Config dosyası denetleniyor: %s%s%s\n", Bold, path, Reset)

	fileData, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		fatalError("Dosya okunamadı", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(fileData, &cfg); err != nil {
		fmt.Printf("%s[HATA] YAML sözdizim hatası: %s%s\n\n", Red, err.Error(), Reset)
		os.Exit(1)
	}

	// Kuralları kontrol edelim
	var errors []string

	if strings.TrimSpace(cfg.Agent.HostID) == "" {
		errors = append(errors, "agent.host_id alanı boş olamaz.")
	}
	if strings.TrimSpace(cfg.Collector.URL) == "" {
		errors = append(errors, "collector.url alanı boş olamaz.")
	}
	if cfg.Agent.IntervalSeconds <= 0 {
		errors = append(errors, "agent.interval_seconds sıfırdan büyük olmalıdır.")
	}

	// Yeni formata göre eşik değeri mantık kontrolleri (Warn < Error)
	if cfg.Checks.CPU.Enabled {
		validateThreshold(&errors, "CPU", cfg.Checks.CPU.ThresholdWarning, cfg.Checks.CPU.ThresholdError)
	}
	if cfg.Checks.RAM.Enabled {
		validateThreshold(&errors, "RAM", cfg.Checks.RAM.ThresholdWarning, cfg.Checks.RAM.ThresholdError)
	}
	if cfg.Checks.Disk.Enabled {
		validateThreshold(&errors, "DİSK", cfg.Checks.Disk.ThresholdWarning, cfg.Checks.Disk.ThresholdError)
		if strings.TrimSpace(cfg.Checks.Disk.Path) == "" {
			errors = append(errors, "checks.disk.path alanı boş olamaz.")
		}
	}

	if len(errors) > 0 {
		fmt.Printf("%s[GEÇERSİZ] Yapılandırma dosyasında %d hata bulundu:%s\n", Red, len(errors), Reset)
		for i, e := range errors {
			fmt.Printf("  %d. %s\n", i+1, e)
		}
		fmt.Println()
		os.Exit(1)
	}

	fmt.Printf("%s[BAŞARILI] Config dosyası tamamen geçerli! Hiçbir sorun yok.%s\n\n", Green, Reset)
	os.Exit(0)
}

// --- 3. PULSECTL REPORT ---
func handleReport() {
	cfg, err := loadLocalConfig("agent/config.yaml")
	if err != nil {
		fatalError("Ajan yapılandırma dosyası (agent/config.yaml) okunamadı", err)
	}

	collectorURL := strings.TrimRight(cfg.Collector.URL, "/")
	if collectorURL == "" {
		collectorURL = "http://localhost:8081"
	}

	url := fmt.Sprintf("%s/api/events?agent_id=%s", collectorURL, cfg.Agent.HostID)
	events, err := fetchEvents(url)
	if err != nil {
		fatalError("Collector ile iletişim kurulamadı", err)
	}

	if len(events) == 0 {
		fmt.Printf("\n%s[UYARI] '%s' için sistemde kayıtlı hiçbir metrik bulunamadı.%s\n\n", Yellow, cfg.Agent.HostID, Reset)
		os.Exit(0)
	}

	// En güncel CPU, RAM, DISK ölçümlerini ayıkla
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	latestMap := make(map[string]Event)
	for _, e := range events {
		if _, exists := latestMap[e.CheckType]; !exists {
			latestMap[e.CheckType] = e
		}
		if len(latestMap) == 3 {
			break
		}
	}

	// Terminalde şık bir ASCII özet tablosu çizelim
	fmt.Printf("\n%s=== '%s' HIZLI SİSTEM ÖZETİ ===%s\n", Bold, cfg.Agent.HostID, Reset)
	fmt.Println("---------------------------------------------------------")
	fmt.Printf(" %-10s | %-12s | %-12s | %-15s\n", "METRİK", "DEĞER", "DURUM", "SON ÖLÇÜM")
	fmt.Println("---------------------------------------------------------")

	printReportRow("CPU", latestMap["cpu"])
	printReportRow("RAM", latestMap["ram"])
	printReportRow("DISK", latestMap["disk"])

	fmt.Println("---------------------------------------------------------")
	fmt.Println()
	os.Exit(0)
}

// --- YARDIMCI FONKSİYONLAR ---

func printReportRow(name string, e Event) {
	if e.CheckType == "" {
		fmt.Printf(" %-10s | %-12s | %-12s | %-15s\n", name, "N/A", "VERİ YOK", "-")
		return
	}

	statusColor := Green
	if e.Level == "WARNING" {
		statusColor = Yellow
	} else if e.Level == "ERROR" {
		statusColor = Red
	}

	fmt.Printf(" %-10s | %-11.1f%% | %s%-12s%s | %-15s\n",
		name,
		e.MetricValue,
		statusColor,
		e.Level,
		Reset,
		e.Timestamp.Local().Format("15:04:05"),
	)
}

func validateThreshold(errors *[]string, name string, warn, errThresh float64) {
	if warn < 0 || warn > 100 || errThresh < 0 || errThresh > 100 {
		*errors = append(*errors, fmt.Sprintf("%s eşikleri %%0 - %%100 arasında olmalıdır.", name))
	}
	if warn >= errThresh && errThresh != 0 {
		*errors = append(*errors, fmt.Sprintf("%s uyarı eşiği (warn=%.1f), kritik eşikten (err=%.1f) küçük olmalıdır.", name, warn, errThresh))
	}
}

func loadLocalConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}

func fetchEvents(url string) ([]Event, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sunucu hatası: HTTP %d", res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	var events []Event
	err = json.Unmarshal(body, &events)
	return events, err
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second
	if m > 0 {
		return fmt.Sprintf("%ddk %dsn", m, s)
	}
	return fmt.Sprintf("%dsn", s)
}

func fatalError(msg string, err error) {
	fmt.Printf("\n%s[HATA] %s:%s %v\n\n", Red, msg, Reset, err)
	os.Exit(1)
}

func printUsageAndExit(errorMsg string, exitCode int) {
	if errorMsg != "" {
		fmt.Printf("\n%s%s%s\n", Red, errorMsg, Reset)
	}
	fmt.Printf(`
%sPulseGuard CLI (pulsectl)%s - Sistem Denetim ve Yönetim Aracı

KULLANIM:
  pulsectl <komut> [argümanlar]

KOMUTLAR:
  status                   Ajanın çalışıp çalışmadığını ve son veri gönderim zamanını gösterir.
  config validate [dosya]  Config YAML dosyasının sözdizimini ve eşik mantığını doğrular.
  report                   CPU, RAM ve Disk için terminalde hızlı bir özet tablo çizer.
  help                     Bu yardım metnini görüntüler.

ÖRNEKLER:
  pulsectl status
  pulsectl config validate agent/config.yaml
  pulsectl report
`, Bold, Reset)
	fmt.Println()
	os.Exit(exitCode)
}
