package main

import "groups/driver/web"

var (
	// Version : version of this executable
	Version string
	// Build : build date of this executable
	Build string
)

func main() {
	if len(Version) == 0 {
		Version = "dev"
	}

	//APIkeys
	webAdapter := web.NewWebAdapter()
	webAdapter.Start()
}
