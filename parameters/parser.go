package parameters

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3" // Converter YAML em go
)

// Carregar parâmetros de um arquivo YAML
func LoadParametersFromFile(filePath string) (*BenchmarkParameters, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	return LoadParametersFromYAML(data)
}

// Carregar parâmetros de dados YAML
func LoadParametersFromYAML(data []byte) (*BenchmarkParameters, error) {
	var parameters BenchmarkParameters
	err := yaml.Unmarshal(data, &parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	// DEBUG: Valores antes dos defaults
	//fmt.Printf("Antes ApplyDefaults - Method: '%s', Timeout: %d\n", parameters.Hey.Method, parameters.Hey.Timeout)

	// Aplicar defaults
	ApplyDefaults(&parameters)

	// DEBUG: Valores depois dos defaults
	//fmt.Printf("Depois ApplyDefaults - Method: '%s', Timeout: %d\n", parameters.Hey.Method, parameters.Hey.Timeout)

	// Validar parâmetros
	if err := ValidateParameters(&parameters); err != nil {
		return nil, err
	}

	return &parameters, nil
}
