// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/maruel/subcommands"

	"github.com/luci/luci-go/common/auth"
	"github.com/luci/luci-go/common/cli"
)

func cmdCancel(defaultAuthOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `cancel [flags] <build id>`,
		ShortDesc: "cancel a build",
		LongDesc:  "Cancel a build.",
		CommandRun: func() subcommands.CommandRun {
			r := &cancelRun{}
			r.SetDefaultFlags(defaultAuthOpts)
			return r
		},
	}
}

type cancelRun struct {
	baseCommandRun
	buildIDArg
}

func (r *cancelRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)

	if err := r.parseArgs(args); err != nil {
		return r.done(ctx, err)
	}

	return r.callAndDone(ctx, "POST", fmt.Sprintf("builds/%d/cancel", r.buildID), nil)
}
