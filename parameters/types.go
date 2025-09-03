package parameters

// Parâmetros para o FaaSKubeBench
type BenchmarkParameters struct {
	// Parâmetros principais da ferramenta
	Requests    int    `yaml:"requests"`
	Concurrency int    `yaml:"concurrency"`
	Time        string `yaml:"time,omitempty"`
	Execution   int    `yaml:"execution,omitempty"`
	Platform    string `yaml:"platform"`
	Function    string `yaml:"function"`
	URL         string `yaml:"url"`
	Workload    string `yaml:"workload"`

	// Parâmetros específicos do Hey - Gerador de Carga
	Hey HeyParameters `yaml:"hey,omitempty"`

	Metadata map[string]string `yaml:"metadata,omitempty"`
}

// HeyParameters agrupa os parâmetros do hey
type HeyParameters struct {
	RateLimit          int               `yaml:"rate_limit,omitempty"`
	Output             string            `yaml:"output,omitempty"`
	Method             string            `yaml:"method,omitempty"`
	Timeout            int               `yaml:"timeout,omitempty"`
	Body               string            `yaml:"body,omitempty"`
	BodyFile           string            `yaml:"body_file,omitempty"`
	ContentType        string            `yaml:"content_type,omitempty"`
	Auth               string            `yaml:"auth,omitempty"`
	Proxy              string            `yaml:"proxy,omitempty"`
	HTTP2              bool              `yaml:"http2,omitempty"`
	Host               string            `yaml:"host,omitempty"`
	DisableCompression bool              `yaml:"disable_compression,omitempty"`
	DisableKeepAlive   bool              `yaml:"disable_keepalive,omitempty"`
	DisableRedirects   bool              `yaml:"disable_redirects,omitempty"`
	CPUs               int               `yaml:"cpus,omitempty"`
	Headers            map[string]string `yaml:"headers,omitempty"`
}
