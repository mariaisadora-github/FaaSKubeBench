package parameters

import (
	"fmt"
)

// ToHeyArgs converte os parâmetros para argumentos do hey
func (p *BenchmarkParameters) ToHeyArgs() []string {
	args := []string{}

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

	if p.Hey.Method != "" {
		args = append(args, "-m", p.Hey.Method)
	}

	if p.Hey.Timeout > 0 {
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

	// Output format - Prefer JSON for structured parsing
	switch p.Hey.Output {
	case "json", "":
		args = append(args, "json")
	case "csv":
		args = append(args, "-o", "csv")
	default:
	}

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
