// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

//usr/bin/env go run "$0" "$@"; exit
//nolint:gocritic

package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/cmd/authelia-scripts/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}
