package commands

import (
	"fmt"
	"strconv"

	"github.com/futurematic/kernel/cmd/dot/client"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <a> <b> <target>",
		Short: "Get differences between two sequence points",
		Long:  "Show changes between two sequence numbers. Use 'now' to reference the latest sequence.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			aStr := args[0]
			bStr := args[1]
			target := args[2]

			// Resolve 'now' alias
			var aSeq, bSeq int64
			if aStr == "now" {
				// Get latest seq from history
				historyReq := client.HistoryRequest{Target: target, Limit: 1}
				ops, err := ctx.Client.History(historyReq)
				if err != nil {
					return fmt.Errorf("failed to resolve 'now': %w", err)
				}
				if len(ops) == 0 {
					return fmt.Errorf("no operations found for target")
				}
				aSeq = ops[0].Seq
			} else {
				seq, err := strconv.ParseInt(aStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid sequence number: %s", aStr)
				}
				aSeq = seq
			}

			if bStr == "now" {
				historyReq := client.HistoryRequest{Target: target, Limit: 1}
				ops, err := ctx.Client.History(historyReq)
				if err != nil {
					return fmt.Errorf("failed to resolve 'now': %w", err)
				}
				if len(ops) == 0 {
					return fmt.Errorf("no operations found for target")
				}
				bSeq = ops[0].Seq
			} else {
				seq, err := strconv.ParseInt(bStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid sequence number: %s", bStr)
				}
				bSeq = seq
			}

			req := client.DiffRequest{
				ASeq:   aSeq,
				BSeq:   bSeq,
				Target: target,
			}

			result, err := ctx.Client.Diff(req)
			if err != nil {
				handleError(err)
				return nil
			}

			return ctx.Formatter.PrintDiff(result)
		},
	}

	return cmd
}
