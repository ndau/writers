# `testwriter`

`testwriter` wraps a `testing.T` in a `linewriter`, calling
`t.Log` on every newline. It implements `io.Writer`.

The intent is that within a test suite, you can redirect all logging calls to the test log, i.e. (`sirupsen/logrus` syntax):

```go
func TestSomething(t *testing.T) {
    logger := logrus.New()
    logger.Out = testwriter(t)

    logger.WithFields(logrus.Fields{
        "animal": "walrus",
        "size":   10,
    }).Info("A group of walrus emerges from the ocean")
    // this generates a log record which is sent to Stdout
    // only if the test fails or the test is run in verbose mode
}
```
