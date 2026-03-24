package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	Cyan  = "\033[38;5;51m"
	White = "\033[38;5;255m"
	Gold  = "\033[38;5;220m"
	Grey  = "\033[38;5;244m"
	Reset = "\033[0m"
)

type SystemInfo struct {
	User       string
	Host       string
	HostModel  string
	RawOS      string
	Shell      string
	Term       string
	Desktop    string
	Kernel     string
	Uptime     string
	Packages   string
	Resolution string
	GPU        string
	RAM        string
	Disk       string
	LoadAvg    string
	Processes  string
	LocalIP    string
	WiFiName   string
	WiFiSignal string
	Timezone   string
	Locale     string
	FSType     string
	GTKTheme   string
	IconTheme  string
	GoVer      string
	CPU        string
	CPUTemp    string
	Swap       string
	Battery    string
	WM         string
	Arch       string
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "n/a"
	}
	return strings.TrimSpace(string(data))
}

func execCmd(cmd string) string {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "n/a"
	}
	return strings.TrimSpace(string(out))
}

func readProc(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return "n/a"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return "n/a"
}

func getCPUTemp() string {
	tempC := readFile("/sys/class/thermal/thermal_zone0/temp")
	if tempC != "n/a" && tempC != "" {
		return execCmd(fmt.Sprintf("awk 'BEGIN{printf \"%.1f°C\", %s/1000}'", tempC))
	}
	return execCmd("sensors 2>/dev/null | awk '/Package id 0:|Tctl:|Tdie:|CPU Temperature:/ {print $2; exit}'")
}

func getWiFiInfo() (string, string) {
	name := execCmd("nmcli -t -f ACTIVE,SSID dev wifi 2>/dev/null | awk -F: '$1==\"yes\" {print $2; exit}'")
	if name == "n/a" || name == "" {
		name = execCmd("iwgetid -r 2>/dev/null")
	}
	if name == "n/a" || name == "" {
		name = "n/a"
	}

	signal := execCmd("nmcli -t -f ACTIVE,SIGNAL dev wifi 2>/dev/null | awk -F: '$1==\"yes\" {print $2\"%\"; exit}'")
	if signal == "n/a" || signal == "" {
		signal = "n/a"
	}

	return name, signal
}

func getBattery() string {
	battery := readFile("/sys/class/power_supply/BAT0/capacity")
	if battery == "n/a" {
		battery = execCmd("cat /sys/class/power_supply/BAT*/capacity 2>/dev/null | head -n1")
	}
	if battery == "n/a" {
		battery = execCmd("upower -e 2>/dev/null | grep -i BAT | head -n1 | xargs -r upower -i 2>/dev/null | grep percentage | awk '{print $2}'")
	}
	return battery + "%"
}

