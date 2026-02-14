package cli

import "flag"

// parseVersionFlag parses the --version flag and returns true if version was requested
func ParseVersionFlag() bool {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show current version")
	flag.BoolVar(&showVersion, "v", false, "Show current version")
	flag.Usage = PrintUsage
	flag.Parse()
	return showVersion
}
