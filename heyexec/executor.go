package heyexec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/mariaisadora-github/FaaSKubeBench/parameters"
)

// Executar o teste de carga usando o hey
type HeyExecutor struct {
	Parameters *parameters.BenchmarkParameters
}

// Criar uma nova instância do executor do hey
func NewHeyExecutor(parameters *parameters.BenchmarkParameters) *HeyExecutor {
	return &HeyExecutor{
		Parameters: parameters,
	}
}

// Execute executa o comando hey com os parâmetros configurados
func (e *HeyExecutor) Execute() (*RunResult, error) {
	args := e.Parameters.ToHeyArgs()

	startTime := time.Now()

	cmd := exec.Command("hey", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()

	endTime := time.Now()

	rawStdout := stdoutBuf.String()
	rawStderr := stderrBuf.String()

	result := &RunResult{
		HeyStdout: rawStdout,
		HeyStderr: rawStderr,
		StartTime: startTime,
		EndTime:   endTime,
		Error:     runErr,
	}

	if runErr != nil {
		return result, fmt.Errorf("erro ao executar hey: %v\nSaída de erro: %s", runErr, rawStderr)
	}

	// Tentar fazer o unmarshalling da saída JSON do hey
	var heyOutput HeyResult

	// Solução de último recurso: Procurar o início do JSON (o primeiro '{')
	// e descartar qualquer texto de cabeçalho (Summary:, etc.) que o hey insiste em imprimir.
	jsonStart := strings.Index(rawStdout, "{")
	if jsonStart == -1 {
		return result, fmt.Errorf("erro ao fazer parse do JSON de saída do hey: não foi encontrado o início do objeto JSON ('{').\nSaída bruta: %s", rawStdout)
	}

	jsonBytes := []byte(rawStdout[jsonStart:])
	err := json.Unmarshal(jsonBytes, &heyOutput)

	if err != nil {
		return result, fmt.Errorf("erro ao fazer parse do JSON de saída do hey: %v\nSaída bruta: %s", err, rawStdout)
	}

	result.HeyOutput = &heyOutput
	return result, nil
}

func (e *HeyExecutor) ExecuteMultiple() ([]*RunResult, error) {
	allResults := []*RunResult{}
	for i := 0; i < e.Parameters.Execution; i++ {
		fmt.Printf("  Execução %d/%d em andamento...\n", i+1, e.Parameters.Execution)

		runResult, err := e.Execute()
		allResults = append(allResults, runResult)

		if err != nil {
			fmt.Printf("    Erro na execução %d: %v\n", i+1, err)
			// Decide se quer parar ou continuar em caso de erro
			// return allResults, err
		} else {
			fmt.Printf("   Execução %d/%d concluída\n", i+1, e.Parameters.Execution)
		}
	}

	return allResults, nil
}
