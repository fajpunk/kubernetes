/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"fmt"
)

// GetResourceUtilizationRatio takes in a set of metrics, a set of matching requests,
// and a target utilization percentage, and calculates the ratio of
// desired to actual utilization (returning that, the actual utilization, and the raw average value)
func GetResourceUtilizationRatio(metrics PodMetricsInfo, requests map[string]int64, targetUtilization int32) (utilizationRatio float64, currentUtilization int32, rawAverageValue int64, err error) {
	metricsTotal := int64(0)
	requestsTotal := int64(0)
	numEntries := 0

	for podName, metric := range metrics {
		request, hasRequest := requests[podName]
		if !hasRequest {
			// we check for missing requests elsewhere, so assuming missing requests == extraneous metrics
			continue
		}

		metricsTotal += metric.Value
		requestsTotal += request
		numEntries++
	}

	// if the set of requests is completely disjoint from the set of metrics,
	// then we could have an issue where the requests total is zero
	if requestsTotal == 0 {
		return 0, 0, 0, fmt.Errorf("no metrics returned matched known pods")
	}

	currentUtilization = int32((metricsTotal * 100) / requestsTotal)

	return float64(currentUtilization) / float64(targetUtilization), currentUtilization, metricsTotal / int64(numEntries), nil
}

// GetMetricUsageRatio takes in a set of metrics and a target usage range (which may have the same value for the upper
// and lower bounds, which signifies a single target value), and calculates the ratio of desired to actual usage
// (returning that and the actual usage)
func GetMetricUsageRatio(metrics PodMetricsInfo, targetUsageLower int64, targetUsageUpper int64) (usageRatio float64, currentUsage int64) {
	metricsTotal := int64(0)
	for _, metric := range metrics {
		metricsTotal += metric.Value
	}

	currentUsage = metricsTotal / int64(len(metrics))

	var targetUsage int64

	if currentUsage >= targetUsageLower && currentUsage <= targetUsageUpper {
		// The current usage is in the acceptable range so return 1 to signify no scaling
		// is necessary in either direction
		targetUsage = currentUsage
	} else if targetUsageLower == targetUsageUpper {
		// We have a single metric
		targetUsage = targetUsageLower
	} else if currentUsage < targetUsageLower {
		// For scale down, use the lower bound of the range to calculate scale percentage
		targetUsage = targetUsageLower
	} else if currentUsage > targetUsageUpper {
		// For scale up, use the upper bound of the range to calculate scale percentage
		targetUsage = targetUsageUpper
	}

	return float64(currentUsage) / float64(targetUsage), currentUsage
}
