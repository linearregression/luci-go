// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cli

import (
	"io"
	"os"

	"github.com/luci/luci-go/common/clock"
	"github.com/luci/luci-go/common/errors"
	log "github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/logdog/api/logpb"
	"github.com/luci/luci-go/logdog/client/coordinator"
	"github.com/luci/luci-go/logdog/common/renderer"
	"github.com/luci/luci-go/logdog/common/types"

	"github.com/maruel/subcommands"
	"golang.org/x/net/context"
)

type latestCommandRun struct {
	subcommands.CommandRunBase

	raw bool
}

func newLatestCommand() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "latest [options] stream",
		ShortDesc: "Write the latest full log record in a stream to STDOUT.",
		LongDesc: "Write the latest full log record in a stream to STDOUT. If the stream " +
			"doesn't have any log entries, will block until a log entry is available.",
		CommandRun: func() subcommands.CommandRun {
			cmd := &latestCommandRun{}

			cmd.Flags.BoolVar(&cmd.raw, "raw", false,
				"Reproduce original log stream, instead of attempting to render for humans.")
			return cmd
		},
	}
}

func (cmd *latestCommandRun) Run(scApp subcommands.Application, args []string, _ subcommands.Env) int {
	a := scApp.(*application)

	// User-friendly: trim any leading or trailing slashes from the path.
	if len(args) != 1 {
		log.Errorf(a, "Exactly one argument, the stream path, must be supplied.")
		return 1
	}

	var addr *types.StreamAddr
	var err error
	if addr, err = types.ParseURL(args[0]); err != nil {
		// Not a log stream address.
		project, path, _, err := a.splitPath(args[0])
		if err != nil {
			log.WithError(err).Errorf(a, "Invalid path specifier.")
			return 1
		}

		addr = &types.StreamAddr{Project: project, Path: types.StreamPath(path)}
		if err := addr.Path.Validate(); err != nil {
			log.Fields{
				log.ErrorKey: err,
				"project":    addr.Project,
				"path":       addr.Path,
			}.Errorf(a, "Invalid command-line stream path.")
			return 1
		}
	}

	coord, err := a.coordinatorClient(addr.Host)
	if err != nil {
		errors.Log(a, errors.Annotate(err).Reason("failed to create Coordinator client").Err())
		return 1
	}

	stream := coord.Stream(addr.Project, addr.Path)

	tctx, _ := a.timeoutCtx(a)
	le, st, err := cmd.getTailEntry(tctx, stream)
	if err != nil {
		log.Fields{
			log.ErrorKey: err,
			"project":    addr.Project,
			"path":       addr.Path,
		}.Errorf(a, "Failed to load latest record.")

		if err == context.DeadlineExceeded {
			return 2
		}
		return 1
	}

	// Render the entry.
	r := renderer.Renderer{
		Source:         &renderer.StaticSource{le},
		Raw:            cmd.raw,
		DatagramWriter: getDatagramWriter(a, &st.Desc),
	}
	if _, err := io.Copy(os.Stdout, &r); err != nil {
		log.WithError(err).Errorf(a, "failed to write to output")
		return 1
	}

	return 0
}

func (cmd *latestCommandRun) getTailEntry(c context.Context, s *coordinator.Stream) (
	*logpb.LogEntry, *coordinator.LogStream, error) {

	// Loop until we either hard fail or succeed.
	var st coordinator.LogStream
	for {
		ls, err := s.Tail(c, coordinator.Complete(), coordinator.WithState(&st))
		switch {
		case err == nil:
			return ls, &st, nil

		case err == coordinator.ErrNoSuchStream, ls == nil:
			log.WithError(err).Warningf(c, "No log entries, sleeping and retry.")

			if ar := <-clock.After(c, noStreamDelay); ar.Incomplete() {
				// Timer stopped prematurely.
				return nil, nil, ar.Err
			}

		default:
			return nil, nil, err
		}
	}
}
