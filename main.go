package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mariaisadora-github/FaaSKubeBench/heyexec"
	"github.com/mariaisadora-github/FaaSKubeBench/metrics"
	"github.com/mariaisadora-github/FaaSKubeBench/parameters"
)

// Constantes para o exporter
const (
	ExporterPort = "8000"
	ExporterURL  = "http://localhost:" + ExporterPort + "/metrics"
)

func main() {
	// 1. Tratamento de Argumentos de Linha de Comando
	if len(os.Args) < 2 {
		log.Fatal("Usage: faaskubebench <path-to-config.yaml>")
	}

	configPath := os.Args[1]

	// 2. Carregar Parâmetros
	params, err := parameters.LoadParametersFromFile(configPath)
	if err != nil {
		log.Fatalf("Erro ao carregar os parâmetros do arquivo %s: %v", configPath, err)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("   INICIANDO BENCHMARK FAASKUBEBENCH")
	fmt.Println(strings.Repeat("=", 80))

	// --- Orquestração do Benchmark ---

	// 3. Iniciar Exporter de Métricas (Docker Compose)
	fmt.Println(" Iniciando Exporter de Métricas...")

	// Define o tempo de início do benchmark ANTES de iniciar o exporter
	benchmarkStartTime := time.Now().UTC()
	benchmarkStartTimeStr := benchmarkStartTime.Format(time.RFC3339Nano)

	os.Setenv("BENCHMARK_START_TIME", benchmarkStartTimeStr)

	// CORREÇÃO: Usar "docker compose" (novo padrão)
	exporterCmd := exec.Command("docker", "compose", "up", "-d")
	// Manter o stdout/stderr para o caso de erro, mas sem logs de sucesso
	exporterCmd.Stdout = os.Stdout
	exporterCmd.Stderr = os.Stderr

	if err := exporterCmd.Run(); err != nil {
		log.Fatalf("Erro ao iniciar o exporter de métricas (Verifique se o Docker e o plugin 'compose' estão instalados e o docker-compose.yaml está na pasta): %v", err)
	}
	time.Sleep(5 * time.Second) // Aguarda estabilização
	fmt.Println(" Exporter iniciado com sucesso\n")

	// Garante que o exporter será parado ao final, mesmo em caso de erro
	defer func() {
		fmt.Println("\n Encerrando exporter de métricas...")
		stopCmd := exec.Command("docker", "compose", "down")
		// Suprimir a saída de sucesso do 'down'
		stopCmd.Stdout = nil
		stopCmd.Stderr = nil
		if err := stopCmd.Run(); err != nil {
			log.Printf("Aviso: Erro ao parar o exporter de métricas: %v", err)
		}
	}()

	// 4. Executar o Benchmark (Hey)
	fmt.Println(" Executando Gerador de Carga...")
	heyExecutor := heyexec.NewHeyExecutor(params)

	// Executa o hey
	allHeyResults, err := heyExecutor.ExecuteMultiple()
	if err != nil {
		log.Fatalf("Erro durante a execução do gerador de carga: %v", err)
	}

	// 5. Coletar Métricas do Exporter
	fmt.Println("\n Coletando Métricas do Prometheus...")

	// Cria o cliente para coletar as métricas do Prometheus do exporter
	postProcessor := metrics.NewPostProcessor(ExporterURL)

	// Coleta as métricas do exporter (started_at, contagem de pods, CPU/Memória do cluster)
	collectedMetrics, err := postProcessor.CollectMetrics(context.Background())
	if err != nil {
		log.Fatalf("Erro ao coletar métricas do exporter: %v", err)
	}

	// 6. Pós-processamento e Consolidação
	fmt.Println(" Processando e consolidando resultados...")

	// Consolida os resultados do hey e as métricas coletadas
	finalReportData := postProcessor.ConsolidateResults(allHeyResults, collectedMetrics, benchmarkStartTime)

	// 7. Exibir Resultados na Tela
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                      RELATÓRIO DE BENCHMARK FAASKUBEBENCH")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n MÉTRICAS DO GERADOR DE CARGA")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("   Requisições por Segundo (RPS):     %.2f req/s\n", finalReportData.RPS)
	fmt.Printf("    Latência Média:                     %.4f s (%.2f ms)\n",
		finalReportData.AvgLatency, finalReportData.AvgLatency*1000)
	fmt.Printf("   Latência de Cauda (p99):            %.4f s (%.2f ms)\n",
		finalReportData.P99Latency, finalReportData.P99Latency*1000)
	fmt.Printf("   Total de Requisições:               %d\n", finalReportData.TotalRequests)
	fmt.Printf("   Taxa de Erros HTTP (4xx/5xx):       %.2f%%\n", finalReportData.ErrorRate*100)
	fmt.Printf("   Tráfego de Dados Total:             %.2f MB\n", float64(finalReportData.TotalData)/(1024*1024))

	fmt.Println("\n  MÉTRICAS DE ORQUESTRAÇÃO")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("   Pods Escalados (Diferença):         %d\n", finalReportData.ScaledPodsDiff)
	fmt.Printf("   Consumo de CPU (Cluster Total):     %.2f mCores\n", finalReportData.ClusterCPUUsage)
	fmt.Printf("   Uso de Memória (Cluster Total):     %.2f MB\n", finalReportData.ClusterMemUsage/(1024*1024))
	if finalReportData.TimeInicialization > 0 {
		fmt.Printf("   Tempo de Inicialização: %s\n", finalReportData.TimeInicialization)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println(" Benchmark concluído com sucesso!")
	fmt.Println(strings.Repeat("=", 80) + "\n")
}
