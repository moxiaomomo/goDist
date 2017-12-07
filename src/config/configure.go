package config

import (
	"common"
	"logger"
)

type LBConfig struct {
	LogLevel common.LogLevelEnum
	LBPolicy common.LBPolicyEnum
}

var globalLBConfig LBConfig = LBConfig{
	LogLevel: common.LOG_INFO,
	LBPolicy: common.LB_RANDOM,
}

func GlobalLBConfig() LBConfig {
	return globalLBConfig
}

func SetGlobalLBConfig(m map[string]interface{}) error {
	var lastErr error = nil
	for k, v := range m {
		err := common.SetStructField(&globalLBConfig, k, v)
		if err != nil {
			lastErr = err
			logger.LogErrorf("Set Config Failed: %s %v", k, v)
		}
	}
	return lastErr
}
