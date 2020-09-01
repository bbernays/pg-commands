package pg_commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	// PGDumpCmd is the path to the `pg_dump` executable
	PGDumpCmd           = "pg_dump"
	PGDumpStdOpts       = []string{"--no-owner", "--no-acl", "--clean", "--blob"}
	PGDumpDefaultFormat = "c"
)

// Dump is an `Exporter` interface that backs up a Postgres database via the `pg_dump` command.
type Dump struct {
	*Postgres
	// Verbose mode
	Verbose bool
	// Path: setup path dump out
	Path string
	// Format: output file format (custom, directory, tar, plain text (default))
	Format *string
	// Extra pg_dump x.FullOptions
	// e.g []string{"--inserts"}
	Options []string

	IgnoreTables []string
}

func NewDump(pg *Postgres) *Dump {
	return &Dump{Options: PGDumpStdOpts, Postgres: pg}
}

// Exec `pg_dump` of the specified database, and creates a gzip compressed tarball archive.
func (x *Dump) Exec() Result {
	result := Result{Mine: "application/x-tar"}
	result.File = x.newFileName()
	options := append(x.dumpOptions(), fmt.Sprintf(`-f%s%v`, x.Path,result.File))
	result.FullCommand = strings.Join(options, " ")
	cmd := exec.Command(PGDumpCmd, options...)
	cmd.Env = append(os.Environ(), x.EnvPassword)
	out, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = &ResultError{Err: err, CmdOutput: string(out)}
		if exitError, ok := err.(*exec.ExitError); ok {
			result.Error.ExitCode = exitError.ExitCode()
		}
	}
	result.Output = string(out)
	return result
}

func (x *Dump) ResetOptions() {
	x.Options = []string{}
}

func (x *Dump) EnableVerbose() {
	x.Verbose = true
}

func (x *Dump) SetupFormat(f string) {
	x.Format = &f
}

func (x *Dump) SetPath(path string) {
	x.Path = path
}

func (x *Dump) newFileName() string {
	return fmt.Sprintf(`%v_%v.sql.tar.gz`, x.DB, time.Now().Unix())
}

func (x *Dump) dumpOptions() []string {
	options := x.Options
	options = append(options, x.Postgres.Parse()...)

	if x.Format != nil {
		options = append(options, fmt.Sprintf(`-F%v`, *x.Format))
	} else {
		options = append(options, fmt.Sprintf(`-F%v`, PGDumpDefaultFormat))
	}
	if x.Verbose {
		options = append(options, "-v")
	}
	if len(x.IgnoreTables) > 0 {
		options = append(options, x.IgnoreTableToString())
	}
	return options
}
func (x *Dump) IgnoreTableToString() string {
	var t string
	for i, tables := range x.IgnoreTables {
		if i == len(x.IgnoreTables)-1 {
			t += "--exclude-table=" + tables
		} else {
			t += "--exclude-table=" + tables + " "
		}
	}
	return t
}
