package utils

import (
	"strconv"

	ctypes "github.com/docker/docker/api/types/container"
	"github.com/ehazlett/interlock/ext"
)

const (
	DefaultHealthCheckInterval = 5000
)

func HealthCheck(config *ctypes.Config) string {
	if v, ok := config.Labels[ext.InterlockHealthCheckLabel]; ok {
		return v
	}

	return ""
}

func HealthCheckInterval(config *ctypes.Config) (int, error) {
	checkInterval := DefaultHealthCheckInterval

	if v, ok := config.Labels[ext.InterlockHealthCheckIntervalLabel]; ok && v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return -1, err
		}
		if i != 0 {
			checkInterval = i
		}
	}

	return checkInterval, nil
}
