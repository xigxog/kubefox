package logger

import (
	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Component interface {
	GetId() string
	GetGitHash() string
	GetName() string
}

type Log struct {
	*zap.SugaredLogger
}

func ProdLogger() *Log {
	z, _ := zap.NewProduction()

	return &Log{SugaredLogger: z.Sugar()}
}

func DevLogger() *Log {
	return devLogger(false)
}

func CLILogger() *Log {
	return devLogger(true).DisableStacktrace()
}

func devLogger(isCli bool) *Log {
	conf := zap.NewDevelopmentConfig()
	// disables stack traces except for error and fatal
	conf.Development = false
	conf.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	conf.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04.05")

	if isCli {
		conf.EncoderConfig.CallerKey = zapcore.OmitKey
		conf.EncoderConfig.TimeKey = zapcore.OmitKey
	}

	z, _ := conf.Build()

	return &Log{SugaredLogger: z.Sugar()}
}

func (log *Log) Named(name string) *Log {
	return &Log{SugaredLogger: log.SugaredLogger.Named(name)}
}

func (log *Log) With(args ...interface{}) *Log {
	return &Log{SugaredLogger: log.SugaredLogger.With(args...)}
}

// func (log *Log) WithOrganization(val string) *Log {
// 	return log.With("organization", val)
// }

func (log *Log) WithPlatform(val string) *Log {
	return log.With("platform", val)
}

func (log *Log) WithComponent(comp Component) *Log {
	return log.With(
		"componentId", comp.GetId(),
		"componentHash", comp.GetGitHash(),
		"componentName", comp.GetName(),
	)
}

// IncreaseLevel increase the level of the logger. It has no effect if the
// passed in level tries to decrease the level of the logger.
func (log *Log) IncreaseLevel(lvl zapcore.LevelEnabler) *Log {
	return &Log{SugaredLogger: log.WithOptions(zap.IncreaseLevel(lvl))}
}

// DisableStacktrace disables writing of stack traces except at panic level.
func (log *Log) DisableStacktrace() *Log {
	return &Log{SugaredLogger: log.WithOptions(zap.AddStacktrace(zapcore.PanicLevel))}
}

// DebugInterface marshals the interface to indented JSON. Since this operation
// is mildly expensive it is only performed if the logger is at debug level.
//
// A newline and marshaled JSON is automatically appended to the log message. It
// should not be included in the template.
func (log *Log) DebugInterface(v interface{}, template string, args ...interface{}) {
	if !log.SugaredLogger.Desugar().Core().Enabled(zap.DebugLevel) {
		return
	}

	template = template + "\n%s\n"
	out, _ := json.MarshalIndent(v, "", "\t")
	args = append(args, out)
	log.Debugf(template, args...)
}
