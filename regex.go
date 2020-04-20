package main

import (
    "regexp"
    "strings"
)

var RestrictedFilesRegex []*regexp.Regexp

func compileUserRestrictedFilesRegex() {
    if *RestrictedFiles == "" {
        /* User not supplied any restricted files, return here */
        listDir = _listDir
        return
    }

    /* Try compiling the RestrictedFilesRegex from finalRegex */
    logSystem("Compiling restricted file regular expressions\n")

    /* Split the user supplied RestrictedFiles string by new-line */
    RestrictedFilesRegex = make([]*regexp.Regexp, 0)
    for _, expr := range strings.Split(*RestrictedFiles, "\n") {
        regex, err := regexp.Compile(expr)
        if err != nil {
            logSystemFatal("Failed compiling user restricted files regex: %s\n", expr)
        }
        RestrictedFilesRegex = append(RestrictedFilesRegex, regex)
    }

    listDir = _listDirRegexMatch
}

func isRestrictedFile(name string) bool {
    for _, regex := range RestrictedFilesRegex {
        if regex.MatchString(name) {
            return true
        }
    }
    return false
}
