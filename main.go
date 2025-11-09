package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/mariaisadora-github/FaaSKubeBench/heyexec"
	"github.com/mariaisadora-github/FaaSKubeBench/metrics"
	"github.com/mariaisadora-github/FaaSKubeBench/parameters"
	"github.com/mariaisadora-github/FaaSKubeBench/report"
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
		log.Fatalf("Failed to load parameters from %s: %v", configPath, err)
	}

	fmt.Printf("Configuration loaded successfully! %+v\n", params)

	// --- Orquestração do Benchmark ---

	// 3. Iniciar Exporter de Métricas (Docker Compose)
	fmt.Println("\n--- Iniciando Exporter de Métricas ---")

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

	// Garante que o exporter será parado ao final, mesmo em caso de erro
	defer func() {
		stopCmd := exec.Command("docker", "compose", "down")
		// Suprimir a saída de sucesso do 'down'
		stopCmd.Stdout = nil
		stopCmd.Stderr = nil
		if err := stopCmd.Run(); err != nil {
			log.Printf("Aviso: Erro ao parar o exporter de métricas: %v", err)
		}
	}()

	// 4. Executar o Benchmark (Hey)
	fmt.Println("\n--- Executando o Gerador de Carga ---")
	heyExecutor := heyexec.NewHeyExecutor(params)

	// Executa o hey
	allHeyResults, err := heyExecutor.ExecuteMultiple()
	if err != nil {
		log.Fatalf("Erro durante a execução do gerador de carga: %v", err)
	}

	// 5. Coletar Métricas do Exporter
	fmt.Println("\n--- Coletando Métricas do Prometheus ---")

	// Cria o cliente para coletar as métricas do Prometheus do exporter
	postProcessor := metrics.NewPostProcessor(ExporterURL)

	// Coleta as métricas do exporter (started_at, contagem de pods, CPU/Memória do cluster)
	collectedMetrics, err := postProcessor.CollectMetrics(context.Background())
	if err != nil {
		log.Fatalf("Erro ao coletar métricas do exporter: %v", err)
	}

	// 6. Pós-processamento e Consolidação
	fmt.Println("\n--- Pós-processamento e Consolidação ---")

	// Consolida os resultados do hey e as métricas coletadas
	finalReportData := postProcessor.ConsolidateResults(allHeyResults, collectedMetrics, benchmarkStartTime)

	// 7. Gerar Relatório
	fmt.Println("\n--- Relatório Final ---")
	reportGenerator := report.NewReportGenerator(finalReportData)

	reportPath := "benchmark_report.md"
	if err := reportGenerator.Generate(reportPath); err != nil {
		log.Fatalf("Erro ao gerar relatório: %v", err)
	}
	fmt.Printf("Relatório gerado com sucesso em: %s\n", reportPath)
}
