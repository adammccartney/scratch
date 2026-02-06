package main

import (
	"os"
	"text/template"
)

func main() {
	t := template.Must(template.New("script").Parse(`
#!/usr/bin/env bash
set -euo pipefail

CMD={{.Cmd}}
SWS_VER={{.Sws}}
TOOLCHAIN={{.Toolchain}}

cmd="some nicely formatted cmd: ${CMD} ${SWS_VER} ${TOOLCHAIN}"
echo "$cmd"
`))

	data := map[string]string{
		"Cmd":        os.Args[1],
		"Sws":        os.Args[2],
		"Toolchain":  os.Args[3],
	}

	t.Execute(os.Stdout, data)
}

