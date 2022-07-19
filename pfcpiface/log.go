// SPDX-License-Identifier: Apache-2.0
// Copyright 2022-present Open Networking Foundation

package pfcpiface

import "go.uber.org/zap"

var log *zap.SugaredLogger

// initialize the sugared zap logger
func Zap_init() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
	defer logger.Sync()
}
