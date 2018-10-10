# `LineWriter`

`LineWriter` wraps an `io.Writer` and buffers output to it, flushing whenever a newline (`0x0a`, `\n`) is detected.

The `bufio.Writer` struct wraps a writer and buffers its output. Howveer, it only does this batched write when the internal buffer fills. Sometimes, you'd prefer to write each line as it's completed, rather than the entire buffer at once. Enter `LineWriter`. It does exactly that.

Like `bufio.Writer`, a `LineWriter`'s buffer will also be flushed when its internal buffer is full. Like `bufio.Writer`, after all data has been written, the client should call the `Flush` method to guarantee that all data has been forwarded to the underlying `io.Writer`.

The fundamental concept here is shamelessly stolen from Rust's [`std::io::LineWriter`](https://doc.rust-lang.org/std/io/struct.LineWriter.html).
