# oneiro writers

The `writers` repo contains some utility writers:

- `linewriter` is a buffered writer which is guaranteed to flush at every newline
- `testwriter` converts each line of input into a `t.Log` call in the provided test object. It's meant to convert application log output into test log lines.
- `ringbuffer` is a buffered io.ReadWriteCloser that is safe to read and write from different goroutines. It's compatible with a Scanner and is intended to be used to read JSON objects that are posted to a log and which may be buffered in awkward ways.
- `filter` is a writer that processes the data written to it and feeds it after processing to an output function
