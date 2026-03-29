package main

import (
	"fmt"
	"os"
	"os/exec"
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

func fetch(command string) string {
	out, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		return "n/a"
	}
	return strings.TrimSpace(string(out))
}

func fetchConcurrent(commands map[string]string) map[string]string {
	var wg sync.WaitGroup
	results := make(map[string]string)
	var mu sync.Mutex

	for key, cmd := range commands {
		wg.Add(1)
		go func(k, c string) {
			defer wg.Done()
			result := fetch(c)
			mu.Lock()
			results[k] = result
			mu.Unlock()
		}(key, cmd)
	}

	wg.Wait()
	return results
}

func main() {
	rawOS := fetch("grep '^PRETTY_NAME=' /etc/os-release | cut -d'\"' -f2")
	if rawOS == "n/a" {
		rawOS = "Linux Generic"
	}

	logo := loadLogo(rawOS)
	user := os.Getenv("USER")
	host, _ := os.Hostname()

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "n/a"
	}
	term := os.Getenv("TERM")
	if term == "" {
		term = "n/a"
	}
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	if desktop == "" {
		desktop = os.Getenv("DESKTOP_SESSION")
		if desktop == "" {
			desktop = "n/a"
		}
	}

	gtkTheme := os.Getenv("GTK_THEME")
	if gtkTheme == "" {
		gtkTheme = "n/a"
	}
	iconTheme := os.Getenv("ICON_THEME")
	if iconTheme == "" {
		iconTheme = "n/a"
	}

	commands := map[string]string{
		"kernel":     "uname -sr",
		"uptime":     "uptime -p | sed 's/up //'",
		"packages":   "dpkg --get-selections 2>/dev/null | wc -l",
		"resolution": "xrandr --current 2>/dev/null | grep '*' | awk '{print $1}' | head -n 1",
		"gpu":        "lspci 2>/dev/null | grep -i 'vga\\|display' | cut -d ':' -f 3 | xargs",
		"ram":        "free -m | grep Mem | awk '{print $3\"MB / \"$2\"MB\"}'",
		"disk":       "df -h --total | grep 'total' | awk '{print $3\" / \"$2}'",
		"loadAvg":    "cat /proc/loadavg 2>/dev/null | awk '{print $1\" \"$2\" \"$3}'",
		"processes":  "ps -e --no-headers 2>/dev/null | wc -l",
		"timezone":   "date +\"%Z %z\"",
		"locale":     "locale | awk -F= '/^LANG=/{gsub(/\"/ ,\"\", $2); print $2}'",
		"fsType":     "df -T / 2>/dev/null | awk 'NR==2 {print $2}'",
		"swap":       "free -m | awk '/^Swap:/ {print $3\"MB / \"$2\"MB\"}'",
		"hostModel":  "cat /sys/devices/virtual/dmi/id/product_name 2>/dev/null || hostnamectl 2>/dev/null | grep -i 'model' | head -n1 | cut -d: -f2- | sed 's/^ //'",
		"wm":         "wmctrl -m 2>/dev/null | awk -F: 'tolower($1) ~ /name/ {gsub(/^[ \\t]+/,\"\",$2); print $2; exit}'",
	}

	cpuCommands := map[string]string{
		"cpu":   "lscpu 2>/dev/null | grep '^Model name:' | sed 's/Model name:\\s*//'",
		"cpu2":  "grep 'model name' /proc/cpuinfo 2>/dev/null | head -n1 | cut -d: -f2- | sed 's/^ //'",
		"tempC": "cat /sys/class/thermal/thermal_zone0/temp 2>/dev/null",
		"temp2": "sensors 2>/dev/null | awk '/Package id 0:|Tctl:|Tdie:|CPU Temperature:/ {print $2; exit}'",
	}

	netCommands := map[string]string{
		"localIP1":    "hostname -I 2>/dev/null | awk '{print $1}'",
		"localIP2":    "ip route get 1.1.1.1 2>/dev/null | awk '/src/ {for(i=1;i<=NF;i++) if($i==\"src\") {print $(i+1); exit}}'",
		"wifiName1":   "sh -c 'nmcli -t -f ACTIVE,SSID dev wifi 2>/dev/null | awk -F: \"$1==\\\"yes\\\" {print $2; exit}\"'",
		"wifiName2":   "iwgetid -r 2>/dev/null",
		"wifiSignal1": "sh -c 'nmcli -t -f ACTIVE,SIGNAL dev wifi 2>/dev/null | awk -F: \"$1==\\\"yes\\\" {print $2\\\"%\\\"; exit}\"'",
		"wifiSignal2": "sh -c 'iw dev 2>/dev/null | awk \"/Interface/ {print \\$2; exit}\" | xargs -r -I{} sh -c \"iw dev {} link 2>/dev/null | awk \\\"/signal/ {print \\$2\\\" dBm\\\"; exit}\\\"\"'",
	}

	batteryCommands := map[string]string{
		"battery1": "sh -c 'for b in /sys/class/power_supply/BAT*; do [ -d \"$b\" ] || continue; cap=$(cat \"$b/capacity\" 2>/dev/null); st=$(cat \"$b/status\" 2>/dev/null); if [ -n \"$cap\" ]; then echo \"${cap}% ${st}\"; exit 0; fi; done; echo n/a'",
		"battery2": "sh -c 'upower -e 2>/dev/null | grep -i BAT | head -n1 | xargs -r upower -i 2>/dev/null | awk -F: \"tolower(\\$1) ~ /percentage/ {gsub(/^[ \\t]+/, \\\"\\\", \\$2); p=\\$2} tolower(\\$1) ~ /state/ {gsub(/^[ \\t]+/, \\\"\\\", \\$2); s=\\$2} END{ if(p!=\"\") print p \" \" s; else print \"n/a\" }\"'",
	}

	results := fetchConcurrent(commands)
	cpuResults := fetchConcurrent(cpuCommands)
	netResults := fetchConcurrent(netCommands)
	batteryResults := fetchConcurrent(batteryCommands)

	kernel := results["kernel"]
	uptime := results["uptime"]
	packages := results["packages"]
	resolution := results["resolution"]
	gpu := results["gpu"]
	ram := results["ram"]
	disk := results["disk"]
	loadAvg := results["loadAvg"]
	processes := results["processes"]
	timezone := results["timezone"]
	locale := results["locale"]
	fsType := results["fsType"]
	swap := results["swap"]
	hostModel := results["hostModel"]
	if hostModel == "n/a" || hostModel == "" {
		hostModel = "Unknown"
	}
	wm := results["wm"]
	if wm == "n/a" || wm == "" {
		wm = "n/a"
	}

	cpu := cpuResults["cpu"]
	if cpu == "n/a" {
		cpu = cpuResults["cpu2"]
	}

	cpuTemp := "n/a"
	tempC := cpuResults["tempC"]
	if tempC != "n/a" && tempC != "" {
		cpuTemp = fetch("awk 'BEGIN{printf \"%.1f°C\", " + tempC + "/1000}'")
	} else {
		cpuTemp = cpuResults["temp2"]
		if cpuTemp == "n/a" || cpuTemp == "" {
			cpuTemp = "n/a"
		}
	}

	localIP := netResults["localIP1"]
	if localIP == "n/a" || localIP == "" {
		localIP = netResults["localIP2"]
	}

	wifiName := netResults["wifiName1"]
	if wifiName == "n/a" || wifiName == "" {
		wifiName = netResults["wifiName2"]
		if wifiName == "n/a" || wifiName == "" {
			wifiName = "n/a"
		}
	}

	wifiSignal := netResults["wifiSignal1"]
	if wifiSignal == "n/a" || wifiSignal == "" {
		wifiSignal = netResults["wifiSignal2"]
		if wifiSignal == "n/a" || wifiSignal == "" {
			wifiSignal = "n/a"
		}
	}

	battery := batteryResults["battery1"]
	if battery == "n/a" {
		battery = batteryResults["battery2"]
	}

	goVer := runtime.Version()

	labelWidth := 12
	info := []string{
		fmt.Sprintf("%s%s%s@%s%s", Cyan, user, White, Cyan, host),
		Grey + "────────────────────────────" + Reset,
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Host:", White, hostModel),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "OS:", White, rawOS),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Kernel:", White, kernel),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Uptime:", White, uptime),
		fmt.Sprintf("%s%-*s%s %s (dpkg)", Grey, labelWidth, "Packages:", White, packages),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Resolution:", White, resolution),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "GPU:", White, gpu),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "CPU:", White, cpu),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "CPU Temp:", White, cpuTemp),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "RAM:", White, ram),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Swap:", White, swap),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Disk:", White, disk),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Battery:", White, battery),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Load:", White, loadAvg),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Procs:", White, processes),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Local IP:", White, localIP),
		fmt.Sprintf("%s%-*s%s %s (%s)", Grey, labelWidth, "Wi-Fi:", White, wifiName, wifiSignal),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Timezone:", White, timezone),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Locale:", White, locale),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Root FS:", White, fsType),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "GTK Theme:", White, gtkTheme),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Icon Theme:", White, iconTheme),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Go:", White, goVer),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Shell:", White, shell),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Terminal:", White, term),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Desktop:", White, desktop),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "WM:", White, wm),
		fmt.Sprintf("%s%-*s%s %s", Grey, labelWidth, "Arch:", White, runtime.GOARCH),
	}

	fmt.Println()

	maxLines := len(info)
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
		if i < len(info) {
			iLine = info[i]
		}

		fmt.Printf(" %-25s %s\n", lLine, iLine)
	}
}

