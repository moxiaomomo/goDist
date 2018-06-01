package config

import (
	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"
)

type LBConfig struct {
	LogLevel util.LogLevelEnum
	LBPolicy util.LBPolicyEnum
}

var globalLBConfig LBConfig = LBConfig{
	LogLevel: util.LOG_INFO,
	LBPolicy: util.LB_RANDOM,
}

func GlobalLBConfig() LBConfig {
	return globalLBConfig
}

func SetGlobalLBConfig(m map[string]interface{}) error {
	var lastErr error = nil
	for k, v := range m {
		err := util.SetStructField(&globalLBConfig, k, v)
		if err != nil {
			lastErr = err
			logger.LogErrorf("Set Config Failed: %s %v", k, v)
		}
	}
	return lastErr
}
