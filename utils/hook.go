package utils

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// 以天为单位分割日志的钩子
// 为 logrus.Logger 添加此钩子仅能实现日志刷新，如果要写入还需将其设为 logrus.Logger.Out
type DateHook struct {
	// 输出日志文件的路径
	// 会利用日志事件的时间进行格式化处理
	// 参考值 "logs/2006-01-02.log"
	Format string

	// 日志等级
	// 为空默认 logrus.AllLevels
	LogLevels []logrus.Level

	// 日志文件
	file *os.File

	// 当前日志文件的日期
	year  int
	month time.Month
	day   int
}

func (d DateHook) Levels() []logrus.Level {
	if d.LogLevels != nil {
		return d.LogLevels
	}
	return logrus.AllLevels
}

func (d *DateHook) Fire(entry *logrus.Entry) (err error) {
	// 获取当前日期
	year, month, day := entry.Time.Date()
	// 与当前日志文件的日期比较
	if d.day != day || d.month != month || d.year != year {
		// 创建新日志文件的前置路径
		file := entry.Time.Format(d.Format)
		err = os.MkdirAll(filepath.Dir(file), os.ModePerm)
		if err != nil {
			return
		}
		// 关闭当前日志
		if d.file != nil {
			err = d.file.Close()
			if err != nil {
				return
			}
		}
		// 打开新日志
		d.file, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			return
		}
		// 刷新日期
		d.year, d.month, d.day = year, month, day
	}
	return
}

// 写入日志
func (d *DateHook) Write(p []byte) (int, error) {
	return d.file.Write(p)
}

// 关闭日志
func (d *DateHook) Close() error {
	return d.file.Close()
}

var _ logrus.Hook = (*DateHook)(nil)
var _ io.WriteCloser = (*DateHook)(nil)
