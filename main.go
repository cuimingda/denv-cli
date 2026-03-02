package main

import (
    "os"

    "github.com/cuimingda/denv-cli/cmd"
)

func main() {
    if err := cmd.NewRootCmd().Execute(); err != nil {
        os.Exit(1)
    }
}
