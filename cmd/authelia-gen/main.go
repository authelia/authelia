// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
