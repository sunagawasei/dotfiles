package verifycolors

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func WriteReport(writer io.Writer, verification Verification) {
	summary := verification.ContractSummary
	total := summary.Enforced + summary.ReportOnly + summary.Waived
	enforcedPercent := 0.0
	if total > 0 {
		enforcedPercent = float64(summary.Enforced) * 100 / float64(total)
	}
	reportBelow := 0
	for _, result := range verification.Evaluation.Results {
		if result.Class == ClassReportOnly && !result.Pass {
			reportBelow++
		}
	}

	fmt.Fprintln(writer, "=== COLOR CONTRACT SCOPE ===")
	fmt.Fprintf(
		writer,
		"ENFORCED     %d / %d pairs (%.1f%%): opaque primary truecolor pairs; AA 4.5 is an exit-code gate\n",
		summary.Enforced,
		total,
		enforcedPercent,
	)
	fmt.Fprintf(
		writer,
		"REPORT-ONLY  %d / %d pairs: ambient/environment-dependent pairs; ratios are observations, not guarantees\n",
		summary.ReportOnly,
		total,
	)
	fmt.Fprintf(
		writer,
		"WAIVED       %d / %d pairs: excluded only with a recorded alternate-cue or intentional-dimming reason\n",
		summary.Waived,
		total,
	)
	fmt.Fprintf(
		writer,
		"Effective report-only profile results below 4.5: %d (do not affect exit status)\n",
		reportBelow,
	)
	fmt.Fprintln(
		writer,
		"IMPORTANT: exit 0 means the enforced subset has no AA failures; it does NOT mean every on-screen color is AA.",
	)
	fmt.Fprintln(
		writer,
		"cterm note: index selection uses xterm-256; indices 0-15 resolve through the WezTerm custom ANSI palette, so other terminals are not guaranteed.",
	)
	fmt.Fprintln(
		writer,
		"surface note: each ratio is a guarantee only for the named environment profile; static values do not observe wallpaper, window position, or blur.",
	)
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "=== USE-SITE PAIRS ===")
	for _, result := range verification.Evaluation.Results {
		writeProfileResult(writer, result)
	}
	for _, pair := range verification.Contract.Pairs {
		if pair.Class != ClassWaived {
			continue
		}
		fmt.Fprintf(
			writer,
			"[WAIVED] %s role=%s reason=%s source=%s\n",
			pair.ConsumerID,
			pair.Role,
			pair.Reason,
			pair.Source,
		)
	}
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "=== TOKEN COVERAGE ===")
	fmt.Fprintf(
		writer,
		"referenced=%d dispositioned=%d unclassified=%d\n",
		summary.ReferencedToken,
		summary.Disposition,
		len(summary.Unclassified),
	)
	dispositions := append([]TokenDisposition(nil), verification.Contract.Dispositions...)
	sort.Slice(dispositions, func(i, j int) bool {
		return dispositions[i].Token < dispositions[j].Token
	})
	for _, disposition := range dispositions {
		fmt.Fprintf(
			writer,
			"[%s] %s reason=%s\n",
			disposition.Kind,
			disposition.Token,
			disposition.Reason,
		)
	}
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "=== COVERAGE NOTES ===")
	fmt.Fprintf(writer, "count=%d\n", len(verification.Contract.Coverage))
	for _, note := range verification.Contract.Coverage {
		fmt.Fprintf(writer, "[NOTE] %s reason=%s source=%s\n", note.ID, note.Reason, note.Source)
	}
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "=== UNKNOWN HEX LITERALS ===")
	fmt.Fprintf(writer, "checked_files=%d unknown_colors=%d\n", verification.HexScan.Checked, countUnknownColors(verification.HexScan))
	for _, issue := range verification.HexScan.Issues {
		fmt.Fprintf(writer, "[NG] %s: %s\n", issue.Path, strings.Join(issue.Colors, ", "))
	}
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "=== ANSI CVD SCREEN (REPORT-ONLY) ===")
	fmt.Fprintf(
		writer,
		"DeltaE76 >= %.0f is a project-specific screening threshold, not a WCAG requirement. Enforced semantic pairs=%d; initial set is empty because no color-only distinction could be justified.\n",
		CVDDeltaEThreshold,
		verification.CVD.Enforced,
	)
	fmt.Fprintf(
		writer,
		"matrix_pairs=%d below_screen=%d (do not affect exit status)\n",
		len(verification.CVD.Results),
		verification.CVD.BelowScreen,
	)
	for _, result := range verification.CVD.Results {
		status := "screen-pass"
		if result.BelowScreen {
			status = "below-screen"
		}
		fmt.Fprintf(
			writer,
			"[REPORT-ONLY %s] %s %s/%s DeltaE76=%.2f\n",
			status,
			result.Type,
			result.First,
			result.Second,
			result.DeltaE,
		)
	}
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "=== FINAL SUMMARY ===")
	fmt.Fprintf(
		writer,
		"enforced_pass=%d enforced_ng=%d report_only_results=%d report_only_below_4.5=%d waived_pairs=%d policy_failures=%d\n",
		verification.Evaluation.EnforcedPass,
		verification.Evaluation.EnforcedNG,
		verification.Evaluation.ReportOnly,
		reportBelow,
		verification.Evaluation.Waived,
		verification.PolicyFailures(),
	)
}

func writeProfileResult(writer io.Writer, result ProfileResult) {
	status := "PASS"
	if !result.Pass {
		status = "BELOW-AA"
	}
	classLabel := strings.ToUpper(string(result.Class))
	foregroundLabel := "foreground"
	backgroundLabel := "background"
	if result.Role == RoleSurface {
		foregroundLabel = "surface"
		backgroundLabel = "against"
	}
	foreground := fmt.Sprintf("%s=%s(%s)", foregroundLabel, result.ForegroundToken, result.ForegroundColor)
	if result.ForegroundIndex != nil {
		foreground += fmt.Sprintf("[idx=%d]", *result.ForegroundIndex)
	}
	backgroundToken := string(result.BackgroundToken)
	if backgroundToken == "" {
		backgroundToken = "environment:" + result.Environment
	}
	background := fmt.Sprintf("%s=%s(%s)", backgroundLabel, backgroundToken, result.BackgroundColor)
	if result.BackgroundIndex != nil {
		background += fmt.Sprintf("[idx=%d]", *result.BackgroundIndex)
	}
	effective := ""
	if result.Class != result.DeclaredClass {
		effective = fmt.Sprintf(" declared=%s", result.DeclaredClass)
	}
	fmt.Fprintf(
		writer,
		"[%s %s] %s profile=%s role=%s %s %s ratio=%.2f%s\n",
		classLabel,
		status,
		result.ConsumerID,
		result.Profile,
		result.Role,
		foreground,
		background,
		result.Ratio,
		effective,
	)
}

func countUnknownColors(result HexScanResult) int {
	count := 0
	for _, issue := range result.Issues {
		count += len(issue.Colors)
	}
	return count
}
