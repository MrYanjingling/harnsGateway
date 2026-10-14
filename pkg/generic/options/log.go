package options

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"

	"harns-platform/pkg/logging"
)

// LoggingConfiguration configures the shared structured logger.
type LoggingConfiguration struct {
	// Format is the log output format: "text" or "json".
	Format string `json:"format"`
	// Verbosity is the V level threshold, mirroring klog -v.
	Verbosity int `json:"verbosity"`
	// Output is the log destination:
	//   ""       -> <binary dir>/../log/<service>.log (falls back to stderr)
	//   "stderr" -> standard error
	//   "stdout" -> standard output
	//   other    -> a file path (relative to the process working directory)
	Output string `json:"output"`

	// Log rotation (only applies when logging to a file).
	// MaxSize is the maximum size in megabytes of a log file before it is rotated.
	MaxSize int `json:"max-size"`
	// MaxBackups is the maximum number of old log files to retain (0 = keep all).
	MaxBackups int `json:"max-backups"`
	// MaxAge is the maximum number of days to retain old log files (0 = keep forever).
	MaxAge int `json:"max-age"`
	// Compress determines whether rotated log files are gzip-compressed.
	Compress bool `json:"compress"`
	// LocalTime uses the local time for the timestamp in the rotated file names.
	LocalTime bool `json:"local-time"`
}

func NewDefaultLoggingConfiguration() LoggingConfiguration {
	return LoggingConfiguration{
		Format:     logging.FormatText,
		Verbosity:  logging.DefaultVerbosity,
		Output:     "",
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   true,
		LocalTime:  true,
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
	case "stderr":
		return os.Stderr, nil
	case "stdout":
		return os.Stdout, nil
	case "":
		path := defaultLogFile()
		if path == "" {
			return os.Stderr, nil
		}
		f, err := l.rotatingWriter(path)
		if err != nil {
			logging.Warnf("failed to open default log file %q, falling back to stderr: %v", path, err)
			return os.Stderr, nil
		}
		return f, nil
	default:
		return l.rotatingWriter(l.Output)
	}
}

// rotatingWriter returns a size/age based rotating file writer.
func (l *LoggingConfiguration) rotatingWriter(path string) (io.Writer, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create log dir %q: %w", dir, err)
		}
	}
	return &lumberjack.Logger{
		Filename:   path,
		MaxSize:    l.MaxSize,
		MaxBackups: l.MaxBackups,
		MaxAge:     l.MaxAge,
		Compress:   l.Compress,
		LocalTime:  l.LocalTime,
	}, nil
}

// defaultLogFile returns <binary dir>/../log/<binary name>.log so that, given a
// package layout of <module>/bin/<service> + <module>/log/, logs are written to
// the log directory next to the bin directory.
func defaultLogFile() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	dir := filepath.Dir(exe)
	if dir == "" || dir == "." {
		return ""
	}
	return filepath.Join(filepath.Dir(dir), "log", filepath.Base(exe)+".log")
}
