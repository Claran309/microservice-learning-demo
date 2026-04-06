package log

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

type MyEncoder struct {
	AppName string
	zapcore.Encoder
	errFile *os.File
	writer  MyLogWriter
}

type MyLogWriter struct {
	mu      sync.Mutex
	logDate string
	file    *os.File
	logPath string
}

func (m *MyEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := m.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	DataStr := buf.String()
	buf.Reset()
	buf.AppendString("[" + m.AppName + "] " + DataStr)

	m.writer.mu.Lock()
	defer m.writer.mu.Unlock()
	currentDate := time.Now().Format("2006-01-02")

	if m.writer.logDate != currentDate {
		if m.writer.file != nil {
			m.writer.file.Close()
		}

		newPath := filepath.Join(m.writer.logPath, "/", currentDate)
		if err := os.MkdirAll(newPath, 0755); err != nil {
			return nil, errors.New("创建文件夹错误")
		}
		filePath := filepath.Join(newPath, " INFO.log")
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, errors.New("打开文件错误" + filePath)
		}
		m.writer.file = file
		m.writer.logDate = currentDate
	}

	if entry.Level >= zapcore.ErrorLevel {
		if m.errFile == nil {
			newPath := filepath.Join(m.writer.logPath, "/", currentDate)
			filePath := filepath.Join(newPath, " ERR.log")
			errFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			m.errFile = errFile
		}
		m.errFile.WriteString(buf.String())
	}

	if m.writer.logDate == currentDate {
		m.writer.file.WriteString(buf.String())
	}

	return buf, nil
}

func InitLogManager(appName string, logPath string) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		&MyEncoder{
			AppName: appName,
			Encoder: zapcore.NewConsoleEncoder(encoderConfig),
			writer:  MyLogWriter{logPath: logPath},
		},
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zapcore.DebugLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	zap.ReplaceGlobals(logger)
}
