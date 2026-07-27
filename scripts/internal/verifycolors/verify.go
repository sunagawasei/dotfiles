package verifycolors

type VerifyConfig struct {
	PalettePath  string
	ScanRoot     string
	ScanFiles    []string
	Environments []EnvironmentProfile
}

type Verification struct {
	Contract        Contract
	ContractSummary ContractSummary
	Evaluation      Evaluation
	HexScan         HexScanResult
	CVD             CVDReport
}

func Verify(config VerifyConfig) (Verification, error) {
	colorPalette, err := LoadPalette(config.PalettePath)
	if err != nil {
		return Verification{}, err
	}
	contract, err := DefaultContract()
	if err != nil {
		return Verification{}, err
	}
	contractSummary, err := ValidateContract(contract, colorPalette)
	if err != nil {
		return Verification{}, err
	}
	evaluation, err := Evaluate(contract, colorPalette, config.Environments)
	if err != nil {
		return Verification{}, err
	}
	hexScan, err := ScanHexLiterals(config.ScanRoot, config.ScanFiles, colorPalette)
	if err != nil {
		return Verification{}, err
	}
	cvdReport, err := EvaluateANSIColorVision(colorPalette)
	if err != nil {
		return Verification{}, err
	}
	return Verification{
		Contract:        contract,
		ContractSummary: contractSummary,
		Evaluation:      evaluation,
		HexScan:         hexScan,
		CVD:             cvdReport,
	}, nil
}

func (verification Verification) PolicyFailures() int {
	failures := verification.Evaluation.EnforcedNG
	for _, issue := range verification.HexScan.Issues {
		failures += len(issue.Colors)
	}
	return failures
}
