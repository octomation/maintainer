package flag

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

const fileType = "file"

func (f *Set) GetFile(name string) (*os.File, error) {
	flag := (*pflag.FlagSet)(f).Lookup(name)
	if flag == nil {
		return nil, fmt.Errorf("flag accessed but not defined: %s", name)
	}

	switch val := flag.Value.(type) {
	case *File:
		if val.lazy == "" {
			return val.file, nil
		}
		if err := val.Set(val.lazy); err != nil {
			return nil, err
		}
		return val.file, nil
	default:
		return nil, fmt.Errorf("trying to get %s value of flag of type %s", fileType, flag.Value.Type())
	}
}

func (f *Set) File(name, value, usage string) *File {
	p := new(File)
	f.FileVarP(p, name, "", File{lazy: value}, usage)
	return p
}

func (f *Set) FileP(name, shorthand, value, usage string) *File {
	p := new(File)
	f.FileVarP(p, name, shorthand, File{lazy: value}, usage)
	return p
}

func (f *Set) FileVarP(p *File, name, shorthand string, value File, usage string) {
	(*pflag.FlagSet)(f).VarP(newFileValue(value, p), name, shorthand, usage)
}

func newFileValue(value File, p *File) *File {
	*p = value
	return p
}

type File struct {
	file *os.File
	lazy string
}

func (val *File) String() string {
	if val.file == nil {
		return val.lazy
	}
	return val.file.Name()
}

func (val *File) Set(name string) error {
	if val.file != nil {
		if err := val.file.Close(); err != nil {
			return err
		}
	}
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	val.file = file
	val.lazy = ""
	return nil
}

func (val *File) Type() string {
	return fileType
}
