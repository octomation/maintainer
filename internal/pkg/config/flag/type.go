package flag

import "github.com/spf13/pflag"

type Set pflag.FlagSet

func Adopt(set *pflag.FlagSet) *Set { return (*Set)(set) }
