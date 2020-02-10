package completion

const (
	bashCompletionFunc = `# call kt $1,
__kt_debug_out()
{
    local cmd="$1"
    __kt_debug "${FUNCNAME[1]}: get completion by ${cmd}"
    eval "${cmd} 2>/dev/null"
}
__kt_override_flag_list=(--kubeconfig --cluster --user --context --namespace --server -n -s)
__kt_override_flags()
{
    local ${__kt_override_flag_list[*]##*-} two_word_of of var
    for w in "${words[@]}"; do
        if [ -n "${two_word_of}" ]; then
            eval "${two_word_of##*-}=\"${two_word_of}=\${w}\""
            two_word_of=
            continue
        fi
        for of in "${__kt_override_flag_list[@]}"; do
            case "${w}" in
                ${of}=*)
                    eval "${of##*-}=\"${w}\""
                    ;;
                ${of})
                    two_word_of="${of}"
                    ;;
            esac
        done
    done
    for var in "${__kt_override_flag_list[@]##*-}"; do
        if eval "test -n \"\$${var}\""; then
            eval "echo -n \${${var}}' '"
        fi
    done
}
__kt_config_get_contexts()
{
    __kt_parse_config "contexts"
}
__kt_config_get_clusters()
{
    __kt_parse_config "clusters"
}
__kt_config_get_users()
{
    __kt_parse_config "users"
}
# $1 has to be "contexts", "clusters" or "users"
__kt_parse_config()
{
    local template kt_out
    template="{{ range .$1  }}{{ .name }} {{ end }}"
    if kt_out=$(__kt_debug_out "kubectl config $(__kt_override_flags) -o template --template=\"${template}\" view"); then
        COMPREPLY=( $( compgen -W "${kt_out[*]}" -- "$cur" ) )
    fi
}
# $1 is the name of resource (required)
# $2 is template string for kubectl get (optional)
__kt_parse_get()
{
    local template
    template="${2:-"{{ range .items  }}{{ .metadata.name }} {{ end }}"}"
    local kt_out
    if kt_out=$(__kt_debug_out "kubectl get $(__kt_override_flags) -o template --template=\"${template}\" \"$1\""); then
        COMPREPLY+=( $( compgen -W "${kt_out[*]}" -- "$cur" ) )
    fi
}
__kt_get_resource()
{
    if [[ ${#nouns[@]} -eq 0 ]]; then
      local kt_out
      if kt_out=$(__kt_debug_out "kubectl api-resources $(__kt_override_flags) -o name --cached --request-timeout=5s --verbs=get"); then
          COMPREPLY=( $( compgen -W "${kt_out[*]}" -- "$cur" ) )
          return 0
      fi
      return 1
    fi
    __kt_parse_get "${nouns[${#nouns[@]} -1]}"
}
__kt_get_resource_namespace()
{
    __kt_parse_get "namespace"
}
__kt_abort() {
    return 1
}
__kt_custom_func() {
    __kt_get_resource
}
`
)

var (
	bashCompletionFlags = map[string]string{
		"namespace": "__kt_get_resource_namespace",
		"context":   "__kt_config_get_contexts",
		"cluster":   "__kt_config_get_clusters",
		"user":      "__kt_config_get_users",

		"container":  "__kt_abort",
		"kubeconfig": "__kt_abort",
		"selector":   "__kt_abort",
		"server":     "__kt_abort",
		"since":      "__kt_abort",
		"since-time": "__kt_abort",
		"tail":       "__kt_abort",
	}
)
