package main

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/shashank-sn/clean-code/internal/adaptive"
)

func runRoute(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("route", flag.ContinueOnError)
	flags.SetOutput(stderr)
	input := flags.String("input", "", "task signals JSON file")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *input == "" {
		fmt.Fprintln(stderr, "route requires --input")
		return 2
	}
	signals, err := adaptive.LoadTaskSignals(*input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	decision, err := adaptive.Route(signals)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeJSON(stdout, stderr, decision)
}

func runArena(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "validate" {
		fmt.Fprintln(stderr, "arena requires validate")
		return 2
	}
	flags := flag.NewFlagSet("arena validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	input := flags.String("input", "", "arena decision record JSON file")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *input == "" {
		fmt.Fprintln(stderr, "arena validate requires --input")
		return 2
	}
	record, err := adaptive.LoadArenaDecision(*input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeJSON(stdout, stderr, map[string]any{
		"status": "PASS",
		"record": record,
	})
}

func runProbe(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "probe requires validate|plan|calendar")
		return 2
	}
	switch args[0] {
	case "validate":
		flags := flag.NewFlagSet("probe validate", flag.ContinueOnError)
		flags.SetOutput(stderr)
		input := flags.String("input", "", "adversarial probe plan JSON file")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *input == "" {
			fmt.Fprintln(stderr, "probe validate requires --input")
			return 2
		}
		plan, err := adaptive.LoadAdversarialProbePlan(*input)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, map[string]any{"status": "PASS", "plan": plan})
	case "plan":
		flags := flag.NewFlagSet("probe plan", flag.ContinueOnError)
		flags.SetOutput(stderr)
		revision := flags.String("revision", "", "revision under probe")
		risk := flags.String("risk", "high", "medium or high")
		requirements := flags.String("requirements", "", "comma-separated requirement IDs")
		calendar := flags.Bool("calendar", true, "include calendar-validity probe")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *revision == "" {
			fmt.Fprintln(stderr, "probe plan requires --revision")
			return 2
		}
		var ids []string
		if strings.TrimSpace(*requirements) != "" {
			for _, part := range strings.Split(*requirements, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					ids = append(ids, part)
				}
			}
		}
		plan, err := adaptive.DeriveDefaultProbes(*revision, *risk, ids, *calendar)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, plan)
	case "calendar":
		flags := flag.NewFlagSet("probe calendar", flag.ContinueOnError)
		flags.SetOutput(stderr)
		date := flags.String("date", "", "YYYY-MM-DD value to validate")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *date == "" {
			fmt.Fprintln(stderr, "probe calendar requires --date")
			return 2
		}
		probe := adaptive.EvaluateCalendarProbe(*date)
		if code := writeJSON(stdout, stderr, probe); code != 0 {
			return code
		}
		if probe.Status != adaptive.TerminalPass {
			return 1
		}
		return 0
	default:
		fmt.Fprintln(stderr, "probe requires validate|plan|calendar")
		return 2
	}
}

func runParallel(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "validate" {
		fmt.Fprintln(stderr, "parallel requires validate")
		return 2
	}
	flags := flag.NewFlagSet("parallel validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	input := flags.String("input", "", "parallel orchestration record JSON file")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *input == "" {
		fmt.Fprintln(stderr, "parallel validate requires --input")
		return 2
	}
	record, err := adaptive.LoadParallelOrchestration(*input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeJSON(stdout, stderr, map[string]any{"status": "PASS", "record": record})
}

func runPlaybook(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "playbook requires list|show")
		return 2
	}
	switch args[0] {
	case "list":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "playbook list accepts no arguments")
			return 2
		}
		return writeJSON(stdout, stderr, adaptive.ListPlaybooks())
	case "show":
		flags := flag.NewFlagSet("playbook show", flag.ContinueOnError)
		flags.SetOutput(stderr)
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 1 {
			fmt.Fprintln(stderr, "playbook show requires an id")
			return 2
		}
		playbook, err := adaptive.GetPlaybook(flags.Arg(0))
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, playbook)
	default:
		fmt.Fprintln(stderr, "playbook requires list|show")
		return 2
	}
}
