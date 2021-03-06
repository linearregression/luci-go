// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package grpcmon

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/luci/luci-go/common/clock"
	"github.com/luci/luci-go/common/tsmon/distribution"
	"github.com/luci/luci-go/common/tsmon/field"
	"github.com/luci/luci-go/common/tsmon/metric"
	"github.com/luci/luci-go/common/tsmon/types"
)

var (
	grpcClientCount = metric.NewCounter(
		"grpc/client/count",
		"Total number of RPCs.",
		nil,
		field.String("method"), // full name of the grpc method
		field.Int("code"))      // grpc.Code of the result

	grpcClientDuration = metric.NewCumulativeDistribution(
		"grpc/client/duration",
		"Distribution of client-side RPC duration (in milliseconds).",
		&types.MetricMetadata{Units: types.Milliseconds},
		distribution.DefaultBucketer,
		field.String("method"), // full name of the grpc method
		field.Int("code"))      // grpc.Code of the result
)

// NewUnaryClientInterceptor returns an interceptor that gathers RPC call
// metrics and sends them to tsmon.
//
// It can be optionally chained with other interceptor. The reported metrics
// include time spent in this other interceptor too.
//
// Can be passed to a gRPC client via WithUnaryInterceptor(...) dial option.
//
// Use option.WithGRPCDialOption(grpc.WithUnaryInterceptor(...)) when
// instrumenting Google Cloud API clients.
//
// It assumes the RPC context has tsmon initialized already.
func NewUnaryClientInterceptor(next grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		started := clock.Now(ctx)
		defer func() {
			reportClientRPCMetrics(ctx, method, err, clock.Now(ctx).Sub(started))
		}()
		if next != nil {
			return next(ctx, method, req, reply, cc, invoker, opts...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// reportClientRPCMetrics sends metrics after RPC call has finished.
func reportClientRPCMetrics(ctx context.Context, method string, err error, dur time.Duration) {
	code := 0
	if err != nil {
		code = int(grpc.Code(err))
	}
	grpcClientCount.Add(ctx, 1, method, code)
	grpcClientDuration.Add(ctx, float64(dur.Nanoseconds()/1e6), method, code)
}
