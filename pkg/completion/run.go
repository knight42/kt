package completion

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func Generate(cmd *cobra.Command, shell string) error {
	cmd.BashCompletionFunction = bashCompletionFunc

	for name, comp := range bashCompletionFlags {
		if cmd.Flag(name) != nil {
			if cmd.Flag(name).Annotations == nil {
				cmd.Flag(name).Annotations = map[string][]string{}
			}
			cmd.Flag(name).Annotations[cobra.BashCompCustom] = append(
				cmd.Flag(name).Annotations[cobra.BashCompCustom],
				comp,
			)
		}
	}

	switch shell {
	case "zsh":
		return runCompletionZsh(os.Stdout, cmd)
	case "bash":
		return cmd.GenBashCompletion(os.Stdout)
	default:
		return fmt.Errorf("unknown shell: %s", shell)
	}
}

func runCompletionZsh(out io.Writer, cmd *cobra.Command) error {
	zshHead := "#compdef kt\n"

	_, _ = out.Write([]byte(zshHead))

	zshInitialization := `
__kt_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__kt_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__kt_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__kt_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__kt_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__kt_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__kt_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__kt_filedir() {
	true
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__kt_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__kt_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__kt_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__kt_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__kt_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__kt_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__kt_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	_, _ = out.Write([]byte(zshInitialization))

	buf := new(bytes.Buffer)
	_ = cmd.GenBashCompletion(buf)
	_, _ = out.Write(buf.Bytes())

	zshTail := `
BASH_COMPLETION_EOF
}
__kt_bash_source <(__kt_convert_bash_to_zsh)
_complete kt 2>/dev/null
`
	_, _ = out.Write([]byte(zshTail))
	return nil
}
