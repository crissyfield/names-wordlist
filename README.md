## Benchmarking

| Ver | Speed   | Description                                |
|-----|---------|--------------------------------------------|
|  1  | 26.473s | Straight forward implementation            |
|  2  | 11.286s | Interleaved casing, made on Fprintf() call |
|  3  |  9.978s | Move from Fprintf() to WriteString()       |
