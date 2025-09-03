package parameters

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ValidateParams valida todos os parâmetros do benchmark
func ValidateParameters(parameters *BenchmarkParameters) error {
	// Validações básicas obrigatórias
	if err := validateRequiredParameters(parameters); err != nil {
		return err
	}

	// Validar plataforma
	if err := validatePlatform(parameters.Platform); err != nil {
		return err
	}

	// Validar tipo de workload
	if err := validateWorkload(parameters.Workload); err != nil {
		return err
	}

	// Validar URL
	if err := validateURL(parameters.URL); err != nil {
		return err
	}

	// Validar parâmetros de tempo/duração
	if err := validateTimeParameters(parameters); err != nil {
		return err
	}

	// Validar parâmetros do hey
	if err := validateHeyParameters(&parameters.Hey); err != nil {
		return err
	}

	return nil
}

// validateRequiredParams valida os parâmetros obrigatórios
func validateRequiredParameters(parameters *BenchmarkParameters) error {
	if parameters.Requests <= 0 {
		return fmt.Errorf("requests must be greater than 0")
	}

	if parameters.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0")
	}

	if parameters.Concurrency > parameters.Requests {
		return fmt.Errorf("concurrency cannot be greater than total requests")
	}

	if parameters.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if parameters.Platform == "" {
		return fmt.Errorf("platform is required")
	}

	if parameters.Function == "" {
		return fmt.Errorf("function is required")
	}

	if parameters.Workload == "" {
		return fmt.Errorf("workload type is required")
	}

	return nil
}

// validatePlatform valida a plataforma serverless
func validatePlatform(platform string) error {
	validPlatforms := map[string]bool{
		"knative": true, "openwhisk": true, "openfaas": true,
	}

	if !validPlatforms[platform] {
		return fmt.Errorf("unsupported platform: %s. Supported platforms: knative, openwhisk, openfaas", platform)
	}

	return nil
}

// validateWorkload valida o tipo de workload
func validateWorkload(workload string) error {
	validWorkloads := map[string]bool{
		"cpu": true, "memory": true, "io": true, "mixed": true,
		"cpu-intensive": true, "memory-intensive": true, "network-intensive": true, "machine-learning": true,
	}

	if !validWorkloads[workload] {
		return fmt.Errorf("unsupported workload: %s. Supported workloads: cpu, memory, io, mixed, cpu-intensive, memory-intensive, network-intensive, machine-learning", workload)
	}

	return nil
}

// validateURL valida a URL do endpoint
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Verificar se começa com http:// ou https://
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	// Tentar fazer parsing da URL
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	return nil
}

// validateTimeParameters valida parâmetros relacionados a tempo
func validateTimeParameters(parameters *BenchmarkParameters) error {
	// Validar formato de tempo se especificado
	if parameters.Time != "" {
		matched, _ := regexp.MatchString(`^\d+[smh]$`, parameters.Time)
		if !matched {
			return fmt.Errorf("invalid time format: %s. Use format like s, m, h", parameters.Time)
		}
	}

	// Validar timeout do hey
	if parameters.Hey.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	// Validar que o timeout não é excessivamente longo
	if parameters.Hey.Timeout > 3600 { // 1 hora
		return fmt.Errorf("timeout cannot exceed 3600 seconds (1 hour)")
	}

	return nil
}

// validateHeyParams valida os parâmetros específicos do hey
func validateHeyParameters(heyParameters *HeyParameters) error {
	// Validar método HTTP
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true,
		"DELETE": true, "HEAD": true, "OPTIONS": true,
		"PATCH": true, "TRACE": true, "CONNECT": true,
	}

	method := strings.ToUpper(heyParameters.Method)
	if !validMethods[method] {
		return fmt.Errorf("invalid HTTP method: %s. Supported methods: GET, POST, PUT, DELETE, HEAD, OPTIONS, PATCH, TRACE, CONNECT", heyParameters.Method)
	}

	// Validar rate limit
	if heyParameters.RateLimit < 0 {
		return fmt.Errorf("rate limit cannot be negative")
	}

	// Validar CPUs
	if heyParameters.CPUs < 0 {
		return fmt.Errorf("CPUs cannot be negative")
	}

	// Validar que body e bodyFile não são usados simultaneamente
	if heyParameters.Body != "" && heyParameters.BodyFile != "" {
		return fmt.Errorf("cannot specify both body and bodyFile parameters")
	}

	// Validar formato de output
	if heyParameters.Output != "" && heyParameters.Output != "csv" {
		return fmt.Errorf("invalid output format: %s. Only 'csv' is supported", heyParameters.Output)
	}

	// Validar conteúdo dos headers
	for key, value := range heyParameters.Headers {
		if key == "" {
			return fmt.Errorf("header key cannot be empty")
		}
		if value == "" {
			return fmt.Errorf("header value for %s cannot be empty", key)
		}
	}

	return nil
}

// ValidateDuration valida uma string de duração (função auxiliar para uso externo)
func ValidateDuration(duration string) error {
	if duration == "" {
		return nil
	}

	_, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration format: %s. Use format like s, m, h", duration)
	}

	return nil
}
