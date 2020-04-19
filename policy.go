package main

const (
    CapsTxtStr   = "caps.txt"
    RobotsTxtStr = "robots.txt"
)

func generateCapsTxt() []byte {
    text := "CAPS"+CrLf
    text += CrLf
    text += "# This is an automatically generated"+CrLf
    text += "# server policy file: caps.txt"+CrLf
    text += CrLf
    text += "CapsVersion=1"+CrLf
    text += "ExpireCapsAfter=3600"+CrLf
    text += CrLf
    text += "PathDelimeter=/"+CrLf
    text += "PathIdentity=."+CrLf
    text += "PathParent=.."+CrLf
    text += "PathParentDouble=FALSE"+CrLf
    text += "PathEscapeCharacter=\\"
    text += "PathKeepPreDelimeter=FALSE"
    text += CrLf
    text += "ServerSoftware=Gophor"+CrLf
    text += "ServerSoftwareVersion="+GophorVersion+CrLf
    text += "ServerDescription="+*ServerDescription+CrLf
    text += "ServerGeolocationString="+*ServerGeoloc+CrLf
    text += CrLf
    text += "ServerSupportsStdinScripts=FALSE"+CrLf
    text += CrLf
    text += "ServerAdmin="+*ServerAdmin+CrLf
    text += CrLf
    text += "DefaultEncoding=ascii"+CrLf
    return []byte(text)
}

func generateRobotsTxt() []byte {
    text := "Usage-agent: *"+CrLf
    text += "Disallow: *"+CrLf
    text += CrLf
    text += "Crawl-delay: 99999"+CrLf
    text += CrLf
    text += "# This server does not support scraping"
    return []byte(text)
}
