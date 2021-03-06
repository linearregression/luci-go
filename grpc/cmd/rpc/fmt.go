// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/subcommands"
	"golang.org/x/net/context"

	"github.com/luci/luci-go/common/auth"
	"github.com/luci/luci-go/common/cli"
)

func cmdFmt(defaultAuthOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `fmt subcommand [arguments]`,
		ShortDesc: "converts a message to/from flagpb and JSON formats",
		LongDesc:  "Converts a message to/from flagpb and JSON formats.",
		CommandRun: func() subcommands.CommandRun {
			c := &fmtRun{defaultAuthOpts: defaultAuthOpts}
			return c
		},
	}
}

type fmtRun struct {
	cmdRun

	defaultAuthOpts auth.Options
}

func (r *fmtRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	app := &cli.Application{
		Name: "fmt",
		Context: func(context.Context) context.Context {
			return cli.GetContext(a, r, env)
		},
		Title: "Converts a message formats.",
		Commands: []*subcommands.Command{
			cmdJ2F(r.defaultAuthOpts),
			cmdF2J(r.defaultAuthOpts),
		},
	}
	return subcommands.Run(app, args)
}
