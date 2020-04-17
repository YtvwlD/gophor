package main

/* RegularFileContents:
 * Very simple implementation of FileContents that just
 * buffered reads from the stored file path, stores the
 * read bytes in a slice and returns when requested.
 */
type RegularFileContents struct {
    path     string
    contents []byte
}

func (fc *RegularFileContents) Render() []byte {
    return fc.contents
}

func (fc *RegularFileContents) Load() *GophorError {
    var gophorErr *GophorError
    fc.contents, gophorErr = bufferedRead(fc.path)
    return gophorErr
}

func (fc *RegularFileContents) Clear() {
    fc.contents = nil
}
