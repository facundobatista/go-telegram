/*
Copyright 2014 Facundo Batista

This program is free software: you can redistribute it and/or modify it
under the terms of the GNU General Public License version 3, as published
by the Free Software Foundation.

This program is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranties of
MERCHANTABILITY, SATISFACTORY QUALITY, or FITNESS FOR A PARTICULAR
PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package logging

import (
	"log"
	"os"
)

const (
	LevelError = iota
	LevelInfo
	LevelDebug
)

type simpleLogger struct {
	nlevel int
	logger *log.Logger
}

func New(level int) *simpleLogger {
	lg := new(simpleLogger)
	lg.nlevel = level
	lg.logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	return lg
}

func (lg *simpleLogger) SetLevel(level int) {
	lg.nlevel = level
}

func (lg *simpleLogger) Error(format string, v ...interface{}) {
	lg.logger.Printf("ERROR "+format, v...)
}

func (lg *simpleLogger) Fatal(format string, v ...interface{}) {
	lg.logger.Printf("ERROR "+format, v...)
	os.Exit(1)
}

func (lg *simpleLogger) Info(format string, v ...interface{}) {
	if lg.nlevel >= LevelInfo {
		lg.logger.Printf("INFO  "+format, v...)
	}
}

func (lg *simpleLogger) Debug(format string, v ...interface{}) {
	if lg.nlevel >= LevelDebug {
		lg.logger.Printf("DEBUG "+format, v...)
	}
}
