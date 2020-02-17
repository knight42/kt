package log

import (
	"log"
	"os"

	"github.com/spf13/pflag"
)

var (
	stderr = log.New(os.Stderr, "", 0)
)

var (
	verbosity int
)

type Verbose bool

func AddFlags(flags *pflag.FlagSet) {
	flags.IntVarP(&verbosity, "verbosity", "v", 0, "number for the log level verbosity")
}

func Errorf(fmt string, v ...interface{}) {
	stderr.Printf(fmt, v...)
}

func V(level int) Verbose {
	return level <= verbosity
}

func (v Verbose) Infof(fmt string, args ...interface{}) {
	if v {
		stderr.Printf(fmt, args...)
	}
}
