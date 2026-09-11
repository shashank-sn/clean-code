package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/shashank-sn/clean-code/internal/agents"
	"github.com/shashank-sn/clean-code/internal/architecture"
	"github.com/shashank-sn/clean-code/internal/audit"
	"github.com/shashank-sn/clean-code/internal/benchmark"
	"github.com/shashank-sn/clean-code/internal/contracts"
	"github.com/shashank-sn/clean-code/internal/discover"
	"github.com/shashank-sn/clean-code/internal/evidence"
	"github.com/shashank-sn/clean-code/internal/gauntlet"
	"github.com/shashank-sn/clean-code/internal/hosts"
	"github.com/shashank-sn/clean-code/internal/policy"
	"github.com/shashank-sn/clean-code/internal/providers"
	"github.com/shashank-sn/clean-code/internal/repository"
	"github.com/shashank-sn/clean-code/internal/review"
	"github.com/shashank-sn/clean-code/internal/runner"
	"github.com/shashank-sn/clean-code/internal/trace"
	"github.com/shashank-sn/clean-code/internal/verify"
)

var version = "0.1.0-dev" // overridden via -ldflags in release and npm builds

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "agent":
		return runAgent(args[1:], stdout, stderr)
	case "provider":
		return runProvider(args[1:], stdout, stderr)
	case "gauntlet":
		return runGauntlet(args[1:], stdout, stderr)
	case "version":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "version accepts no arguments")
			return 2
		}
		displayVersion := resolveVersion()
		fmt.Fprintln(stdout, displayVersion)
		if displayVersion == "0.1.0-dev" {
			exe, err := os.Executable()
			if err == nil {
				fmt.Fprintf(stderr, "clean-code: Go binary at %s (not the npm wrapper).\n", exe)
			}
			fmt.Fprintln(stderr, "clean-code: run `which -a clean-code` — remove ~/.local/bin/clean-code if you use npm.")
			fmt.Fprintln(stderr, "clean-code: npm install -g @shashanksn/clean-code@latest && hash -r")
		}
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
		if *allowRepositoryPolicy && len(discovery.Commands) > 0 {
			fmt.Fprintf(stderr,
				"clean-code: WARNING: --allow-repository-policy is set; executing %d UNapproved command(s) declared by %s\n"+
					"clean-code: WARNING: these repository-declared commands are NOT approved and running them can execute arbitrary code.\n"+
					"clean-code: WARNING: only run verify on repositories you trust.\n",
				len(discovery.Commands), filepath.Join(discovery.Root, ".clean-code.json"))
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
		if len(args) > 1 && args[1] == "view" {
			return runArchitectureView(args[2:], stdout, stderr)
		}
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
		signingKey := flags.String("signing-key", "", "hex-encoded ed25519 private key seed for signing the receipt")
		signingKeyFile := flags.String("signing-key-file", "", "file containing the hex-encoded ed25519 private key seed")
		witnessFile := flags.String("witness-file", "", "external witness file binding the receipt hash (append on output, verify on check)")
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
			if receipt.Signature != "" {
				if err := audit.VerifyReceipt(receipt); err != nil {
					fmt.Fprintln(stderr, err)
					return 1
				}
				fmt.Fprintln(stderr, "audit receipt signature verified")
			}
			if *witnessFile != "" {
				if err := verifyWitness(*witnessFile, *checkPath); err != nil {
					fmt.Fprintln(stderr, err)
					return 1
				}
				fmt.Fprintln(stderr, "audit receipt witness verified")
			}
			return writeJSON(stdout, stderr, receipt)
		}
		seed, err := resolveSigningSeed(*signingKey, *signingKeyFile)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		receipt, err := audit.Build(*inputPath, nil)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(seed) > 0 {
			receipt, err = audit.SignReceipt(receipt, seed)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
		}
		if err := audit.Write(*outputPath, receipt); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if *witnessFile != "" {
			if err := writeWitness(*witnessFile, *outputPath); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
		}
		if code := writeJSON(stdout, stderr, receipt); code != 0 {
			return code
		}
		if !receipt.Complete {
			return 1
		}
		return 0
	case "keygen":
		flags := flag.NewFlagSet("keygen", flag.ContinueOnError)
		flags.SetOutput(stderr)
		output := flags.String("output", "", "optional directory to write the private seed and public key files")
		if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 {
			fmt.Fprintln(stderr, "keygen accepts --output only")
			return 2
		}
		seed, err := generateSigningSeed()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		privateKey := ed25519.NewKeyFromSeed(seed)
		publicKey := privateKey.Public().(ed25519.PublicKey)
		fingerprint := sha256.Sum256(publicKey)
		fmt.Fprintln(stdout, "ed25519 signing key generated")
		fmt.Fprintln(stdout, "private seed (hex, 32 bytes):", hex.EncodeToString(seed))
		fmt.Fprintln(stdout, "public key (base64):", base64.StdEncoding.EncodeToString(publicKey))
		fmt.Fprintln(stdout, "key fingerprint (hex):", hex.EncodeToString(fingerprint[:]))
		fmt.Fprintln(stdout, "keep the private seed secret; distribute the public key and fingerprint to verifiers.")
		if *output != "" {
			if err := os.MkdirAll(*output, 0o700); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			seedPath := filepath.Join(*output, "signing-key.seed")
			if err := os.WriteFile(seedPath, []byte(hex.EncodeToString(seed)+"\n"), 0o600); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			pubPath := filepath.Join(*output, "signing-key.pub")
			if err := os.WriteFile(pubPath, []byte(base64.StdEncoding.EncodeToString(publicKey)+"\n"), 0o644); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stderr, "wrote", seedPath, "and", pubPath)
		}
		return 0
	case "sign-role":
		flags := flag.NewFlagSet("sign-role", flag.ContinueOnError)
		flags.SetOutput(stderr)
		role := flags.String("role", "", "role to attest: implementer, reviewer, or spot_checker (the auditor attests by signing the receipt)")
		revision := flags.String("revision", "", "audit revision being attested")
		evidence := flags.String("evidence", "", "path to the role's evidence file")
		signingKey := flags.String("signing-key", "", "hex-encoded ed25519 private key seed")
		signingKeyFile := flags.String("signing-key-file", "", "file containing the hex-encoded ed25519 private key seed")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *role == "" || *revision == "" || *evidence == "" {
			fmt.Fprintln(stderr, "sign-role requires --role, --revision, and --evidence")
			return 2
		}
		switch *role {
		case "implementer", "reviewer", "spot_checker":
		default:
			fmt.Fprintln(stderr, "sign-role role must be implementer, reviewer, or spot_checker")
			return 2
		}
		seed, err := resolveSigningSeed(*signingKey, *signingKeyFile)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(seed) == 0 {
			fmt.Fprintln(stderr, "sign-role requires a signing key (--signing-key, --signing-key-file, or CLEAN_CODE_SIGNING_KEY)")
			return 2
		}
		signer, err := audit.SignRole(*role, *revision, *evidence, seed)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stderr, "role attestation signed; add it to the audit input under signers."+*role+" (key_id:", signer.KeyID+")")
		return writeJSON(stdout, stderr, signer)
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
	case "compare-workflows":
		flags := flag.NewFlagSet("compare-workflows", flag.ContinueOnError)
		flags.SetOutput(stderr)
		manifestPath := flags.String("manifest", "harness/calibration/workflow-comparison.json", "workflow comparison manifest")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "compare-workflows accepts flags only")
			return 2
		}
		manifest, err := benchmark.LoadWorkflowManifest(*manifestPath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, benchmark.CompareWorkflows(manifest))
	case "benchmark-full-flow":
		flags := flag.NewFlagSet("benchmark-full-flow", flag.ContinueOnError)
		flags.SetOutput(stderr)
		manifestPath := flags.String("manifest", "harness/calibration/full-flow-manifest.json", "full-flow benchmark manifest")
		repoRoot := flags.String("repo", ".", "repository root")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "benchmark-full-flow accepts flags only")
			return 2
		}
		manifest, err := benchmark.LoadFullFlowManifest(*manifestPath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		report, err := benchmark.RunFullFlow(*repoRoot, manifest)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, report)
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
	fmt.Fprintln(output, "usage: clean-code <agent|provider|gauntlet|version|hosts|setup|discover|verify|architecture|trace|review|audit|keygen|sign-role|benchmark|compare-workflows|benchmark-full-flow|learn>")
}

