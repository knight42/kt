package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/knight42/kt/pkg/completion"
	"github.com/knight42/kt/pkg/log"
	"github.com/knight42/kt/pkg/version"
)

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
		Example: ` # Begin streaming the logs from all containers in pods belong to Service foo
 kt svc foo

 # Begin streaming the logs from all containers in pods belong to Deployment bar in the last hour
 kt --since=1h deploy bar

 # Begin streaming the logs from container nginx-1 in pods belong to Deployment nginx
 kt -c nginx-1 deployment nginx

 # Filter pods by name or regexp
 kt 'foo'
 kt 'foo-\w+'`,
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

	flags.StringVarP(&o.selector, "selector", "l", o.selector, "Selector (label query) to filter on pods, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	flags.StringVarP(&o.container, "container", "c", o.container, "Regular expression to match container names.")
	flags.Int64Var(&o.tail, "tail", 10, "Lines of recent log file to display. Defaults to 10. If set to 0 it will return all logs.")
	flags.BoolVar(&o.timestamps, "timestamps", o.timestamps, "Include timestamps on each line in the log output")
	flags.BoolVar(&o.previous, "previous", false, "If true, print the logs for the previous instance of the container in a pod if it exists.")
	flags.BoolVar(&o.exitWithPods, "exit-with-pods", false, "Exit if all watched pods are deleted.")
	flags.BoolVar(&o.noPrefix, "no-prefix", false, "Display original log without prefix.")
	flags.StringVar(&o.sinceTime, "since-time", o.sinceTime, "Only return logs after a specific date (RFC3339). Only one of since-time / since may be used.")
	flags.StringVar(&o.color, "color", "auto", "Colorize the output. One of: auto|always|never")
	flags.DurationVar(&o.sinceSeconds, "since", o.sinceSeconds, "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	flags.StringVar(&o.nodeName, "node-name", "", "The name of the node that pods running on")

	log.AddFlags(flags)

	_ = cmd.Execute()
}
