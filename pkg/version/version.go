package version

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
)

type Info struct {
	GitCommit string
	BuildDate string
	Version   string
	GoVersion string
	Compiler  string
	Platform  string
}

var (
	Version   = "0.0.0"
	GitCommit = "$Format:%H$"          // sha1 from git, output of $(git rev-parse HEAD)
	BuildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

func Run() error {
	e := json.NewEncoder(os.Stdout)
	return e.Encode(Info{
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		Version:   Version,

		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	})
}
