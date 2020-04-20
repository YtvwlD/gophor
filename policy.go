package main

func generateCapsTxt() []byte {
    text := "CAPS"+DOSLineEnd
    text += DOSLineEnd
    text += "# This is an automatically generated"+DOSLineEnd
    text += "# server policy file: caps.txt"+DOSLineEnd
    text += DOSLineEnd
    text += "CapsVersion=1"+DOSLineEnd
    text += "ExpireCapsAfter=1800"+DOSLineEnd
    text += DOSLineEnd
    text += "PathDelimeter=/"+DOSLineEnd
    text += "PathIdentity=."+DOSLineEnd
    text += "PathParent=.."+DOSLineEnd
    text += "PathParentDouble=FALSE"+DOSLineEnd
    text += "PathEscapeCharacter=\\"+DOSLineEnd
    text += "PathKeepPreDelimeter=FALSE"+DOSLineEnd
    text += DOSLineEnd
    text += "ServerSoftware=Gophor"+DOSLineEnd
    text += "ServerSoftwareVersion="+GophorVersion+DOSLineEnd
    text += "ServerDescription="+*ServerDescription+DOSLineEnd
    text += "ServerGeolocationString="+*ServerGeoloc+DOSLineEnd
    text += "ServerDefaultEncoding=ascii"+DOSLineEnd
    text += DOSLineEnd
    text += "ServerAdmin="+*ServerAdmin+DOSLineEnd
    return []byte(text)
}

func generateRobotsTxt() []byte {
    text := "Usage-agent: *"+DOSLineEnd
    text += "Disallow: *"+DOSLineEnd
    text += DOSLineEnd
    text += "Crawl-delay: 99999"+DOSLineEnd
    text += DOSLineEnd
    text += "# This server does not support scraping"+DOSLineEnd
    return []byte(text)
}
