package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
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

	kernel := fetch("uname -sr")
	uptime := fetch("uptime -p | sed 's/up //'")
	packages := fetch("dpkg --get-selections | wc -l")
	resolution := fetch("xrandr --current | grep '*' | awk '{print $1}' | head -n 1")
	gpu := fetch("lspci | grep -i 'vga\\|display' | cut -d ':' -f 3 | xargs")
	ram := fetch("free -m | grep Mem | awk '{print $3\"MB / \"$2\"MB\"}'")
	disk := fetch("df -h --total | grep 'total' | awk '{print $3\" / \"$2}'")

	loadAvg := fetch("cat /proc/loadavg 2>/dev/null | awk '{print $1\" \"$2\" \"$3}'")
	processes := fetch("ps -e --no-headers 2>/dev/null | wc -l")
	localIP := fetch("hostname -I 2>/dev/null | awk '{print $1}'")
	if localIP == "n/a" || localIP == "" {
		localIP = fetch("ip route get 1.1.1.1 2>/dev/null | awk '/src/ {for(i=1;i<=NF;i++) if($i==\"src\") {print $(i+1); exit}}'")
	}

	wifiName := fetch("sh -c 'nmcli -t -f ACTIVE,SSID dev wifi 2>/dev/null | awk -F: \"$1==\\\"yes\\\" {print $2; exit}\"'")
	if wifiName == "n/a" || wifiName == "" {
		wifiName = fetch("iwgetid -r 2>/dev/null")
		if wifiName == "n/a" || wifiName == "" {
			wifiName = "n/a"
		}
	}

	wifiSignal := fetch("sh -c 'nmcli -t -f ACTIVE,SIGNAL dev wifi 2>/dev/null | awk -F: \"$1==\\\"yes\\\" {print $2\\\"%\\\"; exit}\"'")
	if wifiSignal == "n/a" || wifiSignal == "" {
		wifiSignal = fetch("sh -c 'iw dev 2>/dev/null | awk \"/Interface/ {print \\$2; exit}\" | xargs -r -I{} sh -c \"iw dev {} link 2>/dev/null | awk \\\"/signal/ {print \\$2\\\" dBm\\\"; exit}\\\"\"'")
		if wifiSignal == "n/a" || wifiSignal == "" {
			wifiSignal = "n/a"
		}
	}

	timezone := fetch("date +\"%Z %z\"")
	locale := fetch("locale | awk -F= '/^LANG=/{gsub(/\"/ ,\"\", $2); print $2}'")
	if locale == "n/a" || locale == "" {
		locale = "n/a"
	}

	fsType := fetch("df -T / 2>/dev/null | awk 'NR==2 {print $2}'")
	if fsType == "n/a" || fsType == "" {
		fsType = "n/a"
	}

	gtkTheme := os.Getenv("GTK_THEME")
	if gtkTheme == "" {
		gtkTheme = "n/a"
	}
	iconTheme := os.Getenv("ICON_THEME")
	if iconTheme == "" {
		iconTheme = "n/a"
	}

	goVer := runtime.Version()

	cpu := fetch("lscpu | grep '^Model name:' | sed 's/Model name:\\s*//'")
	if cpu == "n/a" {
		cpu = fetch("grep 'model name' /proc/cpuinfo | head -n1 | cut -d: -f2- | sed 's/^ //'")
	}

	hostModel := fetch("cat /sys/devices/virtual/dmi/id/product_name 2>/dev/null || hostnamectl | grep -i 'model' | head -n1 | cut -d: -f2- | sed 's/^ //'")
	if hostModel == "n/a" || hostModel == "" {
		hostModel = "Unknown"
	}

	swap := fetch("free -m | awk '/^Swap:/ {print $3\"MB / \"$2\"MB\"}'")

	tempC := fetch("cat /sys/class/thermal/thermal_zone0/temp 2>/dev/null")
	cpuTemp := "n/a"
	if tempC != "n/a" && tempC != "" {
		cpuTemp = fetch("awk 'BEGIN{printf \"%.1f°C\", " + tempC + "/1000}'")
	} else {
		cpuTemp = fetch("sensors 2>/dev/null | awk '/Package id 0:|Tctl:|Tdie:|CPU Temperature:/ {print $2; exit}'")
	}
	
	wm := fetch("wmctrl -m 2>/dev/null | awk -F: 'tolower($1) ~ /name/ {gsub(/^[ \\t]+/,\"\",$2); print $2; exit}'")
	if wm == "n/a" || wm == "" {
		wm = "n/a"
	}

	battery := fetch("sh -c 'for b in /sys/class/power_supply/BAT*; do [ -d \"$b\" ] || continue; cap=$(cat \"$b/capacity\" 2>/dev/null); st=$(cat \"$b/status\" 2>/dev/null); if [ -n \"$cap\" ]; then echo \"${cap}% ${st}\"; exit 0; fi; done; echo n/a'")
	if battery == "n/a" {
		battery = fetch("sh -c 'upower -e 2>/dev/null | grep -i BAT | head -n1 | xargs -r upower -i 2>/dev/null | awk -F: \"tolower(\\$1) ~ /percentage/ {gsub(/^[ \\t]+/, \\\"\\\", \\$2); p=\\$2} tolower(\\$1) ~ /state/ {gsub(/^[ \\t]+/, \\\"\\\", \\$2); s=\\$2} END{ if(p!=\"\") print p \" \" s; else print \"n/a\" }\"'")
	}

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

