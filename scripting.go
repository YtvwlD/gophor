package main

import (
    "regexp"
)

var (
    r = regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
)

func parseCommandString(commandStr string) []string {
    commands := r.FindAllString(commandStr, -1)
    return commands
}
