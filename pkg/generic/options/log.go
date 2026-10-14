package options

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/pflag"

	"harns-platform/pkg/logging"
)

// LoggingConfiguration configures the shared structured logger.
type LoggingConfiguration struct {
	// Format is the log output format: "text" or "json".
	Format string `json:"format"`
	// Verbosity is the V level threshold, mirroring klog -v.
	Verbosity int `json:"verbosity"`
	// Output is the log destination: "stderr", "stdout" or a file path.
	Output string `json:"output"`
}

func NewDefaultLoggingConfiguration() LoggingConfiguration {
	return LoggingConfiguration{
		Format:    logging.FormatText,
		Verbosity: logging.DefaultVerbosity,
		Output:    "stderr",
	}
}

func (l *LoggingConfiguration) ValidateAndApply() error {
	if l.Format != logging.FormatText && l.Format != logging.FormatJSON {
		return fmt.Errorf("unsupported log format %q, permitted formats are %q and %q", l.Format, logging.FormatText, logging.FormatJSON)
	}
	w, err := l.outputWriter()
	if err != nil {
		return err
	}
	logging.SetLogger(logging.NewLogger(l.Format, w))
	logging.SetVerbosity(l.Verbosity)
	return nil
}

func (l *LoggingConfiguration) outputWriter() (io.Writer, error) {
	switch l.Output {
	case "", "stderr":
		return os.Stderr, nil
	case "stdout":
		return os.Stdout, nil
	default:
		f, err := os.OpenFile(l.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %q: %w", l.Output, err)
		}
		return f, nil
	}
}

func (l *LoggingConfiguration) BindLoggingFlags(fs *pflag.FlagSet) {
	fs.StringVar(&l.Format, "logging-format", l.Format, "Sets the log format. Permitted formats: \"text\", \"json\".")
	fs.IntVarP(&l.Verbosity, "verbosity", "v", l.Verbosity, "number for the log level verbosity")
	fs.StringVar(&l.Output, "logging-output", l.Output, "Log output destination: \"stderr\", \"stdout\" or a file path")
}
