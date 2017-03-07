# pophttpd

pophttpd is a trivial webserver for serving a queue of files.  It adds
a non-standard `POP` HTTP method which atomically serves a file from
the specified directory at most once.

Once a file has been popped, it does not appear in the directory
listing nor is offered in future `POP`s.  The file is, however, still
in the filesystem and served at its original path to accomodate
partial or retried `GET`s.

## building

    go install

## running

    (cd $QUEUE_DIR && $GOPATH/bin/pophttpd -port=8666)

## client usage

    curl -JOXPOP server:8666/

Warning: this renames one of the files in `$QUEUE_DIR` into
`$QUEUE_DIR`/.pop/, which may need to be garbage collected.
