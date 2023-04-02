// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"github.com/go-rod/rod"
	"github.com/stretchr/testify/suite"
)

func NewRodSuite(name string) *RodSuite {
	return &RodSuite{
		BaseSuite: &BaseSuite{
			Name: name,
		},
	}
}

// RodSuite is a go-rod suite.
type RodSuite struct {
	*BaseSuite

	*RodSession
	*rod.Page
}

type BaseSuite struct {
	suite.Suite

	Name string
}

// CommandSuite is a command line interface suite.
type CommandSuite struct {
	*BaseSuite

	*DockerEnvironment
}
