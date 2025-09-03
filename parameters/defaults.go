package parameters

// DefaultParams retorna os valores padrão para todos os parâmetros
func DefaultParameters() *BenchmarkParameters {
	return &BenchmarkParameters{
		Requests:    200,
		Concurrency: 50,
		Execution:   1,
		Platform:    "knative",
		Workload:    "cpu",
		Hey: HeyParameters{
			Method:  "GET",
			Timeout: 20,
			CPUs:    1,
		},
	}
}

// ApplyDefaults preenche valores ausentes com defaults
func ApplyDefaults(parameters *BenchmarkParameters) {
	defaults := DefaultParameters()

	if parameters.Requests == 0 {
		parameters.Requests = defaults.Requests
	}

	if parameters.Concurrency == 0 {
		parameters.Concurrency = defaults.Concurrency
	}

	if parameters.Execution == 0 {
		parameters.Execution = defaults.Execution
	}

	if parameters.Platform == "" {
		parameters.Platform = defaults.Platform
	}

	if parameters.Workload == "" {
		parameters.Workload = defaults.Workload
	}

	if parameters.Hey.Method == "" {
		parameters.Hey.Method = defaults.Hey.Method
	}

	if parameters.Hey.Timeout == 0 {
		parameters.Hey.Timeout = defaults.Hey.Timeout
	}

	if parameters.Hey.CPUs == 0 {
		parameters.Hey.CPUs = defaults.Hey.CPUs
	}
}