func runProvider(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "validate" && args[0] != "result") {
		fmt.Fprintln(stderr, "provider requires validate or result")
		return 2
	}
	flags := flag.NewFlagSet("provider "+args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("input", "", "provider manifest or result JSON file")
	var manifest *string
	if args[0] == "validate" {
		path = flags.String("manifest", "", "provider manifest JSON file")
	} else {
		manifest = flags.String("manifest", "", "provider manifest JSON file for result binding")
	}
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || *path == "" {
		fmt.Fprintln(stderr, "provider "+args[0]+" requires its input flag")
		return 2
	}
	if args[0] == "validate" {
		spec, err := providers.Load(*path)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, spec)
	}
	body, err := os.ReadFile(*path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	result, err := providers.ParseResult(body)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if manifest != nil && *manifest != "" {
		spec, err := providers.Load(*manifest)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if err := providers.ValidateResult(spec, result); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	return writeJSON(stdout, stderr, result)
}

func runAgent(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "agent requires list, describe, validate, or emit")
		return 2
	}
	switch args[0] {
	case "list":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "agent list accepts no arguments")
			return 2
		}
		values, err := agents.List()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, values)
	case "describe":
		flags := flag.NewFlagSet("agent describe", flag.ContinueOnError)
		flags.SetOutput(stderr)
		host := flags.String("host", "generic", "runtime host identifier")
		if len(args) < 2 || flags.Parse(args[2:]) != nil || flags.NArg() != 0 {
			fmt.Fprintln(stderr, "agent describe requires an id")
			return 2
		}
		value, err := agents.Describe(args[1], *host)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, value)
	case "validate":
		flags := flag.NewFlagSet("agent validate", flag.ContinueOnError)
		flags.SetOutput(stderr)
		if err := flags.Parse(args[1:]); err != nil || flags.NArg() > 1 {
			fmt.Fprintln(stderr, "agent validate accepts zero or one id")
			return 2
		}
		id := ""
		if flags.NArg() == 1 {
			id = flags.Arg(0)
		}
		if err := agents.Validate(id); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, map[string]string{"schema_version": agents.SchemaVersion, "status": "PASS", "agent": id})
	case "emit":
		flags := flag.NewFlagSet("agent emit", flag.ContinueOnError)
		flags.SetOutput(stderr)
		host := flags.String("host", "generic", "runtime host identifier")
		mode := flags.String("mode", "prompt", "emit mode: prompt or json")
		output := flags.String("output", "", "new output file")
		if len(args) < 2 || flags.Parse(args[2:]) != nil || flags.NArg() != 0 || (*mode != "prompt" && *mode != "json") {
			fmt.Fprintln(stderr, "agent emit requires an id and prompt or json mode")
			return 2
		}
		var body []byte
		var err error
		if *mode == "prompt" {
			var prompt string
			prompt, err = agents.EmitPrompt(args[1], *host)
			body = []byte(prompt)
		} else {
			body, err = agents.EmitJSON(args[1], *host)
			body = append(body, '\n')
		}
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if *output != "" {
			if err := writeNewFile(*output, body); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stderr, "agent artifact:", *output)
			return 0
		}
		_, err = stdout.Write(body)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	default:
		fmt.Fprintln(stderr, "agent requires list, describe, validate, or emit")
		return 2
	}
}

