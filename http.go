package main

import (
    "strconv"
)

func generateHttpResponse(code ErrorResponseCode, data string) []byte {
    dataLength := len(data)
    return []byte(
        "HTTP/1.1 "+code.String()+"\n"+
        "Content-Length: "+strconv.Itoa(dataLength)+"\n"+
        "Connection: close\n"+
        "Content-Type: text/html\n\n"+
        data)
}
