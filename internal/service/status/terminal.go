package status

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// Interactive owns raw mode only while the view is active. Polling the input
// descriptor permits cancellation and resize without a stranded input reader.
func Interactive(ctx context.Context, in, out *os.File, rows []Row) (err error) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	fd := int(in.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, term.Restore(fd, old)) }()
	defer func() {
		_, restoreErr := fmt.Fprint(out, "\x1b[0m\x1b[?25h\x1b[?1049l")
		err = errors.Join(err, restoreErr)
	}()
	if _, err := fmt.Fprint(out, "\x1b[?1049h\x1b[?25l\x1b[2J"); err != nil {
		return err
	}
	t := newTable(rows)
	selection := new(Selection)
	last, pending := "", ""
	var escapeAt time.Time
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		width, height, err := term.GetSize(int(out.Fd()))
		if err != nil {
			return err
		}
		current := frame(t, rows, selection, width, height)
		if current != last {
			if _, err := fmt.Fprint(out, current); err != nil {
				return err
			}
			last = current
		}
		poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		n, err := unix.Poll(poll, 100)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			return err
		}
		if n == 0 {
			if pending == "\x1b" && time.Since(escapeAt) > 150*time.Millisecond {
				return nil
			}
			continue
		}
		if poll[0].Revents&(unix.POLLHUP|unix.POLLERR|unix.POLLNVAL) != 0 {
			return nil
		}
		var data [64]byte
		n, err = unix.Read(fd, data[:])
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
		pending += string(data[:n])
		escapeAt = time.Now()
		for len(pending) > 0 {
			key, rest, complete := nextKey(pending)
			if !complete {
				break
			}
			pending = rest
			if key == "q" || key == "\x03" || key == "\x04" {
				return nil
			}
			selection.Move(key, len(rows), max(1, height-6))
		}
	}
}

func nextKey(input string) (key, rest string, complete bool) {
	if input[0] != '\x1b' {
		return input[:1], input[1:], true
	}
	if len(input) == 1 {
		return "", input, false
	}
	if input[1] != '[' && input[1] != 'O' {
		return input[:2], input[2:], true
	}
	for i := 2; i < len(input); i++ {
		if input[i] >= 0x40 && input[i] <= 0x7e {
			return input[:i+1], input[i+1:], true
		}
	}
	return "", input, false
}
