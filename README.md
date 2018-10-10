# oneiro writers

The `writers` repo contains some utility writers:

- `linewriter` is a buffered writer which is guaranteed to flush at every newline
- `testwriter` converts each line of input into a `t.Log` call in the provided test object. It's meant to convert application log output into test log lines.
