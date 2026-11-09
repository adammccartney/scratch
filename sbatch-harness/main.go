package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

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

func main() {
	sub := NewSbatchSubmitter(nil)
	payload := []byte("#!/usr/bin/env bash\necho build\n")
	jobID, err := sub.Submit(payload)
	if err != nil {
		log.Fatalf("Submit: %v", err)
	}
}
