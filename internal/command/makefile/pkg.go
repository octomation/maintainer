package makefile

import "io"

type Writer func([]byte) (int, error)

func (w Writer) Write(input []byte) (int, error) { return w(input) }

func DeduplicateNewLines(stream io.Writer) Writer {
	var counter int
	return func(input []byte) (int, error) {
		filtered := input[:0]
		for _, r := range input {
			if counter > 1 && r == '\n' {
				continue
			}
			if r == '\n' {
				counter++
			} else {
				counter = 0
			}
			filtered = append(filtered, r)
		}

		_, err := stream.Write(filtered)
		return len(input), err
	}
}
