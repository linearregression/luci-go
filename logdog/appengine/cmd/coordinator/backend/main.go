// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package module

import (
	"net/http"

	// Importing pprof implicitly installs "/debug/*" profiling handlers.
	_ "net/http/pprof"

	"github.com/luci/luci-go/appengine/gaemiddleware"
	"github.com/luci/luci-go/logdog/appengine/coordinator"
	"github.com/luci/luci-go/server/router"
	"github.com/luci/luci-go/tumble"

	// Include mutations package so its Mutations will register with tumble via
	// init().
	_ "github.com/luci/luci-go/logdog/appengine/coordinator/mutations"
)

func init() {
	tmb := tumble.Service{}

	r := router.New()
	base := gaemiddleware.BaseProd().Extend(coordinator.ProdCoordinatorService)
	tmb.InstallHandlers(r, base)
	gaemiddleware.InstallHandlersWithMiddleware(r, base)

	http.Handle("/", r)
}
