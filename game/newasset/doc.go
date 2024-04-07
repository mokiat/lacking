// Package asseet provides data transfer objects for the game's asset types.
package asset

/*
Note: Here are some findings when doing benchmark tests for various I/O approaches.

goos: linux
goarch: amd64
cpu: AMD Ryzen 7 3700X 8-Core Processor
disk: SSD

The GOB encoder/decoder is twice slower than a manual (using own storage API)
binary serialization, which is really surprising. The main downside of Gob is that
it uses twice the memory to load a resource, whereas the manual one uses only
what is needed.

Adding a bufio.NewWriter and bufio.NewReader wrappers over file Writers and Readers
respectively makes no difference.

Adding default Zlib wrappers decreases performance by a factor of 30 compared to
the binary serialization. It overallocates a bit as well, though not much. It
does reduce the size to about 30% the one produced by the binary approach.

Changing to best-compression zlib does not make any difference compared to the
default settings.

Changing to best-speed zlib leads to the worst zlib performance, compression,
and memory usage, which was unexpected.

Without any compression, the size is increased 5x the original source image for
HDR images.
*/
