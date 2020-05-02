package main

import (
    "regexp"
    "strings"
)

func compileUserRestrictedFilesRegex(restrictedFiles string) []*regexp.Regexp {
    /* Return slice */
    restrictedFilesRegex := make([]*regexp.Regexp, 0)

    /* Split the user supplied RestrictedFiles string by new-line */
    for _, expr := range strings.Split(restrictedFiles, "\n") {
        regex, err := regexp.Compile(expr)
        if err != nil {
            Config.SysLog.Fatal("Failed compiling user restricted files regex: %s\n", expr)
        }
        restrictedFilesRegex = append(restrictedFilesRegex, regex)
    }

    return restrictedFilesRegex
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
