package main

import "github.com/spf13/cobra"

// RunUnitTest run the unit tests
func RunUnitTest(cobraCmd *cobra.Command, args []string) {
	err := CommandWithStdout("go", "test", "./...").Run()
	if err != nil {
		panic(err)
	}
}
