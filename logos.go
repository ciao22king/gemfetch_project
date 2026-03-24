package main

// Default ASCII logo: Tux (more recognizable)
var tuxLogo = []string{
	"        .--.            ",
	"       |o_o |           ",
	"       |:_/ |           ",
	"      //   \\ \\          ",
	"     (|     | )         ",
	"    /'\\_   _/`\\         ",
	"    \\___)=(___/         ",
	"          Tux           ",
}

// For now we always show Tux for any distro.
func loadLogo(_ string) []string {
	return tuxLogo
}
