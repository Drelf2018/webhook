package utils

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type DateHook struct {
	Format string

	year  int
	month time.Month
	day   int
	file  *os.File
}

func (DateHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (d *DateHook) Fire(entry *logrus.Entry) (err error) {
	year, month, day := entry.Time.Date()
	if d.day != day || d.month != month || d.year != year {
		file := entry.Time.Format(d.Format)
		err = os.MkdirAll(filepath.Dir(file), os.ModePerm)
		if err != nil {
			return
		}
		if d.file != nil {
			err = d.file.Close()
			if err != nil {
				return
			}
		}
		d.day, d.month, d.year = day, month, year
		d.file, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	}
	return
}

func (d *DateHook) Write(p []byte) (int, error) {
	return d.file.Write(p)
}

func (d *DateHook) Open() (*os.File, error) {
	return os.Open(time.Now().Format(d.Format))
}

func (d *DateHook) Close() error {
	return d.file.Close()
}

var _ logrus.Hook = (*DateHook)(nil)
var _ io.WriteCloser = (*DateHook)(nil)
