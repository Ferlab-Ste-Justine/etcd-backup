package cmd

import (
	"fmt"
	"os"
)

func AbortOnErr(tmpl string, err error) {
	if err != nil {
		fmt.Println(fmt.Sprintf(tmpl, err.Error()))
		os.Exit(1)
	}
}