func getCPU() string {
	cpu := execCmd("grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ //'"
	if cpu == "n/a" {
		cpu = execCmd("lscpu | grep 'Model name' | cut -d: -f2 | sed 's/^ //'" )
	}
	return cpu
}

func getLocalIP() string {
	ip := execCmd("hostname -I 2>/dev/null | awk '{print $1}'")
	if ip == "n/a" || ip == "" {
		ip = execCmd("ip route get 1.1.1.1 2>/dev/null | awk '/src/ {print $NF}'")
	}
	return ip
}

func getGPU() string {
	return execCmd("lspci 2>/dev/null | grep -i 'vga\\|3d\\|display' | head -n1 | cut -d: -f3 | xargs")
}

func collectSystemInfo() SystemInfo {
	info := SystemInfo{
		User:   os.Getenv("USER"),
		Shell:  os.Getenv("SHELL"),
		Term:   os.Getenv("TERM"),
		GoVer:  runtime.Version(),
		Arch:   runtime.GOARCH,
	}

	if info.Shell == "" {
		info.Shell = "n/a"
	}
	if info.Term == "" {
		info.Term = "n/a"
	}

	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	if desktop == "" {
		desktop = os.Getenv("DESKTOP_SESSION")
	}
	if desktop == "" {
		desktop = "n/a"
	}
	info.Desktop = desktop
	info.GTKTheme = os.Getenv("GTK_THEME")
	if info.GTKTheme == "" {
		info.GTKTheme = "n/a"
	}
	info.IconTheme = os.Getenv("ICON_THEME")
	if info.IconTheme == "" {
		info.IconTheme = "n/a"
	}

	var wg sync.WaitGroup
	mu := &sync.Mutex{}

	fetchData := func(key string, fn func() string) {
		defer wg.Done()
		result := fn()
		mu.Lock()
		switch key {
		case "host":
			info.Host, _ = os.Hostname()
		case "rawOS":
			info.RawOS = result
		case "kernel":
			info.Kernel = result
		case "uptime":
			info.Uptime = result
		case "packages":
			info.Packages = result
		case "resolution":
			info.Resolution = result
		case "gpu":
			info.GPU = result
		case "ram":
			info.RAM = result
		case "disk":
			info.Disk = result
		case "loadAvg":
			info.LoadAvg = result
		case "processes":
			info.Processes = result
		case "localIP":
			info.LocalIP = result
		case "timezone":
			info.Timezone = result
		case "locale":
			info.Locale = result
		case "fsType":
			info.FSType = result
		case "cpu":
			info.CPU = result
		case "cpuTemp":
			info.CPUTemp = result
		case "swap":
			info.Swap = result
		case "wm":
			info.WM = result
		case "hostModel":
			info.HostModel = result
		case "battery":
			info.Battery = result
		}
		mu.Unlock()
	}

	wg.Add(21)

	go fetchData("rawOS", func() string {
		return execCmd("grep '^PRETTY_NAME=' /etc/os-release 2>/dev/null | cut -d'\"' -f2")
	})

	go func() {
		defer wg.Done()
		info.Host, _ = os.Hostname()
	}()

	go fetchData("kernel", func() string {
		return execCmd("uname -sr")
	})

	go fetchData("uptime", func() string {
		return execCmd("uptime -p 2>/dev/null | sed 's/up //' || echo 'n/a'")
	})

	go fetchData("packages", func() string {
		return execCmd("dpkg --get-selections 2>/dev/null | wc -l || echo n/a")
	})

	go fetchData("resolution", func() string {
		return execCmd("xrandr --current 2>/dev/null | grep '*' | awk '{print $1}' | head -n1 || echo n/a")
	})

	go fetchData("gpu", getGPU)

	go fetchData("ram", func() string {
		return execCmd("free -h 2>/dev/null | awk '/^Mem:/ {print $3 \" / \" $2}' || echo n/a")
	})

	go fetchData("disk", func() string {
		return execCmd("df -h --total 2>/dev/null | grep total | awk '{print $3 \" / \" $2}' || echo n/a")
	})

	go fetchData("loadAvg", func() string {
		return readProc("/proc/loadavg")
	})

	go fetchData("processes", func() string {
		return execCmd("ps -e --no-headers 2>/dev/null | wc -l || echo n/a")
	})

	go fetchData("localIP", getLocalIP)

	go fetchData("timezone", func() string {
		return execCmd("date +\"%Z %z\" 2>/dev/null || echo n/a")
	})

	go fetchData("locale", func() string {
		return execCmd("locale 2>/dev/null | awk -F= '/^LANG=/{gsub(/\"/, \"\", $2); print $2}' || echo n/a")
	})

	go fetchData("fsType", func() string {
		return execCmd("df -T / 2>/dev/null | awk 'NR==2 {print $2}' || echo n/a")
	})

	go fetchData("cpu", getCPU)

	go fetchData("cpuTemp", getCPUTemp)

	go fetchData("swap", func() string {
		return execCmd("free -h 2>/dev/null | awk '/^Swap:/ {print $3 \" / \" $2}' || echo n/a")
	})

	go fetchData("wm", func() string {
		return execCmd("wmctrl -m 2>/dev/null | awk -F: 'tolower($1) ~ /name/ {gsub(/^[ \t]+/,\"\",$2); print $2; exit}' || echo n/a")
	})

	go fetchData("hostModel", func() string {
		model := readFile("/sys/devices/virtual/dmi/id/product_name")
		if model == "n/a" {
			model = execCmd("hostnamectl 2>/dev/null | grep -i chassis | head -n1 | cut -d: -f2 | sed 's/^ //' ")
		}
		if model == "n/a" || model == "" {
			model = "Unknown"
		}
		return model
	})

	go fetchData("battery", getBattery)

	wg.Wait()

	wifiName, wifiSignal := getWiFiInfo()
	info.WiFiName = wifiName
	info.WiFiSignal = wifiSignal

	if info.RawOS == "n/a" || info.RawOS == "" {
		info.RawOS = "Linux Generic"
	}
	if info.Locale == "n/a" || info.Locale == "" {
		info.Locale = "n/a"
	}
	if info.FSType == "n/a" || info.FSType == "" {
		info.FSType = "n/a"
	}
	if info.LoadAvg != "n/a" {
		parts := strings.Fields(info.LoadAvg)
		if len(parts) >= 3 {
			info.LoadAvg = parts[0] + " " + parts[1] + " " + parts[2]
		}
	}

	return info
}

func main() {
	logo := loadLogo("")
	info := collectSystemInfo()

	labelWidth := 12
	infoLines := []string{
		fmt.Sprintf("%s%s%s@%s%s", Cyan, info.User, White, Cyan, info.Host),
		Grey + "────────────────────────────" + Reset,
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Host:", White, info.HostModel),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "OS:", White, info.RawOS),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Kernel:", White, info.Kernel),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Uptime:", White, info.Uptime),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Packages:", White, info.Packages),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Resolution:", White, info.Resolution),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "GPU:", White, info.GPU),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "CPU:", White, info.CPU),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "CPU Temp:", White, info.CPUTemp),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "RAM:", White, info.RAM),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Swap:", White, info.Swap),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Disk:", White, info.Disk),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Battery:", White, info.Battery),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Load:", White, info.LoadAvg),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Procs:", White, info.Processes),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Local IP:", White, info.LocalIP),
		fmt.Sprintf("%s%-*s%s %s (%s)", Grey, labelWidth, "Wi-Fi:", White, info.WiFiName, info.WiFiSignal),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Timezone:", White, info.Timezone),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Locale:", White, info.Locale),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Root FS:", White, info.FSType),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "GTK Theme:", White, info.GTKTheme),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Icon Theme:", White, info.IconTheme),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Go:", White, info.GoVer),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Shell:", White, info.Shell),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Terminal:", White, info.Term),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Desktop:", White, info.Desktop),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "WM:", White, info.WM),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Arch:", White, info.Arch),
	}

	fmt.Println()

	maxLines := len(infoLines)
	if len(logo) > maxLines {
		maxLines = len(logo)
	}

	logoOffset := 0
	if maxLines > len(logo) {
		logoOffset = (maxLines - len(logo)) / 2
	}

	for i := 0; i < maxLines; i++ {
		lLine := " "
		logoIndex := i - logoOffset
		if logoIndex >= 0 && logoIndex < len(logo) {
			lLine = logo[logoIndex]
		}

		iLine := ""
		if i < len(infoLines) {
			iLine = infoLines[i]
		}

		fmt.Printf(" %-25s %s\n", lLine, iLine)
	}

	fmt.Printf("\n %s © Gem 2026 - MacBook Pro 2012 Original %s\n\n", Gold, Reset)
}