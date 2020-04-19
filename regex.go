package main

import (
    "regexp"
    "strings"
)

var RestrictedFilesRegex *regexp.Regexp

func compileUserRestrictedFilesRegex() {
    if *RestrictedFiles == "" {
        /* User not supplied any restricted files, return here */
        getItemType = _getItemType
        return
    }

    /* Split the user supplied RestrictedFiles string by new-line */
    finalRegex := ""
    for _, pattern := range strings.Split(*RestrictedFiles, "\n") {
        if finalRegex != "" {
            finalRegex += "|"
        }
        finalRegex += "(" + pattern + ")"
    }

    /* Try compiling the RestrictedFilesRegex from finalRegex */
    logSystem("Compiling restricted files regex\n")

    var err error
    RestrictedFilesRegex, err = regexp.Compile(finalRegex)
    if err != nil {
        logSystemFatal("Failed compiling user restricted files regex: %s\n", finalRegex)
    }

    getItemType = _getItemTypeMatchRestricted
}
