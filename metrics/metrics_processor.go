package metrics

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mariaisadora-github/FaaSKubeBench/heyexec"
)

// Estruturas para armazenar as métricas consolidadas
type ConsolidatedMetrics struct {
	// Hey Metrics
	AvgLatency    float64
	P99Latency    float64
	RPS           float64
	ErrorRate     float64
	FailureRate   float64
	TotalRequests int
	TotalData     int

	// Kubernetes/Exporter Metrics
	ScaledPodsDiff     int
	ClusterCPUUsage    float64       // Millicores
	ClusterMemUsage    float64       // Bytes
	TimeInicialization time.Duration // Média dos tempos de inicialização

	// Dados brutos para o cálculo do Cold Start
	PodStartedAt map[string]float64 // podName: started_at_timestamp (Unix seconds)
}

// PostProcessor é responsável por coletar métricas do exporter e consolidar os resultados
type PostProcessor struct {
	exporterURL string
	httpClient  *http.Client
}

// NewPostProcessor cria uma nova instância do PostProcessor
func NewPostProcessor(exporterURL string) *PostProcessor {
	return &PostProcessor{
		exporterURL: exporterURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// CollectMetrics coleta as métricas do endpoint Prometheus do exporter
func (p *PostProcessor) CollectMetrics(ctx context.Context) (ConsolidatedMetrics, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.exporterURL, nil)
	if err != nil {
		return ConsolidatedMetrics{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return ConsolidatedMetrics{}, fmt.Errorf("failed to reach exporter at %s: %w", p.exporterURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ConsolidatedMetrics{}, fmt.Errorf("exporter returned non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ConsolidatedMetrics{}, fmt.Errorf("failed to read exporter response body: %w", err)
	}

	return p.parsePrometheusMetrics(string(body))
}

// parsePrometheusMetrics extrai os valores das métricas desejadas do formato Prometheus
func (p *PostProcessor) parsePrometheusMetrics(metricsBody string) (ConsolidatedMetrics, error) {
	metrics := ConsolidatedMetrics{
		PodStartedAt: make(map[string]float64),
	}

	// Variáveis auxiliares para o cálculo de Cold Start
	//TimeInicialization := []float64{}

	lines := strings.Split(metricsBody, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		metricNameAndLabels := parts[0]
		valueStr := parts[1]

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			// Ignora linhas com valores não numéricos
			continue
		}

		if strings.HasPrefix(metricNameAndLabels, "kubernetes_cluster_cpu_usage_millicores") {
			metrics.ClusterCPUUsage = value
		} else if strings.HasPrefix(metricNameAndLabels, "kubernetes_cluster_memory_usage_bytes") {
			metrics.ClusterMemUsage = value
		} else if strings.HasPrefix(metricNameAndLabels, "serverless_pod_scaled_difference") {
			metrics.ScaledPodsDiff = int(value)
		} else if strings.HasPrefix(metricNameAndLabels, "serverless_pod_container_started_at_seconds") {
			// Exemplo de linha: serverless_pod_container_started_at_seconds{namespace="default",pod="func-xxx",function="myfunc"} 1678886400.123
			// Extrai o nome do pod do label
			start := strings.Index(metricNameAndLabels, "pod=\"")
			end := strings.Index(metricNameAndLabels[start+5:], "\"")
			if start != -1 && end != -1 {
				podName := metricNameAndLabels[start+5 : start+5+end]
				metrics.PodStartedAt[podName] = value
			}
		}
	}

	return metrics, nil
}

// ConsolidateResults combina os resultados do hey com as métricas do exporter e calcula o Cold Start
func (p *PostProcessor) ConsolidateResults(heyResults []*heyexec.RunResult, collectedMetrics ConsolidatedMetrics, benchmarkStartTime time.Time) ConsolidatedMetrics {

	// 1. Consolidar Hey Metrics (assumindo apenas a primeira execução para simplificar)
	if len(heyResults) > 0 && heyResults[0].HeyOutput != nil {
		firstRun := heyResults[0].HeyOutput

		// RPS
		collectedMetrics.RPS = firstRun.RequestsPerSecond

		// Latência Média (AvgLatency)
		collectedMetrics.AvgLatency = firstRun.Summary.Average

		// Latência P99 (procura no LatencyDistribution)
		for _, dist := range firstRun.LatencyDistribution {
			if dist.Percentage >= 0.99 {
				collectedMetrics.P99Latency = dist.Latency
				break
			}
		}

		// Taxa de Erros e Falhas
		collectedMetrics.TotalRequests = firstRun.Requests
		collectedMetrics.TotalData = firstRun.BytesTotal

		// Taxa de Erros (assumindo que o hey não fornece um campo direto, calculamos a partir do status code)
		// Isso é uma estimativa, pois o hey fornece o Status Code Distribution.
		// Assumindo que 2xx são sucesso e 4xx/5xx são falhas.
		errorCount := 0
		for codeStr, count := range firstRun.StatusCodeDist {
			if len(codeStr) == 3 && (codeStr[0] == '4' || codeStr[0] == '5') {
				errorCount += count
			}
		}

		if firstRun.Requests > 0 {
			collectedMetrics.ErrorRate = float64(errorCount) / float64(firstRun.Requests)
		}

		// Falhas de conexão (Hey não expõe diretamente, mas pode ser inferido do ErrorDist)
		// Vamos simplificar e usar a taxa de erros HTTP como principal indicador.
	}

	// 2. Calcular Cold Start Time
	var totalColdStartDuration time.Duration
	var coldStartCount int

	benchmarkStartTimeUnix := float64(benchmarkStartTime.UnixNano()) / float64(time.Second)

	for _, startedAtUnix := range collectedMetrics.PodStartedAt {
		// O startedAtUnix deve ser maior que o benchmarkStartTimeUnix
		if startedAtUnix > benchmarkStartTimeUnix {
			coldStartDuration := startedAtUnix - benchmarkStartTimeUnix

			// Converte o float64 de segundos para time.Duration
			duration := time.Duration(coldStartDuration * float64(time.Second))
			totalColdStartDuration += duration
			coldStartCount++
		}
	}

	if coldStartCount > 0 {
		collectedMetrics.TimeInicialization = totalColdStartDuration / time.Duration(coldStartCount)
	}

	return collectedMetrics
}
