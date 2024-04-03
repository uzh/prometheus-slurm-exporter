/* Copyright 2020 Joeri Hermans, Victor Penso, Matteo Dessalvi

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

type GPUsMetrics struct {
	alloc       uint64
	idle        uint64
	total       uint64
	utilization float64
}

func GPUsGetMetrics() *GPUsMetrics {
	args := []string{"-a", "-h", "--Format=Nodes: ,Gres: ,GresUsed: ", "--state=idle,allocated"}
	output := string(Execute("sinfo", args))
	return ParseGPUsMetrics(output)
}

func ParseGPUsMetrics(input string) *GPUsMetrics {
	var gm GPUsMetrics
	var total_gpus = uint64(0)
	var allocated_gpus = uint64(0)
	var ptrn_gpu = regexp.MustCompile(`^gpu:[^:]+:(\d+).*$`)
	var ptrn_gpu_used = regexp.MustCompile(`^gpu:[^:]+:(\d+).*$`)

	if len(input) > 0 {
		for _, row := range strings.Split(input, "\n") {
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
			row_gpus, row_gpus_used := uint64(0), uint64(0)
			m_gpu := ptrn_gpu.FindStringSubmatch(s_gpu)
			if m_gpu != nil {
				// It cannot fail because the regexp only matches digits
				row_gpus, _ = strconv.ParseUint(m_gpu[1], 10, 32)
			}
			m_gpu_used := ptrn_gpu_used.FindStringSubmatch(s_gpu_used)
			if m_gpu_used != nil {
				// It cannot fail because the regexp only matches digits
				row_gpus_used, _ = strconv.ParseUint(m_gpu_used[1], 10, 32)
			}
			total_gpus += row_gpus * nodes
			allocated_gpus += row_gpus_used * nodes
		}
	}
	utilization := 0.0
	if total_gpus > 0 {
		utilization = float64(allocated_gpus) / float64(total_gpus)
	}
	gm.alloc = allocated_gpus
	gm.idle = total_gpus - allocated_gpus
	gm.total = total_gpus
	gm.utilization = utilization
	return &gm
}


/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm scheduler metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewGPUsCollector() *GPUsCollector {
	return &GPUsCollector{
		alloc: prometheus.NewDesc("slurm_gpus_alloc", "Allocated GPUs", nil, nil),
		idle:  prometheus.NewDesc("slurm_gpus_idle", "Idle GPUs", nil, nil),
		total: prometheus.NewDesc("slurm_gpus_total", "Total GPUs", nil, nil),
		utilization: prometheus.NewDesc("slurm_gpus_utilization", "Total GPU utilization", nil, nil),
	}
}

type GPUsCollector struct {
	alloc       *prometheus.Desc
	idle        *prometheus.Desc
	total       *prometheus.Desc
	utilization *prometheus.Desc
}

// Send all metric descriptions
func (cc *GPUsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.alloc
	ch <- cc.idle
	ch <- cc.total
	ch <- cc.utilization
}
func (cc *GPUsCollector) Collect(ch chan<- prometheus.Metric) {
	cm := GPUsGetMetrics()
	ch <- prometheus.MustNewConstMetric(cc.alloc, prometheus.GaugeValue, float64(cm.alloc))
	ch <- prometheus.MustNewConstMetric(cc.idle, prometheus.GaugeValue, float64(cm.idle))
	ch <- prometheus.MustNewConstMetric(cc.total, prometheus.GaugeValue, float64(cm.total))
	ch <- prometheus.MustNewConstMetric(cc.utilization, prometheus.GaugeValue, cm.utilization)
}
