package makefile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	distributionDir = "dist"
	outputFilename  = "Makefile"

	// https://www.gnu.org/software/make/manual/html_node/Include.html
	includeDirective  = "include "
	sincludeDirective = "-include "
)

type Errors []error

func (batch Errors) Error() string {
	transformed := make([]string, 0, len(batch))
	for _, err := range batch {
		transformed = append(transformed, err.Error())
	}
	return strings.Join(transformed, "\n")
}

func (batch Errors) Reduce() error {
	reduced := batch[:0]
	for _, err := range batch {
		if err != nil {
			reduced = append(reduced, err)
		}
	}
	if len(reduced) == 0 {
		return nil
	}
	return reduced
}

type Makefile string

func (makefile Makefile) AppendTo(output io.Writer) error {
	file, err := os.Open(makefile.Name())
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, includeDirective) && !strings.HasPrefix(text, sincludeDirective) {
			if _, err := fmt.Fprintln(output, text); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(text, sincludeDirective) {
			name := strings.TrimSpace(strings.TrimPrefix(text, sincludeDirective))
			if err := Makefile(name).AppendTo(output); err == nil {
				_, _ = fmt.Fprintln(output)
			}
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(text, includeDirective))
		if err := Makefile(name).AppendTo(output); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(output); err != nil {
			return err
		}
	}
	return nil
}

func (makefile Makefile) Name() string {
	return string(makefile)
}

func (makefile Makefile) CompileTo(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	output, err := os.Create(filepath.Join(dir, outputFilename))
	if err != nil {
		return err
	}
	defer func() { _ = output.Close() }()
	return makefile.AppendTo(output)
}

type Makefiles []Makefile

func (batch Makefiles) CompileTo(dir string) error {
	wg, err := sync.WaitGroup{}, make(Errors, len(batch))
	for idx, makefile := range batch {
		wg.Add(1)
		go func(idx int, makefile Makefile) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					err[idx] = fmt.Errorf("%#+v", r)
				}
			}()
			err[idx] = makefile.CompileTo(
				filepath.Join(dir,
					strings.TrimSuffix(
						filepath.Base(makefile.Name()),
						filepath.Ext(makefile.Name()),
					),
				),
			)
		}(idx, makefile)
	}
	wg.Wait()
	return err.Reduce()
}
