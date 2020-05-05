package main

import (
    "regexp"
    "strings"
)

func compileCmdParseRegex() *regexp.Regexp {
    return regexp.MustCompile(` `)
}

func compileUserRestrictedFilesRegex(restrictedFiles string) []*regexp.Regexp {
    /* Return slice */
    restrictedFilesRegex := make([]*regexp.Regexp, 0)

    /* Split the user supplied RestrictedFiles string by new-line */
    for _, expr := range strings.Split(restrictedFiles, "\n") {
        if len(expr) == 0 {
            continue
        }
        regex, err := regexp.Compile(expr)
        if err != nil {
            Config.SysLog.Fatal("Failed compiling user restricted files regex: %s\n", expr)
        }
        restrictedFilesRegex = append(restrictedFilesRegex, regex)
    }

    return restrictedFilesRegex
}

func compileUserRestrictedCommandsRegex(restrictedCommands string) []*regexp.Regexp {
    /* Return slice */
    restrictedCommandsRegex := make([]*regexp.Regexp, 0)

    /* Split the user supplied RestrictedFiles string by new-line */
    for _, expr := range strings.Split(restrictedCommands, "\n") {
        if len(expr) == 0 {
            continue
        }
        regex, err := regexp.Compile(expr)
        if err != nil {
            Config.SysLog.Fatal("Failed compiling user restricted commands regex: %s\n", expr)
        }
        restrictedCommandsRegex = append(restrictedCommandsRegex, regex)
    }

    return restrictedCommandsRegex
}


/* Iterate through restricted file expressions, check if file _is_ restricted */
func isRestrictedFile(name string) bool {
    for _, regex := range Config.RestrictedFiles {
        if regex.MatchString(name) {
            return true
        }
    }
    return false
}

func isRestrictedCommand(name string) bool {
    for _, regex := range Config.RestrictedCommands {
        if regex.MatchString(name) {
            return true
        }
    }
    return false
}