func runGauntlet(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "gauntlet requires plan or run")
		return 2
	}
	flags := flag.NewFlagSet("gauntlet "+args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "", "gauntlet manifest JSON file")
	output := flags.String("output", "", "output directory")
	repoRoot := flags.String("repo", ".", "repository root for artifact validation")
	if (args[0] != "plan" && args[0] != "run") || flags.Parse(args[1:]) != nil || flags.NArg() != 0 || *manifestPath == "" || *output == "" {
		fmt.Fprintln(stderr, "gauntlet plan|run requires --manifest and --output")
		return 2
	}
	manifest, err := gauntlet.Load(*manifestPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if args[0] == "plan" {
		packets := gauntlet.Packets(manifest)
		if err := gauntlet.WritePackets(*output, packets); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, map[string]any{"schema_version": "1.0.0", "revision": manifest.Revision, "packets": len(packets), "status": "PLANNED"})
	}
	report := gauntlet.ValidateArtifacts(manifest, *repoRoot)
	if err := os.MkdirAll(*output, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := writeNewFile(filepath.Join(*output, "gauntlet-report.json"), append(body, '\n')); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if code := writeJSON(stdout, stderr, report); code != 0 {
		return code
	}
	passes, failures, notRuns := summarizeGauntlet(report)
	if failures > 0 {
		fmt.Fprintf(stderr, "gauntlet run: %d stage(s) FAILED artifact validation\n", failures)
		return 1
	}
	if notRuns > 0 {
		fmt.Fprintf(stderr, "gauntlet run: portable core validates artifacts only; %d stage(s) not validated — full execution requires a host adapter (see docs/gauntlet.md)\n", notRuns)
		return 2
	}
	fmt.Fprintf(stderr, "gauntlet run: all %d stage(s) PASS artifact validation\n", passes)
	return 0
}

func summarizeGauntlet(report gauntlet.Report) (passes, failures, notRuns int) {
	for _, story := range report.Stories {
		for _, stage := range story.Stages {
			switch stage.Status {
			case contracts.StatusPass:
				passes++
			case contracts.StatusFail:
				failures++
			default:
				notRuns++
			}
		}
	}
	return passes, failures, notRuns
}

