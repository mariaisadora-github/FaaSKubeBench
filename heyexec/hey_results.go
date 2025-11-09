package heyexec

import "time"

// HeyResult representa a estrutura da saída JSON do hey
type HeyResult struct {
	URL                 string         `json:"url"`
	Requests            int            `json:"requests"`
	Duration            float64        `json:"duration"` // em segundos
	Wait                float64        `json:"wait"`     // em segundos
	Total               float64        `json:"total"`    // em segundos
	BytesTotal          int            `json:"bytes_total"`
	BytesPerSecond      float64        `json:"bytes_per_second"`
	RequestsPerSecond   float64        `json:"requests_per_second"`
	StatusCodeDist      map[string]int `json:"status_code_dist"`
	ErrorDist           map[string]int `json:"error_dist"`
	LatencyDistribution []struct {
		Percentage float64 `json:"percentage"`
		Latency    float64 `json:"latency"` // em segundos
	} `json:"latency_distribution"`
	Histogram []struct {
		Mark    float64 `json:"mark"`
		Count   int     `json:"count"`
		Percent float64 `json:"percent"`
	} `json:"histogram"`
	Fastest           float64 `json:"fastest"`
	Slowest           float64 `json:"slowest"`
	Average           float64 `json:"average"`
	RequestsLatency   float64 `json:"requests_latency"`
	TotalDataTransfer int     `json:"total_data_transfer"`
	TotalRequests     int     `json:"total_requests"`
	Concurrency       int     `json:"concurrency"`
	// Adicione outros campos conforme necessário com suas tags json

	// Campos adicionados com base na saída detalhada do hey
	Summary struct {
		Total          float64 `json:"total"`
		Slowest        float64 `json:"slowest"`
		Fastest        float64 `json:"fastest"`
		Average        float64 `json:"average"`
		RequestsPerSec float64 `json:"requests_per_sec"`
		TotalData      int     `json:"total_data"`
		SizePerRequest int     `json:"size_per_request"`
	} `json:"summary"`
	Latency struct {
		Distribution []struct {
			Percentage float64 `json:"percentage"`
			Latency    float64 `json:"latency"`
		} `json:"distribution"`
		Histogram []struct {
			Mark    float64 `json:"mark"`
			Count   int     `json:"count"`
			Percent float64 `json:"percent"`
		} `json:"histogram"`
		Details struct {
			DNSDialup float64 `json:"dns_dialup"`
			DNSLookup float64 `json:"dns_lookup"`
			ReqWrite  float64 `json:"req_write"`
			RespWait  float64 `json:"resp_wait"`
			RespRead  float64 `json:"resp_read"`
		} `json:"details"`
	} `json:"latency"`
	StatusCodeCount map[string]int `json:"status_code_count"`
}

// RunResult armazena os resultados de uma única execução do hey
type RunResult struct {
	HeyOutput *HeyResult
	HeyStdout string
	HeyStderr string
	StartTime time.Time
	EndTime   time.Time
	Error     error
}
