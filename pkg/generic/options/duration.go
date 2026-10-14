package options

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// Duration wraps time.Duration so that it can be written in YAML config files
// using human-readable values such as "15s" or "1m", while still being bound to
// a pflag flag as a duration. It is shared by all components' option structs.
type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(time.Duration(d).String())), nil
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*d = 0
		return nil
	}
	if parsed, err := time.ParseDuration(s); err == nil {
		*d = Duration(parsed)
		return nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		*d = Duration(n)
		return nil
	}
	return fmt.Errorf("invalid duration %q", s)
}

// DurationVar defines a Duration flag with the specified name and usage string,
// mirroring pflag.DurationVar but accepting human-readable duration values.
func DurationVar(fs *pflag.FlagSet, p *Duration, name, usage string) {
	fs.Var(&durationValue{p}, name, usage)
}

type durationValue struct {
	target *Duration
}

func (v durationValue) String() string {
	return time.Duration(*v.target).String()
}

func (v durationValue) Set(s string) error {
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*v.target = Duration(parsed)
	return nil
}

func (v durationValue) Type() string {
	return "duration"
}
