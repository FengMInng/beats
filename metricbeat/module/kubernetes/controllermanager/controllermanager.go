// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package controllermanager

import (
	"fmt"

	"github.com/elastic/beats/v7/metricbeat/helper/prometheus"
	"github.com/elastic/beats/v7/metricbeat/mb"
	k8smod "github.com/elastic/beats/v7/metricbeat/module/kubernetes"
	"github.com/elastic/beats/v7/metricbeat/module/kubernetes/util"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

var mapping = &prometheus.MetricsMapping{
	Metrics: map[string]prometheus.MetricMap{
		"process_cpu_seconds_total":     prometheus.Metric("process.cpu.sec"),
		"process_resident_memory_bytes": prometheus.Metric("process.memory.resident.bytes"),
		"process_virtual_memory_bytes":  prometheus.Metric("process.memory.virtual.bytes"),
		"process_open_fds":              prometheus.Metric("process.fds.open.count"),
		"process_max_fds":               prometheus.Metric("process.fds.max.count"),
		"process_start_time_seconds":    prometheus.Metric("process.started.sec"),

		"rest_client_request_duration_seconds": prometheus.Metric("client.request.duration.us", prometheus.OpMultiplyBuckets(1000000)),
		"rest_client_requests_total":           prometheus.Metric("client.request.count"),
		"rest_client_response_size_bytes":      prometheus.Metric("client.response.size.bytes"),
		"rest_client_request_size_bytes":       prometheus.Metric("client.request.size.bytes"),

		"workqueue_longest_running_processor_seconds": prometheus.Metric("workqueue.longestrunning.sec"),
		"workqueue_unfinished_work_seconds":           prometheus.Metric("workqueue.unfinished.sec"),
		"workqueue_adds_total":                        prometheus.Metric("workqueue.adds.count"),
		"workqueue_depth":                             prometheus.Metric("workqueue.depth.count"),
		"workqueue_retries_total":                     prometheus.Metric("workqueue.retries.count"),

		"node_collector_evictions_total":         prometheus.Metric("node.collector.eviction.count"),
		"node_collector_unhealthy_nodes_in_zone": prometheus.Metric("node.collector.unhealthy.count"),
		"node_collector_zone_health":             prometheus.Metric("node.collector.health.pct"),
		"node_collector_zone_size":               prometheus.Metric("node.collector.count"),

		"leader_election_master_status": prometheus.BooleanMetric("leader.is_master"),
	},

	Labels: map[string]prometheus.LabelMap{
		"code":   prometheus.KeyLabel("code"),
		"method": prometheus.KeyLabel("method"),
		"host":   prometheus.KeyLabel("host"),
		"name":   prometheus.KeyLabel("name"),
		"zone":   prometheus.KeyLabel("zone"),
		"verb":   prometheus.KeyLabel("verb"),
	},
}

func init() {
	mb.Registry.MustAddMetricSet("kubernetes", "controllermanager", New,
		mb.WithHostParser(prometheus.HostParser))
}

// MetricSet type defines all fields of the MetricSet
// The event MetricSet listens to events from Kubernetes API server and streams them to the output.
// MetricSet implements the mb.PushMetricSet interface, and therefore does not rely on polling.
type MetricSet struct {
	mb.BaseMetricSet
	prometheusClient   prometheus.Prometheus
	prometheusMappings *prometheus.MetricsMapping
	clusterMeta        mapstr.M
	mod                k8smod.Module
}

// New create a new instance of the MetricSet
// Part of new is also setting up the configuration by processing additional
// configuration entries if needed.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	pc, err := prometheus.NewPrometheusClient(base)
	if err != nil {
		return nil, err
	}

	mod, ok := base.Module().(k8smod.Module)
	if !ok {
		return nil, fmt.Errorf("must be child of kubernetes module")
	}

	ms := &MetricSet{
		BaseMetricSet:      base,
		prometheusClient:   pc,
		prometheusMappings: mapping,
		clusterMeta:        util.AddClusterECSMeta(base),
		mod:                mod,
	}
	return ms, nil
}

// Fetch gathers information from the apiserver and reports events with this information.
func (m *MetricSet) Fetch(reporter mb.ReporterV2) error {
	events, err := m.prometheusClient.GetProcessedMetrics(m.prometheusMappings)
	if err != nil {
		return fmt.Errorf("error getting metrics: %w", err)
	}

	for _, e := range events {
		event := mb.TransformMapStrToEvent("kubernetes", e, nil)
		if len(m.clusterMeta) != 0 {
			event.RootFields.DeepUpdate(m.clusterMeta)
		}
		isOpen := reporter.Event(event)
		if !isOpen {
			return nil
		}

	}
	return nil
}
