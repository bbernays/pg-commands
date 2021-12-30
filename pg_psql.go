package pg_commands

import (
	"os"
	"os/exec"
)

type Exec struct {
	*Postgres
	// Verbose mode
	Verbose bool
	// Path: where to save
	Path string

	Query string

	//File: name of the file that pgdump will create
	File string

	IgnoreTableData []string
}

const PGPsqlCmd = "psql"

func NewExec(pg *Postgres) *Exec {
	return &Exec{Postgres: pg}
}

func (x *Exec) Exec(opts ExecOptions) Result {
	result := Result{}
	cmd := exec.Command(PGPsqlCmd, x.dumpOptions()...)
	cmd.Env = append(os.Environ(), x.EnvPassword)
	stderrIn, _ := cmd.StderrPipe()
	go func() {
		result.Output = streamExecOutput(stderrIn, opts)
	}()
	cmd.Start()
	err := cmd.Wait()
	if exitError, ok := err.(*exec.ExitError); ok {
		result.Error = &ResultError{Err: err, ExitCode: exitError.ExitCode(), CmdOutput: result.Output}
	}
	return result
}

func (x *Exec) dumpOptions() (options []string) {
	options = append(options, x.Postgres.Parse()...)

	if x.Query != "" {
		options = append(options, "-c", x.Query)
	}

	return options
}
