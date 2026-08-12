package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"clean-code/internal/architecture"
	"clean-code/internal/audit"
	"clean-code/internal/benchmark"
	"clean-code/internal/contracts"
	"clean-code/internal/discover"
	"clean-code/internal/evidence"
	"clean-code/internal/hosts"
	"clean-code/internal/policy"
	"clean-code/internal/repository"
	"clean-code/internal/review"
	"clean-code/internal/runner"
	"clean-code/internal/sloppiness"
	"clean-code/internal/trace"
	"clean-code/internal/verify"
)

var version = "0.1.0-dev"

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
		output := flags.String("output", "", "optional directory for portable host instructions")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "setup accepts flags only")
			return 2
		}
		if *output != "" {
			path, err := hosts.WritePackage(*output, *host)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stderr, "host package:", path)
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
	case "verify":
		flags := flag.NewFlagSet("verify", flag.ContinueOnError)
		flags.SetOutput(stderr)
		outputDirectory := flags.String("output", "", "directory for a protected evidence bundle")
		trustedPolicy := flags.String("trusted-policy", "", "previously approved command policy")
		allowRepositoryPolicy := flags.Bool("allow-repository-policy", false, "explicitly approve this repository's command policy")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() > 1 {
			fmt.Fprintln(stderr, "verify accepts at most one repository path")
			return 2
		}
		if *trustedPolicy != "" && *allowRepositoryPolicy {
			fmt.Fprintln(stderr, "verify accepts either --trusted-policy or --allow-repository-policy, not both")
			return 2
		}
		root := "."
		if flags.NArg() == 1 {
			root = flags.Arg(0)
		}
		discovery, err := discover.Inspect(root)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		var trustedCommands []contracts.CommandSpec
		trustedSource := ""
		if *trustedPolicy != "" {
			trustedPath, err := filepath.Abs(*trustedPolicy)
			if err != nil {
				fmt.Fprintln(stderr, "resolve trusted policy:", err)
				return 1
			}
			if _, err := os.Lstat(trustedPath); err != nil {
				fmt.Fprintln(stderr, "inspect trusted policy:", err)
				return 1
			}
			trustedCommands, err = discover.LoadCommands(trustedPath)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			trustedSource = trustedPath
		} else if !*allowRepositoryPolicy && len(discovery.Commands) > 0 {
			trustedSource = "unapproved repository policy"
		}
		revision := repository.Revision(discovery.Root)
		report := verify.Service{
			Runner:          runner.Runner{Root: discovery.Root},
			CurrentRevision: repository.Revision,
		}.Run(
			context.Background(), discovery.Root, revision,
			discovery.Commands, trustedCommands, trustedSource,
		)
		if *outputDirectory != "" {
			path, err := evidence.WriteBundle(*outputDirectory, report)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stderr, "evidence:", path)
		}
		if code := writeJSON(stdout, stderr, report); code != 0 {
			return code
		}
		if !report.Successful() {
			return 1
		}
		return 0
	case "architecture":
		flags := flag.NewFlagSet("architecture", flag.ContinueOnError)
		flags.SetOutput(stderr)
		policyPath := flags.String("policy", "", "architecture policy JSON file")
		graphPath := flags.String("graph", "", "dependency graph JSON file")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *policyPath == "" || *graphPath == "" {
			fmt.Fprintln(stderr, "architecture requires --policy and --graph")
			return 2
		}
		policy, err := architecture.LoadPolicy(*policyPath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		graph, err := architecture.LoadGraph(*graphPath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		report := architecture.Evaluate(policy, graph)
		if code := writeJSON(stdout, stderr, report); code != 0 {
			return code
		}
		if report.Status != "PASS" {
			return 1
		}
		return 0
	case "trace":
		flags := flag.NewFlagSet("trace", flag.ContinueOnError)
		flags.SetOutput(stderr)
		planPath := flags.String("plan", "", "test plan JSON file")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *planPath == "" {
			fmt.Fprintln(stderr, "trace requires --plan")
			return 2
		}
		plan, err := trace.Load(*planPath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		report := trace.Evaluate(plan)
		if code := writeJSON(stdout, stderr, report); code != 0 {
			return code
		}
		if report.Status != "PASS" {
			return 1
		}
		return 0
	case "review":
		flags := flag.NewFlagSet("review", flag.ContinueOnError)
		flags.SetOutput(stderr)
		inputPath := flags.String("input", "", "review input JSON file")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *inputPath == "" {
			fmt.Fprintln(stderr, "review requires --input")
			return 2
		}
		input, err := review.Load(*inputPath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		report := review.Evaluate(input)
		if code := writeJSON(stdout, stderr, report); code != 0 {
			return code
		}
		if report.Status != "PASS" {
			return 1
		}
		return 0
	case "audit":
		flags := flag.NewFlagSet("audit", flag.ContinueOnError)
		flags.SetOutput(stderr)
		inputPath := flags.String("input", "", "audit input JSON file")
		outputPath := flags.String("output", "", "immutable audit receipt JSON file")
		checkPath := flags.String("check", "", "existing receipt to verify against current evidence")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *inputPath == "" || (*outputPath == "") == (*checkPath == "") {
			fmt.Fprintln(stderr, "audit requires --input and exactly one of --output or --check")
			return 2
		}
		if *checkPath != "" {
			receipt, err := audit.Check(*inputPath, *checkPath)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			return writeJSON(stdout, stderr, receipt)
		}
		receipt, err := audit.Build(*inputPath, nil)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if err := audit.Write(*outputPath, receipt); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if code := writeJSON(stdout, stderr, receipt); code != 0 {
			return code
		}
		if !receipt.Complete {
			return 1
		}
		return 0
	case "slop":
		flags := flag.NewFlagSet("slop", flag.ContinueOnError)
		flags.SetOutput(stderr)
		previousPath := flags.String("previous", "", "first-pass sloppiness report; supplying it makes this the final pass")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() > 1 {
			fmt.Fprintln(stderr, "slop accepts at most one repository path")
			return 2
		}
		root := "."
		if flags.NArg() == 1 {
			root = flags.Arg(0)
		}
		var previous *sloppiness.Report
		if *previousPath != "" {
			loaded, err := sloppiness.LoadFile(*previousPath)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			previous = &loaded
		}
		report, err := sloppiness.Assess(root, previous)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if code := writeJSON(stdout, stderr, report); code != 0 {
			return code
		}
		if report.Cycle.Status != "DONE" {
			return 1
		}
		return 0
	case "benchmark":
		flags := flag.NewFlagSet("benchmark", flag.ContinueOnError)
		flags.SetOutput(stderr)
		manifestPath := flags.String("manifest", "", "benchmark manifest JSON or JSON-compatible YAML")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *manifestPath == "" {
			fmt.Fprintln(stderr, "benchmark requires --manifest")
			return 2
		}
		manifest, err := benchmark.Load(*manifestPath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, benchmark.Score(manifest))
	case "learn":
		flags := flag.NewFlagSet("learn", flag.ContinueOnError)
		flags.SetOutput(stderr)
		proposalPath := flags.String("proposal", "", "policy change proposal JSON file")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *proposalPath == "" {
			fmt.Fprintln(stderr, "learn requires --proposal")
			return 2
		}
		proposal, err := policy.LoadChangeProposal(*proposalPath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		report := policy.EvaluateChangeProposal(proposal)
		if code := writeJSON(stdout, stderr, report); code != 0 {
			return code
		}
		if report.Status != "PASS" {
			return 1
		}
		return 0
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
	fmt.Fprintln(output, "usage: clean-code <version|hosts|setup|discover|verify|architecture|trace|review|audit|slop|benchmark|learn>")
}
