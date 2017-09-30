// Copyright (c) 2017 OysterPack, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nats_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/messaging/nats"
	"github.com/oysterpack/oysterpack.go/pkg/messaging/natstest"
	"github.com/oysterpack/oysterpack.go/pkg/metrics"

	"encoding/json"
	"fmt"
	"strings"
	"time"

	natsio "github.com/nats-io/go-nats"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
)

func TestConnManager_Metrics(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()
	server := natstest.RunServer()
	defer server.Shutdown()

	nats.RegisterMetrics()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()
	pubConn := mustConnect(t, connMgr)
	subConn := mustConnect(t, connMgr)

	const TOPIC = "TestConnManager_Metrics"

	subConn.Subscribe(TOPIC, func(msg *natsio.Msg) {
		t.Logf("received message : %v", string(msg.Data))
	})

	for _, healthcheck := range connMgr.HealthChecks() {
		result := healthcheck.Run()
		if !result.Success() {
			t.Errorf("healthcheck failed: %v", result)
		}
	}

	const MSG_COUNT = 10
	for i := 1; i <= MSG_COUNT; i++ {
		pubConn.Publish(TOPIC, []byte(fmt.Sprintf("#%d", i)))
	}

	metrics, err := metrics.Registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics : %v", err)
	}

	for _, metric := range metrics {
		if strings.HasPrefix(*metric.Name, nats.MetricsNamespace) {
			jsonBytes, _ := json.MarshalIndent(metric, "", "   ")
			t.Logf("%v", string(jsonBytes))
		}
	}

}

func TestConnManager_Metrics_Registered(t *testing.T) {
	metrics.ResetRegistry()
	defer metrics.ResetRegistry()

	nats.RegisterMetrics()

	backup := nats.DefaultReConnectTimeout
	const ReConnectTimeout = 10 * time.Millisecond
	nats.DefaultReConnectTimeout = natsio.ReconnectWait(ReConnectTimeout)
	defer func() { nats.DefaultReConnectTimeout = backup }()

	server := natstest.RunServer()
	defer server.Shutdown()

	connMgr := nats.NewConnManager()
	defer connMgr.CloseAll()
	pubConn := mustConnect(t, connMgr)
	subConn := mustConnect(t, connMgr)

	server.Shutdown()
	time.Sleep(10 * time.Millisecond)

	server = natstest.RunServer()
	defer server.Shutdown()

	time.Sleep(10 * time.Millisecond)

	const TOPIC = "TestConnManager_Metrics"

	ch := make(chan *natsio.Msg)
	subConn.ChanSubscribe(TOPIC, ch)

	pubConn.Publish(TOPIC, []byte("TEST"))
	pubConn.Flush()

	select {
	case <-ch:
	default:
	}

	gatheredMetrics, err := metrics.Registry.Gather()
	if value := *gauge(gatheredMetrics, nats.ConnCountOpts).GetMetric()[0].GetGauge().Value; value != 2 {
		t.Errorf("conn count is wrong : %v", value)
	}

	if value := *gauge(gatheredMetrics, nats.MsgsInGauge).GetMetric()[0].GetGauge().Value; uint64(value) != subConn.InMsgs {
		t.Errorf("msgs in is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.MsgsOutGauge).GetMetric()[0].GetGauge().Value; uint64(value) != pubConn.OutMsgs {
		t.Errorf("msgs out is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.BytesInGauge).GetMetric()[0].GetGauge().Value; uint64(value) != subConn.InBytes {
		t.Errorf("bytes in is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.BytesOutGauge).GetMetric()[0].GetGauge().Value; uint64(value) != pubConn.OutBytes {
		t.Errorf("bytes out is wrong : %v", value)
	}

	connMgr.CloseAll()
	time.Sleep(10 * time.Millisecond)

	gatheredMetrics, err = metrics.Registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather gatheredMetrics : %v", err)
	}

	for _, metric := range gatheredMetrics {
		if strings.HasPrefix(*metric.Name, nats.MetricsNamespace) {
			jsonBytes, _ := json.MarshalIndent(metric, "", "   ")
			t.Logf("%v", string(jsonBytes))
		}
	}

	for _, opts := range nats.MetricOpts.CounterOpts {
		if counter(gatheredMetrics, opts) == nil {
			t.Errorf("Metric was not gathered : %v", prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
		}
	}

	for _, opts := range nats.MetricOpts.GaugeOpts {
		if gauge(gatheredMetrics, opts) == nil {
			t.Errorf("Metric was not gathered : %v", prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
		}
	}

	if value := *counter(gatheredMetrics, nats.CreatedCounterOpts).GetMetric()[0].GetCounter().Value; value != 2 {
		t.Errorf("created count is wrong : %v", value)
	}

	if value := *counter(gatheredMetrics, nats.ClosedCounterOpts).GetMetric()[0].GetCounter().Value; value != 2 {
		t.Errorf("closed count is wrong : %v", value)
	}

	if value := *gauge(gatheredMetrics, nats.ConnCountOpts).GetMetric()[0].GetGauge().Value; value != 0 {
		t.Errorf("closed count is wrong : %v", value)
	}

	// disconnect events also occur on closing
	if value := *counter(gatheredMetrics, nats.DisconnectedCounterOpts).GetMetric()[0].GetCounter().Value; value != 4 {
		t.Errorf("disconnects count is wrong : %v", value)
	}

	if value := *counter(gatheredMetrics, nats.ReconnectedCounterOpts).GetMetric()[0].GetCounter().Value; value != 2 {
		t.Errorf("reconnects count is wrong : %v", value)
	}

	t.Logf("pubConn : %v", pubConn)
	t.Logf("subConn : %v", subConn)

	if value := *counter(gatheredMetrics, nats.SubscriberErrorCounterOpts).GetMetric()[0].GetCounter().Value; int(value) != subConn.Errors() {
		t.Errorf("subscriber error count is wrong : %v", value)
	}

	if value := *gauge(gatheredMetrics, nats.MsgsInGauge).GetMetric()[0].GetGauge().Value; uint64(value) != 0 {
		t.Errorf("msgs in is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.MsgsOutGauge).GetMetric()[0].GetGauge().Value; uint64(value) != 0 {
		t.Errorf("msgs out is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.BytesInGauge).GetMetric()[0].GetGauge().Value; uint64(value) != 0 {
		t.Errorf("bytes in is wrong : %v", value)
	}
	if value := *gauge(gatheredMetrics, nats.BytesOutGauge).GetMetric()[0].GetGauge().Value; uint64(value) != 0 {
		t.Errorf("bytes out is wrong : %v", value)
	}
}

func counter(metricFamilies []*io_prometheus_client.MetricFamily, opts *prometheus.CounterOpts) *io_prometheus_client.MetricFamily {
	name := prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
	for _, metric := range metricFamilies {
		if name == *metric.Name {
			return metric
		}
	}
	return nil
}

func gauge(metricFamilies []*io_prometheus_client.MetricFamily, opts *prometheus.GaugeOpts) *io_prometheus_client.MetricFamily {
	name := prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
	for _, metric := range metricFamilies {
		if name == *metric.Name {
			return metric
		}
	}
	return nil
}