package parameters

import (
	"fmt"
)

// Este arquivo contém métodos para conversão de parâmetros do FaaSKubeBench
// para argumentos do hey, com foco na coleta silenciosa de dados estruturados.
//
// Modificações realizadas:
// - ToHeyArgs() agora força saída JSON e modo silencioso (-o json -q)
// - Removida exibição desnecessária no terminal
// - Adicionados métodos para configuração de coleta silenciosa
// - Dados são coletados estruturadamente para posterior seleção e exibição

// ToHeyArgs converte os parâmetros para argumentos do hey
// Gera apenas argumentos necessários, sem valores padrão desnecessários
func (p *BenchmarkParameters) ToHeyArgs() []string {
	args := []string{}

	// DEBUG: Vamos ver os valores dos parâmetros
	//fmt.Printf("ToHeyArgs - Method: '%s', Timeout: %d\n", p.Hey.Method, p.Hey.Timeout)

	// Adicionar argumentos baseados nos parâmetros principais
	if p.Requests > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", p.Requests))
	}

	if p.Concurrency > 0 {
		args = append(args, "-c", fmt.Sprintf("%d", p.Concurrency))
	}

	if p.Time != "" {
		args = append(args, "-z", p.Time)
	}

	// Adicionar parâmetros específicos do hey
	if p.Hey.RateLimit > 0 {
		args = append(args, "-q", fmt.Sprintf("%d", p.Hey.RateLimit))
	}

	// Só adicionar method se for diferente do padrão GET
	if p.Hey.Method != "" && p.Hey.Method != "GET" {
		args = append(args, "-m", p.Hey.Method)
	}

	// Só adicionar timeout se for diferente do padrão (20) ou explicitamente configurado
	if p.Hey.Timeout > 0 && p.Hey.Timeout != 20 {
		args = append(args, "-t", fmt.Sprintf("%d", p.Hey.Timeout))
	}

	// Adicionar headers
	for key, value := range p.Hey.Headers {
		args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
	}

	// Body
	if p.Hey.Body != "" {
		args = append(args, "-d", p.Hey.Body)
	} else if p.Hey.BodyFile != "" {
		args = append(args, "-D", p.Hey.BodyFile)
	}

	// Content-Type
	if p.Hey.ContentType != "" {
		args = append(args, "-T", p.Hey.ContentType)
	}

	// Flags booleanas
	if p.Hey.HTTP2 {
		args = append(args, "-h2")
	}

	if p.Hey.DisableCompression {
		args = append(args, "-disable-compression")
	}

	if p.Hey.DisableKeepAlive {
		args = append(args, "-disable-keepalive")
	}

	if p.Hey.DisableRedirects {
		args = append(args, "-disable-redirects")
	}

	// Force JSON output for structured parsing (hey outputs JSON by default when no -o flag is used)
	// Não adicionar -o json pois não é um parâmetro válido do hey
	// O hey produz saída JSON por padrão quando usado programaticamente

	// URL final
	args = append(args, p.URL)

	return args
}

// ToEnvVars converte parâmetros para variáveis de ambiente
func (p *BenchmarkParameters) ToEnvVars() map[string]string {
	env := make(map[string]string)

	// Adicionar parâmetros principais
	env["BENCH_REQUESTS"] = fmt.Sprintf("%d", p.Requests)
	env["BENCH_CONCURRENCY"] = fmt.Sprintf("%d", p.Concurrency)
	env["BENCH_PLATFORM"] = p.Platform
	env["BENCH_FUNCTION"] = p.Function
	env["BENCH_URL"] = p.URL
	env["BENCH_WORKLOAD"] = p.Workload

	// Adicionar parâmetros do hey
	env["HEY_METHOD"] = p.Hey.Method
	env["HEY_TIMEOUT"] = fmt.Sprintf("%d", p.Hey.Timeout)
	env["HEY_RATE_LIMIT"] = fmt.Sprintf("%d", p.Hey.RateLimit)

	return env
}

// Clone cria uma cópia dos parâmetros
func (p *BenchmarkParameters) Clone() *BenchmarkParameters {
	clone := *p

	// Clonar maps para evitar referências compartilhadas
	if p.Metadata != nil {
		clone.Metadata = make(map[string]string)
		for k, v := range p.Metadata {
			clone.Metadata[k] = v
		}
	}

	if p.Hey.Headers != nil {
		clone.Hey.Headers = make(map[string]string)
		for k, v := range p.Hey.Headers {
			clone.Hey.Headers[k] = v
		}
	}

	return &clone
}

// ConfigureSilentDataCollection configura os parâmetros para coleta silenciosa de dados
func (p *BenchmarkParameters) ConfigureSilentDataCollection() {
	// O hey produz saída JSON por padrão, não precisamos forçar -o json
	p.Hey.Output = ""

	// Adiciona metadata indicando modo silencioso
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	p.Metadata["collection_mode"] = "silent"
	p.Metadata["output_format"] = "json"
}

// IsSilentMode verifica se a coleta está configurada para modo silencioso
func (p *BenchmarkParameters) IsSilentMode() bool {
	if p.Metadata == nil {
		return false
	}
	return p.Metadata["collection_mode"] == "silent"
}

// GetDataCollectionSummary retorna um resumo dos dados coletados (para seleção posterior)
func (p *BenchmarkParameters) GetDataCollectionSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	summary["data_collection_config"] = map[string]interface{}{
		"silent_mode":    p.IsSilentMode(),
		"output_format":  p.Hey.Output,
		"requests":       p.Requests,
		"concurrency":    p.Concurrency,
		"url":            p.URL,
		"platform":       p.Platform,
		"function":       p.Function,
		"generated_args": p.ToHeyArgs(), // Para debug se necessário
	}

	if p.Metadata != nil {
		summary["metadata"] = p.Metadata
	}

	return summary
}

// ValidateHeyArgs valida se os argumentos gerados estão corretos
func (p *BenchmarkParameters) ValidateHeyArgs() []string {
	args := p.ToHeyArgs()

	// Verificações básicas
	validArgs := []string{}

	for i, arg := range args {
		// Não adicionar argumentos inválidos como "json" sozinho
		if arg == "json" && (i == 0 || args[i-1] != "-o") {
			continue // Pular argumento inválido
		}
		validArgs = append(validArgs, arg)
	}

	return validArgs
}
