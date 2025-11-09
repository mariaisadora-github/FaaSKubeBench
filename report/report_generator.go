package report

import (
	"fmt"
	"os"
	"time"

	"github.com/mariaisadora-github/FaaSKubeBench/metrics"
)

// ReportGenerator é responsável por gerar o relatório final em Markdown
type ReportGenerator struct {
	Metrics metrics.ConsolidatedMetrics
}

// NewReportGenerator cria uma nova instância do ReportGenerator
func NewReportGenerator(metrics metrics.ConsolidatedMetrics) *ReportGenerator {
	return &ReportGenerator{
		Metrics: metrics,
	}
}

// Generate gera o relatório em Markdown no caminho especificado
func (r *ReportGenerator) Generate(filePath string) error {
	content := r.generateMarkdown()

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

func (r *ReportGenerator) generateMarkdown() string {
	m := r.Metrics

	// Formatação de tempo
	formatDuration := func(d time.Duration) string {
		if d == 0 {
			return "N/A"
		}
		return d.String()
	}

	// Formatação de porcentagem
	formatPercent := func(f float64) string {
		return fmt.Sprintf("%.2f%%", f*100)
	}

	// Formatação de bytes
	formatBytes := func(b float64) string {
		const unit = 1024
		if b < unit {
			return fmt.Sprintf("%.2f B", b)
		}
		div, exp := float64(unit), 0
		for n := b / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
		return fmt.Sprintf("%.2f %cB", b/div, "KMGTPE"[exp])
	}

	// Formatação de milicores
	formatMillicores := func(mc float64) string {
		return fmt.Sprintf("%.2f mCores", mc)
	}

	markdown := "# Relatório de Benchmark FaaSKubeBench\n\n"
	markdown += "## 1. Métricas de Desempenho (Hey)\n\n"
	markdown += "| Métrica | Valor |\n"
	markdown += "| :--- | :--- |\n"
	markdown += fmt.Sprintf("| Requisições por Segundo (RPS) | %.2f |\n", m.RPS)
	markdown += fmt.Sprintf("| Latência Média | %.4f s |\n", m.AvgLatency)
	markdown += fmt.Sprintf("| Latência de Cauda (p99) | %.4f s |\n", m.P99Latency)
	markdown += fmt.Sprintf("| Total de Requisições | %d |\n", m.TotalRequests)
	markdown += fmt.Sprintf("| Taxa de Erros HTTP (4xx/5xx) | %s |\n", formatPercent(m.ErrorRate))
	markdown += fmt.Sprintf("| Tráfego de Dados Total | %s |\n", formatBytes(float64(m.TotalData)))
	markdown += fmt.Sprintf("| Tempo de Inicialização | %s |\n", formatDuration(m.TimeInicialization))

	markdown += "\n## 2. Métricas de Orquestração (Kubernetes Exporter)\n\n"
	markdown += "| Métrica | Valor |\n"
	markdown += "| :--- | :--- |\n"
	markdown += fmt.Sprintf("| Pods Escalados (Diferença) | %d |\n", m.ScaledPodsDiff)
	markdown += fmt.Sprintf("| Consumo de CPU (Cluster Total) | %s |\n", formatMillicores(m.ClusterCPUUsage))
	markdown += fmt.Sprintf("| Uso de Memória (Cluster Total) | %s |\n", formatBytes(m.ClusterMemUsage))

	// Nota sobre Warm Start e Tráfego de Rede
	markdown += "\n## 4. Notas Adicionais\n\n"
	markdown += "A métrica de **Tempo de Inicialização** reportada acima é o **Cold Start ou Warm Start** (tempo até o container estar `running` após o início do benchmark).\n\n"

	return markdown
}
