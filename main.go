package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/knight42/kt/pkg/completion"
	"github.com/knight42/kt/pkg/log"
	"github.com/knight42/kt/pkg/version"
)

// Usage:
// kt -lapp=qwe
// kt deploy app
// kt 'app\w+'

func checkError(err error) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	var (
		shell        string
		printVersion bool
	)

	o := Options{}
	f := genericclioptions.NewConfigFlags(true)
	cmd := &cobra.Command{
		Use: "kt (NAME_REGEXP | TYPE NAME) [-c CONTAINER] [options]",
		Run: func(cmd *cobra.Command, args []string) {
			if printVersion {
				checkError(version.Run())
				return
			}
			if len(shell) > 0 {
				checkError(completion.Generate(cmd, shell))
				return
			}
			checkError(o.Complete(f, args))
			checkError(o.Run(cmd))
		},
		DisableFlagsInUseLine: true,
	}
	flags := cmd.Flags()
	flags.StringVar(f.KubeConfig, "kubeconfig", *f.KubeConfig, "Path to the kubeconfig file to use for CLI requests.")
	flags.StringVar(f.ClusterName, "cluster", *f.ClusterName, "The name of the kubeconfig cluster to use")
	flags.StringVarP(f.Namespace, "namespace", "n", *f.Namespace, "If present, the namespace scope for this CLI request")
	flags.StringVar(f.Context, "context", *f.Context, "The name of the kubeconfig context to use")
	flags.StringVar(f.AuthInfoName, "user", *f.AuthInfoName, "The name of the kubeconfig user to use")
	flags.StringVarP(f.APIServer, "server", "s", *f.APIServer, "The address and port of the Kubernetes API server")

	flags.StringVar(&shell, "completion", "", "Print completion script. One of: bash|zsh.")
	flags.BoolVarP(&printVersion, "version", "V", false, "Print version and exit.")

	flags.StringVarP(&o.selector, "selector", "l", o.selector, "Selector (label query) to filter on pods.")
	flags.StringVarP(&o.container, "container", "c", o.container, "Print the logs of this container")
	flags.Int64Var(&o.tail, "tail", o.tail, "Lines of recent log file to display. Defaults to 0 with no selector, showing all log lines otherwise 10, if a selector is provided.")
	flags.BoolVar(&o.timestamps, "timestamps", o.timestamps, "Include timestamps on each line in the log output")
	flags.StringVar(&o.sinceTime, "since-time", o.sinceTime, "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	flags.StringVar(&o.color, "color", "auto", "Colorize the output. One of: auto|always|never")
	flags.DurationVar(&o.sinceSeconds, "since", o.sinceSeconds, "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")

	log.AddFlags(flags)

	_ = cmd.Execute()
}
