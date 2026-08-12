package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"clean-code/internal/discover"
	"clean-code/internal/hosts"
)

const version = "0.1.0-dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "version":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "version accepts no arguments")
			return 2
		}
		fmt.Fprintln(stdout, version)
		return 0
	case "hosts":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "hosts accepts no arguments")
			return 2
		}
		return writeJSON(stdout, stderr, hosts.Catalog())
	case "setup":
		flags := flag.NewFlagSet("setup", flag.ContinueOnError)
		flags.SetOutput(stderr)
		host := flags.String("host", "generic", "coding host identifier")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "setup accepts flags only")
			return 2
		}
		return writeJSON(stdout, stderr, hosts.Resolve(*host))
	case "discover":
		flags := flag.NewFlagSet("discover", flag.ContinueOnError)
		flags.SetOutput(stderr)
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		root := "."
		if flags.NArg() > 1 {
			fmt.Fprintln(stderr, "discover accepts at most one repository path")
			return 2
		}
		if flags.NArg() == 1 {
			root = flags.Arg(0)
		}
		result, err := discover.Inspect(root)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, result)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func writeJSON(stdout, stderr io.Writer, value any) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "usage: clean-code <version|hosts|setup|discover>")
}
