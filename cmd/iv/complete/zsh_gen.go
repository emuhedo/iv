// +build ignore

// Generate APL completion for zsh
package main

import (
	"fmt"
	"os"
)

// This program generates _iv. It is called by "go generate".

func main() {
	w, err := os.Create("_iv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	defer w.Close()

	fmt.Fprintf(w, `#compdef _iv iv

# autogenerated by zsh_gen.go (do not edit!)
#
# Activate manually by:
# 	. ./_iv iv
#	compdef _iv iv
#
# This only replaces complete apl names with it's symbol.

function _iv {

	# word is the current word to be completed.
	word=${words[CURRENT]}
	
        save="${IFS}"
        IFS=$'\n' a=($(iv -complete-bash "dummy" "${word}" ))
        IFS="${save}"

        if [[ ${a[1]} == "Symbols" ]]; then
                compadd -x "${a}"
        else
                compadd -S '' -U ${a}
        fi
}
`)
}
