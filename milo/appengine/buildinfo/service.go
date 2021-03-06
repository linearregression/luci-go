// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package buildinfo

import (
	"github.com/luci/luci-go/grpc/grpcutil"
	"github.com/luci/luci-go/luci_config/common/cfgtypes"
	milo "github.com/luci/luci-go/milo/api/proto"
	"github.com/luci/luci-go/milo/appengine/buildbot"
	"github.com/luci/luci-go/milo/appengine/swarming"

	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
)

// Service is a BuildInfoServer implementation.
type Service struct {
	// BuildBot is the BuildInfoProvider for the BuildBot service.
	BuildBot buildbot.BuildInfoProvider
	// Swarming is the BuildInfoProvider for the Swarming service.
	Swarming swarming.BuildInfoProvider
}

var _ milo.BuildInfoServer = (*Service)(nil)

// Get implements milo.BuildInfoServer.
func (svc *Service) Get(c context.Context, req *milo.BuildInfoRequest) (*milo.BuildInfoResponse, error) {
	projectHint := cfgtypes.ProjectName(req.ProjectHint)
	if projectHint != "" {
		if err := projectHint.Validate(); err != nil {
			return nil, grpcutil.Errf(codes.InvalidArgument, "invalid project hint: %s", err.Error())
		}
	}

	switch {
	case req.GetBuildbot() != nil:
		resp, err := svc.BuildBot.GetBuildInfo(c, req.GetBuildbot(), projectHint)
		if err != nil {
			return nil, err
		}
		return resp, nil

	case req.GetSwarming() != nil:
		resp, err := svc.Swarming.GetBuildInfo(c, req.GetSwarming(), projectHint)
		if err != nil {
			return nil, err
		}
		return resp, nil

	default:
		return nil, grpcutil.Errf(codes.InvalidArgument, "must supply a build")
	}
}
