package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/futurematic/kernel/cmd/dot/client"
	"github.com/futurematic/kernel/internal/domain"
)

// planAndApply executes the plan/apply workflow
func planAndApply(ctx *CommandContext, intents []domain.Intent, dryRun, yes bool) error {
	// Create plan request
	planReq := client.PlanRequest{
		ActorID:     ctx.Config.ActorID,
		Capabilities: ctx.Config.Capabilities,
		NamespaceID: getNamespaceID(ctx.Config),
		AsOf:        domain.AsOf{},
		Intents:     intents,
	}

	// Call plan
	plan, err := ctx.Client.Plan(planReq)
	if err != nil {
		handleError(err)
		return nil
	}

	// Print plan (convert to PlanResponse for formatter)
	planResp := &client.PlanResponse{
		ID:           plan.ID,
		CreatedAt:    plan.CreatedAt,
		ActorID:      plan.ActorID,
		NamespaceID:  plan.NamespaceID,
		AsOfSeq:      plan.AsOfSeq,
		Intents:      plan.Intents,
		Expanded:     plan.Expanded,
		Class:        plan.Class,
		PolicyReport: plan.PolicyReport,
		Hash:         plan.Hash,
	}
	if err := ctx.Formatter.PrintPlan(planResp); err != nil {
		return err
	}

	// Check for denies
	if plan.PolicyReport.HasDenies() {
		fmt.Fprintf(os.Stderr, "Policy denies detected. Exiting.\n")
		os.Exit(2)
	}

	// If dry-run, exit here
	if dryRun {
		return nil
	}

	// Prompt for confirmation unless --yes
	if !yes {
		fmt.Print("Apply this plan? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Apply plan
	applyReq := client.ApplyRequest{
		ActorID:     ctx.Config.ActorID,
		Capabilities: ctx.Config.Capabilities,
		PlanID:      plan.ID,
		PlanHash:    plan.Hash,
	}

	op, err := ctx.Client.Apply(applyReq)
	if err != nil {
		handleError(err)
		return nil
	}

	// Print operation (convert to ApplyResponse for formatter)
	opResp := &client.ApplyResponse{
		ID:           op.ID,
		Seq:          op.Seq,
		OccurredAt:   op.OccurredAt,
		ActorID:      op.ActorID,
		Capabilities: op.Capabilities,
		PlanID:       op.PlanID,
		PlanHash:     op.PlanHash,
		Class:        op.Class,
		Changes:      op.Changes,
	}
	return ctx.Formatter.PrintOperation(opResp)
}
