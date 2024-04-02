/*

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type GPUMetrics struct {
	alloc       uint64
	idle        uint64
	total       uint64
	utilization float64
}

func GPUGetMetrics() map[string]*GPUMetrics {
	args := []string{"-a", "-h", "--Format='Nodes: ,Gres: ,GresUsed:'", "--state=idle,allocated"}
	output := string(Execute("sinfo", args))
	return ParseGPUMetrics(output)
}

func ParseGPUMetrics(input string) map[string]*GPUMetrics {
	var gpus = make(map[string]*GPUMetrics)
	var ptrn_gpu = regexp.MustCompile(`^gpu:([^:]+):(\d+).*$`)
	var ptrn_gpu_used = regexp.MustCompile(`^gpu:([^:]+):(\d+).*$`)

	for _, row := range strings.Split(input, "\n") {
		var gpu_type string
		var gpu_used_type string
		var row_gpus uint64
		var row_gpus_used uint64
		tokens := strings.Split(row, " ")
		if len(tokens) < 3 {
			continue
		}
		nodes, err := strconv.ParseUint(tokens[0], 10, 32)
		if err != nil {
			log.Printf("Invalid number of nodes in '%s'.\n", row)
			continue
		}
		s_gpu, s_gpu_used := tokens[1], tokens[2]
		m_gpu := ptrn_gpu.FindStringSubmatch(s_gpu)
		if m_gpu != nil {
			gpu_type = m_gpu[1]
			// It shouldn't fail because the regexp only matches digits
			row_gpus, _ = strconv.ParseUint(m_gpu[2], 10, 32)
		} else {
			continue
		}
		m_gpu_used := ptrn_gpu_used.FindStringSubmatch(s_gpu_used)
		if m_gpu_used != nil {
			gpu_used_type = m_gpu_used[1]
			// It shouldn't fail because the regexp only matches digits
			row_gpus_used, _ = strconv.ParseUint(m_gpu_used[2], 10, 32)
		} else {
			continue
		}
		if gpu_type != gpu_used_type {
			log.Printf("GPU types in Gres '%s' and GresUsed '%s' do not match\n", gpu_type, gpu_used_type)
			continue
		}
		row_gpus_idle := row_gpus - row_gpus_used
		_, ok := gpus[gpu_type]
		if !ok {
			gpus[gpu_type] = &GPUMetrics{row_gpus_used * nodes, row_gpus_idle * nodes, row_gpus * nodes, 0.0}
		} else {
			gpus[gpu_type].alloc += row_gpus_used * nodes
			gpus[gpu_type].idle += row_gpus_idle * nodes
			gpus[gpu_type].total += row_gpus * nodes
		}
	}
	for key, _ := range gpus {
		if gpus[key].total > 0 {
			gpus[key].utilization = float64(gpus[key].alloc) / float64(gpus[key].total)
		}
	}
	return gpus
}

/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm scheduler metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewGPUCollector() *GPUCollector {
	labels := []string{"gpu"}

	return &GPUCollector{
		alloc: prometheus.NewDesc("slurm_gpu_type_alloc", "Allocated GPUs per type", labels, nil),
		idle:  prometheus.NewDesc("slurm_gpu_type_idle", "Idle GPUs per type", labels, nil),
		total: prometheus.NewDesc("slurm_gpu_type_total", "Total GPUs per type", labels, nil),
		utilization: prometheus.NewDesc("slurm_gpu_type_utilization", "GPU utilization per type", labels, nil),
	}
}

type GPUCollector struct {
	alloc       *prometheus.Desc
	idle        *prometheus.Desc
	total       *prometheus.Desc
	utilization *prometheus.Desc
}

// Send all metric descriptions
func (cc *GPUCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.alloc
	ch <- cc.idle
	ch <- cc.total
	ch <- cc.utilization
}

func (cc *GPUCollector) Collect(ch chan<- prometheus.Metric) {
	cm := GPUGetMetrics()
	for gpu_type := range cm {
		ch <- prometheus.MustNewConstMetric(cc.alloc, prometheus.GaugeValue, float64(cm[gpu_type].alloc))
		ch <- prometheus.MustNewConstMetric(cc.idle, prometheus.GaugeValue, float64(cm[gpu_type].idle))
		ch <- prometheus.MustNewConstMetric(cc.total, prometheus.GaugeValue, float64(cm[gpu_type].total))
		ch <- prometheus.MustNewConstMetric(cc.utilization, prometheus.GaugeValue, cm[gpu_type].utilization)
	}
}
