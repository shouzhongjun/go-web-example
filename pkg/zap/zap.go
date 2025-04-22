package zap

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"goWebExample/internal/configs"
)

// parseLevel 将字符串转换为 zapcore.Level
func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// customLevelEncoder 自定义日志级别编码器，添加颜色
func customLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var colorCode string
	switch l {
	case zapcore.DebugLevel:
		colorCode = "\x1b[36m" // Cyan
	case zapcore.InfoLevel:
		colorCode = "\x1b[34m" // Blue
	case zapcore.WarnLevel:
		colorCode = "\x1b[33m" // Yellow
	case zapcore.ErrorLevel:
		colorCode = "\x1b[31m" // Red
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		colorCode = "\x1b[35m" // Magenta
	default:
		colorCode = "\x1b[0m" // Reset
	}
	enc.AppendString(fmt.Sprintf("\t%s%s%s\t", colorCode, l.CapitalString(), "\x1b[0m"))
}

// customCallerEncoder 自定义调用者编码器
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%s\t", caller.TrimmedPath()))
}

// customTimeEncoder 自定义时间编码器
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.DateTime))
}

// RotateFileWriter 实现日志文件轮转和压缩
type RotateFileWriter struct {
	filename   string    // 日志文件名
	maxSize    int       // 单个日志文件最大大小，单位MB
	maxBackups int       // 保留的旧日志文件最大数量
	maxAge     int       // 保留的旧日志文件最大天数
	compress   bool      // 是否压缩旧日志文件
	size       int64     // 当前日志文件大小
	file       *os.File  // 当前日志文件
	mu         sync.Mutex // 互斥锁
}

// NewRotateFileWriter 创建一个新的日志文件轮转器
func NewRotateFileWriter(filename string, maxSize, maxBackups, maxAge int, compress bool) *RotateFileWriter {
	return &RotateFileWriter{
		filename:   filename,
		maxSize:    maxSize,
		maxBackups: maxBackups,
		maxAge:     maxAge,
		compress:   compress,
	}
}

// Write 实现io.Writer接口
func (w *RotateFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err = w.openFile(); err != nil {
			return 0, err
		}
	}

	// 检查文件大小是否超过限制
	if w.size+int64(len(p)) > int64(w.maxSize*1024*1024) {
		if err = w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// Sync 实现io.Closer接口
func (w *RotateFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

// Close 关闭文件
func (w *RotateFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

// openFile 打开日志文件
func (w *RotateFileWriter) openFile() error {
	// 确保目录存在
	dir := filepath.Dir(w.filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// 打开文件
	file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = file

	// 获取文件大小
	info, err := file.Stat()
	if err != nil {
		return err
	}
	w.size = info.Size()

	return nil
}

// rotate 轮转日志文件
func (w *RotateFileWriter) rotate() error {
	// 关闭当前文件
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}

	// 清理旧日志文件
	if err := w.cleanup(); err != nil {
		return err
	}

	// 重命名当前日志文件
	backupName := w.backupName()
	if err := os.Rename(w.filename, backupName); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 压缩旧日志文件
	if w.compress {
		go w.compressFile(backupName)
	}

	// 打开新的日志文件
	return w.openFile()
}

// backupName 生成备份文件名
func (w *RotateFileWriter) backupName() string {
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	return fmt.Sprintf("%s.%s", w.filename, timestamp)
}

// cleanup 清理旧日志文件
func (w *RotateFileWriter) cleanup() error {
	dir := filepath.Dir(w.filename)
	base := filepath.Base(w.filename)

	// 获取所有备份文件
	files, err := filepath.Glob(filepath.Join(dir, base+".*"))
	if err != nil {
		return err
	}

	// 按修改时间排序
	type fileInfo struct {
		name    string
		modTime time.Time
	}
	var fileInfos []fileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{name: file, modTime: info.ModTime()})
	}

	// 按修改时间排序
	if len(fileInfos) > 1 {
		// 按时间从新到旧排序
		for i := 0; i < len(fileInfos)-1; i++ {
			for j := i + 1; j < len(fileInfos); j++ {
				if fileInfos[i].modTime.Before(fileInfos[j].modTime) {
					fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
				}
			}
		}
	}

	// 删除超过最大备份数量的文件
	if w.maxBackups > 0 && len(fileInfos) > w.maxBackups {
		for i := w.maxBackups; i < len(fileInfos); i++ {
			os.Remove(fileInfos[i].name)
		}
	}

	// 删除超过最大保留天数的文件
	if w.maxAge > 0 {
		cutoff := time.Now().Add(-time.Duration(w.maxAge) * 24 * time.Hour)
		for _, f := range fileInfos {
			if f.modTime.Before(cutoff) {
				os.Remove(f.name)
			}
		}
	}

	return nil
}

// compressFile 压缩文件
func (w *RotateFileWriter) compressFile(filename string) error {
	// 打开源文件
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// 创建压缩文件
	gzFilename := filename + ".gz"
	gzf, err := os.Create(gzFilename)
	if err != nil {
		return err
	}
	defer gzf.Close()

	// 创建gzip写入器
	gz := gzip.NewWriter(gzf)
	defer gz.Close()

	// 复制内容
	if _, err := io.Copy(gz, f); err != nil {
		return err
	}

	// 关闭gzip写入器
	if err := gz.Close(); err != nil {
		return err
	}

	// 删除原文件
	return os.Remove(filename)
}

// getLogWriter 获取日志文件写入器
func getLogWriter(path, filename string, config *configs.AllConfig) zapcore.WriteSyncer {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic(err)
	}

	logFile := filepath.Join(path, filename)

	// 设置默认值
	maxSize := 100    // 默认100MB
	maxBackups := 0   // 默认保留所有
	maxAge := 0       // 默认保留所有
	compress := false // 默认不压缩

	// 如果配置了日志压缩相关参数，则使用配置的值
	if config.Log.MaxSize > 0 {
		maxSize = config.Log.MaxSize
	}
	if config.Log.MaxBackups > 0 {
		maxBackups = config.Log.MaxBackups
	}
	if config.Log.MaxAge > 0 {
		maxAge = config.Log.MaxAge
	}
	compress = config.Log.Compress

	// 创建支持日志轮转和压缩的写入器
	writer := NewRotateFileWriter(logFile, maxSize, maxBackups, maxAge, compress)

	return zapcore.AddSync(writer)
}

// NewZap 创建一个新的 zap 日志记录器
func NewZap(config *configs.AllConfig) *zap.Logger {
	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    customLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   customCallerEncoder,
	}

	// 解析日志级别
	level := parseLevel(config.Log.Level)

	var cores []zapcore.Core

	// 如果启用了控制台输出
	if config.Log.EnableConsole {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 如果启用了文件输出
	if config.Log.EnableFile {
		today := time.Now().Format("2006-01-02")

		// 普通日志（info及以下级别）
		normalLogWriter := getLogWriter(config.Log.Path, today+".info.log", config)
		normalCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			normalLogWriter,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl <= zapcore.InfoLevel && lvl >= level
			}),
		)
		cores = append(cores, normalCore)

		// 错误日志（warn及以上级别）
		errorLogWriter := getLogWriter(config.Log.Path, today+".error.log", config)
		errorCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			errorLogWriter,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl > zapcore.InfoLevel && lvl >= level
			}),
		)
		cores = append(cores, errorCore)
	}

	// 创建日志记录器
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller())
	return logger
}