func runArchitectureView(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("architecture view", flag.ContinueOnError)
	flags.SetOutput(stderr)
	policyPath := flags.String("policy", "", "architecture policy JSON file")
	graphPath := flags.String("graph", "", "dependency graph JSON file")
	previousPath := flags.String("previous", "", "optional prior dependency graph JSON file")
	producer := flags.String("producer", "", "graph producer id")
	scope := flags.String("collection-scope", "", "producer collection scope")
	evidence := flags.String("evidence-sha256", "", "coverage evidence SHA-256")
	evidenceFile := flags.String("evidence-file", "", "coverage evidence JSON file")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || *policyPath == "" || *graphPath == "" {
		fmt.Fprintln(stderr, "architecture view requires --policy and --graph")
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
	var previous *architecture.Graph
	if *previousPath != "" {
		loaded, loadErr := architecture.LoadGraph(*previousPath)
		if loadErr != nil {
			fmt.Fprintln(stderr, loadErr)
			return 1
		}
		previous = &loaded
	}
	var proof *architecture.CoverageProof
	if *producer != "" || *scope != "" || *evidence != "" || *evidenceFile != "" {
		proof = &architecture.CoverageProof{ProducerID: *producer, CollectionScope: *scope, EvidenceSHA256: *evidence, EvidencePath: *evidenceFile}
	}
	view, err := architecture.BuildViewWithCoverage(policy, graph, previous, proof)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeJSON(stdout, stderr, view)
}

func writeNewFile(path string, body []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return errors.New("output file already exists")
		}
		return err
	}
	defer file.Close()
	_, err = file.Write(body)
	return err
}

// resolveSigningSeed returns the ed25519 private-key seed from --signing-key,
// then --signing-key-file, then the CLEAN_CODE_SIGNING_KEY environment variable.
// It returns nil when no source is configured.
func resolveSigningSeed(signingKey, signingKeyFile string) ([]byte, error) {
	var encoded string
	switch {
	case signingKey != "":
		encoded = signingKey
	case signingKeyFile != "":
		body, err := os.ReadFile(signingKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read signing key file: %w", err)
		}
		encoded = strings.TrimSpace(string(body))
	default:
		encoded = strings.TrimSpace(os.Getenv("CLEAN_CODE_SIGNING_KEY"))
	}
	if encoded == "" {
		return nil, nil
	}
	seed, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("signing key must be a hex-encoded ed25519 seed")
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("signing key seed must be %d bytes, got %d", ed25519.SeedSize, len(seed))
	}
	return seed, nil
}

// generateSigningSeed returns a fresh random ed25519 private-key seed.
func generateSigningSeed() ([]byte, error) {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		return nil, fmt.Errorf("generate signing seed: %w", err)
	}
	return seed, nil
}

// sha256File returns the hex SHA-256 of a file's bytes.
func sha256File(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", path, err)
	}
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:]), nil
}

// writeWitness appends a "<sha256>  <path>" record for the receipt to the
// witness file. The witness file is append-only by design so records accumulate
// and regenerating a receipt leaves the old record visible as a mismatch.
func writeWitness(witnessFile, receiptPath string) error {
	digest, err := sha256File(receiptPath)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(witnessFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open witness file: %w", err)
	}
	defer file.Close()
	if _, err := fmt.Fprintf(file, "%s  %s\n", digest, receiptPath); err != nil {
		return fmt.Errorf("append to witness file: %w", err)
	}
	return nil
}

// verifyWitness confirms the on-disk receipt hash matches the most recent
// witness record for that receipt path. A mismatch means the receipt was
// regenerated or edited after it was witnessed. Records are
// "<64-hex-digest>  <path>"; the path may itself contain spaces, so only the
// fixed two-space separator after the digest is used to split.
func verifyWitness(witnessFile, receiptPath string) error {
	digest, err := sha256File(receiptPath)
	if err != nil {
		return err
	}
	body, err := os.ReadFile(witnessFile)
	if err != nil {
		return fmt.Errorf("read witness file: %w", err)
	}
	var found string
	for _, line := range strings.Split(string(body), "\n") {
		if len(line) < 66 || line[64:66] != "  " {
			continue
		}
		if strings.TrimSpace(line[66:]) == receiptPath {
			found = line[:64]
		}
	}
	if found == "" {
		return fmt.Errorf("audit witness: no record for %q in %s", receiptPath, witnessFile)
	}
	if found != digest {
		return fmt.Errorf("audit witness: %q was regenerated or edited after witnessing (recorded %s, current %s)", receiptPath, found, digest)
	}
	return nil
}
