package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReadMEOptions supplies content for README.md in a bundle directory.
type ReadMEOptions struct {
	Trigger       string
	Timestamp     time.Time
	AgentID       string
	SessionID     string
	Outcome       string // e.g. "DENY", "STOP"
	ReasonCodes   []string
	ReasonSummary string
	PanicEnabled  bool
	PanicExpiresAt *time.Time
	NextSteps     []string
	BundleDirName string
}

// WriteREADME writes README.md into dir (and optionally README.txt). No secrets.
func WriteREADME(dir string, opts ReadMEOptions) error {
	outcome := opts.Outcome
	if outcome == "" {
		outcome = "DENY/STOP"
	}
	when := opts.Timestamp.UTC().Format(time.RFC3339)
	agent := opts.AgentID
	if agent == "" {
		agent = "N/A"
	}
	sess := opts.SessionID
	if sess == "" {
		sess = "N/A"
	}

	var why strings.Builder
	for i, c := range opts.ReasonCodes {
		if i > 0 {
			why.WriteString("\n")
		}
		why.WriteString(fmt.Sprintf("%d. %s", i+1, c))
		if opts.ReasonSummary != "" && i == 0 {
			why.WriteString(" — ")
			why.WriteString(opts.ReasonSummary)
		}
	}
	if why.Len() == 0 && opts.ReasonSummary != "" {
		why.WriteString(opts.ReasonSummary)
	}
	if why.Len() == 0 {
		why.WriteString("See decision_records.jsonl for details.")
	}

	panicLine := "off"
	if opts.PanicEnabled {
		panicLine = "on"
		if opts.PanicExpiresAt != nil {
			panicLine += " (expires " + opts.PanicExpiresAt.UTC().Format(time.RFC3339) + ")"
		}
	}

	var nextSteps string
	for _, s := range opts.NextSteps {
		if s == "" || strings.HasPrefix(s, "#") {
			nextSteps += s + "\n"
		} else {
			nextSteps += "- " + s + "\n"
		}
	}
	if nextSteps == "" {
		nextSteps = "- `ctrldot panic off`  # if appropriate\n"
		nextSteps += "- `ctrldot resolve allow-once --agent <agent_id> --ttl 120s`  # to allow one action\n"
	}

	verifyCmd := "ctrldot bundle verify ."
	if opts.BundleDirName != "" {
		verifyCmd = fmt.Sprintf("ctrldot bundle verify %s", opts.BundleDirName)
	}

	md := fmt.Sprintf(`# Ctrl Dot Bundle Summary

## What happened
- **Trigger:** %s
- **When:** %s
- **Agent:** %s
- **Session:** %s
- **Outcome:** %s

## Why
%s

## Panic mode
%s

## Suggested next steps
%s

## Verify this bundle
`+"```"+`
%s
`+"```"+`

## Notes
This bundle is signed. Share the entire directory.
`,
		opts.Trigger,
		when,
		agent,
		sess,
		outcome,
		why.String(),
		panicLine,
		nextSteps,
		verifyCmd,
	)

	readmePath := filepath.Join(dir, "README.md")
	return os.WriteFile(readmePath, []byte(md), 0644)
}
