package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const scriptText = `#!/usr/bin/env bash
#ASC --vanilla
#SBATCH -p zen4_0768
#SBATCH -N 1
#SBATCH --exclusive

set -euo pipefail

## Build

module --force purge
module load EESSI/{{ .Env.SwsVariant }}
module load buildenv/default-foss-{{ .Env.BuildEnv }}

mkdir -p build/{{ .Env.BuildEnv }}
gcc -o build/{{ .Env.BuildEnv }}/time_dgemm time_dgemm.c -lopenblas


## Logfile

BUILD=./build/{{ .Env.BuildEnv }}
LOGFILE=${BUILD}/time_dgemm.out
touch ${LOGFILE}

echo -e "Openblas test report: EESSI {{ .Env.SwsVariant }}; foss-{{ .Env.BuildEnv }} \n-----------\n" | tee -a ${LOGFILE}

## Run


{{range $val := .Env.OmpVars }}
export OMP_NUM_THREADS={{ $val.OmpNumThreads }}
export OMP_PLACES={{ $val.OmpPlaces }}
export OMP_PROC_BIND={{ $val.OmpProcBind }}

OUTFILE=${BUILD}/time_dgemm_T{{ $val.OmpNumThreads }}
${BUILD}/time_dgemm -f ${OUTFILE} 20000 20000 20000 | tee -a ${LOGFILE}
echo -e "\n\n" | tee -a ${LOGFILE}
{{ end }}
`

var sbatchTemplate = template.Must(
	template.New("sbatch").Parse(scriptText),
)

type OmpVars struct {
	OmpNumThreads string
	OmpPlaces     string
	OmpProcBind   string
}

func NewOmpVars(nThreads, places string) *OmpVars {
	return &OmpVars{
		OmpNumThreads: nThreads,
		OmpPlaces:     places,
		OmpProcBind:   "true",
	}
}

var runEnv = []*OmpVars{
	NewOmpVars("4", "{0:1},{1:1},{2:1},{3:1}"),
	NewOmpVars("8", "{0:1},{1:1},{2:1},{3:1},{4:1},{5:1},{6:1},{7:1}"),
	NewOmpVars("16", "{0:1},{1:1},{2:1},{3:1},{4:1},{5:1},{6:1},{7:1},{24:1},{25:1},{26:1},{27:1},{28:1},{29:1},{30:1},{31:1}"),
	NewOmpVars("32", "{0:8:1},{24:8:1},{48:8:1},{72:8:1}"),
}

type BuildEnvTest struct {
	SwsVariant string
	BuildEnv   string
	OmpVars    []*OmpVars
}

type ScriptData struct {
	Env BuildEnvTest
}

type CmdConfig struct {
	Args    []string
	Timeout time.Duration
	Stdout  io.Writer
	Stderr  io.Writer
	Stdin   io.Reader
	Env     []string
}

// NewCmdConfig creates a bare bones CmdConfig with some sane defaults
func NewCmdConfig(args []string) *CmdConfig {
	cfg := &CmdConfig{
		Args:    args,
		Timeout: 3 * time.Second,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Stdin:   os.Stdin,
	}
	return cfg
}

// Run executes the configured command.
func (c *CmdConfig) Run() error {
	if len(c.Args) == 0 {
		return fmt.Errorf("subproc.Run: empty Args")
	}
	ctx := context.Background()
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, c.Args[0], c.Args[1:]...)
	cmd.Env = c.Env
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	cmd.Stdin = c.Stdin

	return cmd.Run()
}

type SbatchSubmitter struct {
	Cfg *CmdConfig
}

// NewSbatchSubmitter creates a Submitter that pipes scripts to
// `sbatch --parsable <flags...>` with a 30s timeout.
func NewSbatchSubmitter(flags []string) *SbatchSubmitter {
	prg := []string{"sbatch", "--parsable"}
	prg = append(prg, flags...)
	cfg := NewCmdConfig(prg)
	cfg.Timeout = time.Minute * 3
	return &SbatchSubmitter{Cfg: cfg}
}

// Submit returns the job ID reported by sbatch --parsable.
func (s *SbatchSubmitter) Submit(script []byte) (string, error) {

	var stdout, stderr bytes.Buffer
	s.Cfg.Stdin = bytes.NewReader(script)
	s.Cfg.Stdout = &stdout
	s.Cfg.Stderr = &stderr

	if err := s.Cfg.Run(); err != nil {
		return "", fmt.Errorf("sbatch --parsable: %w: %s",
			err, strings.TrimSpace(stderr.String()))
	}

	jobID := strings.TrimSpace(stdout.String())
	if jobID == "" {
		return "", fmt.Errorf("sbatch --parsable returned no job id; stderr: %s",
			strings.TrimSpace(stderr.String()))
	}
	return jobID, nil
}

func RenderScript(data ScriptData, w io.Writer) error {
	if err := sbatchTemplate.Execute(w, data); err != nil {
		return fmt.Errorf("execute sbatch template: %w", err)
	}
	return nil
}

func main() {
	sub := NewSbatchSubmitter(nil)
	sub.Cfg.Env = os.Environ()

	var tests = []BuildEnvTest{
		BuildEnvTest{
			SwsVariant: "2023.06",
			BuildEnv:   "2023b",
			OmpVars:    runEnv,
		},
		BuildEnvTest{
			SwsVariant: "2025.06",
			BuildEnv:   "2024a",
			OmpVars:    runEnv,
		},
		BuildEnvTest{
			SwsVariant: "2025.06",
			BuildEnv:   "2025a",
			OmpVars:    runEnv,
		},
		BuildEnvTest{
			SwsVariant: "2025.06",
			BuildEnv:   "2025b",
			OmpVars:    runEnv,
		},
		BuildEnvTest{
			SwsVariant: "2026.06",
			BuildEnv:   "2026.1",
			OmpVars:    runEnv,
		},
	}

	for _, t := range tests {
		var script bytes.Buffer
		data := ScriptData{
			Env: t,
		}

		if err := RenderScript(data, &script); err != nil {
			log.Fatalf("RenderScript: %v", err)
		}

		jobID, err := sub.Submit(script.Bytes())
		if err != nil {
			log.Fatalf("Submit: %v", err)
		}

		fmt.Printf("submitted job with id: %s\n", jobID)
	}
}
