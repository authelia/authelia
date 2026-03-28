package cmd

import (
	"fmt"
	"go/ast"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

func newFuzzTestCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "fuzztest",
		Short:   cmdFuzzTestShort,
		Long:    cmdFuzzTestLong,
		Example: cmdFuzzTestExample,
		Args:    cobra.ExactArgs(1),
		RunE:    cmdFuzzTestRunE,

		DisableAutoGenTag: true,
	}

	cmd.Flags().Duration("individual-budget", 2*time.Minute, "Individual budget for each fuzz test.")
	cmd.Flags().Duration("total-budget", 8*time.Minute, "Total budget for all fuzz tests.")

	return cmd
}

func cmdFuzzTestRunE(cmd *cobra.Command, args []string) (err error) {
	var individualBudget, totalBudget time.Duration

	if individualBudget, err = cmd.Flags().GetDuration("individual-budget"); err != nil {
		return err
	}

	if totalBudget, err = cmd.Flags().GetDuration("total-budget"); err != nil {
		return err
	}

	failed := false

	tests, err := findFuzzTests(args[0], nil, individualBudget, totalBudget)

	for _, test := range tests {
		pkg := filepath.Dir(test.File)

		//nolint:gosec // False positive.
		ecmd := exec.Command(
			"go",
			"test",
			"-run", "^$",
			"-fuzz", "^"+test.Name+"$",
			"-fuzztime", test.Duration.String(),
			pkg,
		)

		ecmd.Env = os.Environ()
		ecmd.Stdout = os.Stdout
		ecmd.Stderr = os.Stderr

		fmt.Printf("--- :game_die: Running %s for %s\n\n > Command: %s\n", test.Name, test.Duration.String(), ecmd.String())

		if err = ecmd.Run(); err != nil {
			fmt.Println("^^^ +++")
			fmt.Printf("FAILED: :boom: Failed to run %s for %s\n\n > Error: %s\n", test.Name, test.Duration.String(), err.Error())

			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}

	return
}

type fuzzTest struct {
	Name     string
	File     string
	Duration time.Duration
}

type fuzzEntry struct {
	name string
	file string
}

func findFuzzTests(pattern string, weights map[string]float64, individualBudget, totalBudget time.Duration) ([]fuzzTest, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
		Tests: true,
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, err
	}

	var found []fuzzEntry

	for _, pkg := range pkgs {
		for i, file := range pkg.Syntax {
			fileName := pkg.GoFiles[i]

			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name == nil || !isFuzzFunc(fn) {
					continue
				}

				found = append(found, fuzzEntry{
					name: fn.Name.Name,
					file: fileName,
				})
			}
		}
	}

	return allocateTimes(found, weights, individualBudget, totalBudget), nil
}

func isFuzzFunc(fn *ast.FuncDecl) bool {
	if fn.Name == nil || len(fn.Name.Name) < 5 {
		return false
	}

	if fn.Name.Name[:4] != "Fuzz" {
		return false
	}

	params := fn.Type.Params
	if params == nil || len(params.List) != 1 {
		return false
	}

	star, ok := params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	sel, ok := star.X.(*ast.SelectorExpr)

	return ok && sel.Sel.Name == "F"
}

func allocateTimes(found []fuzzEntry, weights map[string]float64, individualBudget, totalBudget time.Duration) []fuzzTest {
	if len(found) == 0 {
		return nil
	}

	const minPerTest = 5 * time.Second

	reserved := minPerTest * time.Duration(len(found))
	if reserved >= totalBudget {
		each := totalBudget / time.Duration(len(found))
		tests := make([]fuzzTest, len(found))

		duration := each
		if individualBudget < each {
			duration = individualBudget
		}

		for i, e := range found {
			tests[i] = fuzzTest{Name: e.name, File: e.file, Duration: duration}
		}

		return tests
	}

	remainder := totalBudget - reserved

	totalWeight := 0.0

	for _, e := range found {
		totalWeight += weightOf(e.name, weights)
	}

	tests := make([]fuzzTest, len(found))

	var allocated time.Duration

	for i, e := range found {
		var extra time.Duration

		if i == len(found)-1 {
			extra = remainder - allocated
		} else {
			extra = time.Duration(math.Round(float64(remainder) * weightOf(e.name, weights) / totalWeight))
			allocated += extra
		}

		d := minPerTest + extra
		if individualBudget > 0 && d > individualBudget {
			d = individualBudget
		}

		tests[i] = fuzzTest{Name: e.name, File: e.file, Duration: d}
	}

	return tests
}

func weightOf(name string, weights map[string]float64) float64 {
	if weights != nil {
		if w, ok := weights[name]; ok && w > 0 {
			return w
		}
	}

	return 1.0
}
