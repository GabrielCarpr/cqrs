package log

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/google/uuid"
)

type LogLevel int

const (
	DEBUG LogLevel = iota + 1
	INFO
	WARN
	ERROR
)

var level = INFO

func SetLevel(lvl LogLevel) {
	level = lvl
}

type F map[string]string

func logln(ctx context.Context, lvl string, loc bool, msg string, fields F) {
	fn, file, line, _ := runtime.Caller(2)
	fun := runtime.FuncForPC(fn).Name()
	if id := GetID(ctx); id != uuid.Nil {
		log.Printf("[%s] %s: %s %v [%s %s:%d]", id, lvl, msg, fields, fun, file, line)
	} else {
		log.Printf("%s: %s %v [%s %s:%d]", lvl, msg, fields, fun, file, line)
	}
}

func Debug(ctx context.Context, msg string, fields F) {
	if level <= DEBUG {
		logln(ctx, "DEBUG", false, msg, fields)
	}
}

func Info(ctx context.Context, msg string, fields F) {
	if level <= INFO {
		logln(ctx, "INFO", false, msg, fields)
	}
}

func Warn(ctx context.Context, msg string, fields F) {
	if level <= WARN {
		logln(ctx, "WARN", true, msg, fields)
	}
}

func Error(ctx context.Context, msg interface{}, fields F) error {
	err := interfaceToError(msg, fields)
	logln(ctx, "ERROR", true, err.Error(), fields)
	return err
}

func Fatal(msg string, fields F) {
	log.Printf("FATAL: %s %v", msg, fields)
	os.Exit(1)
}

func Panic(ctx context.Context, msg interface{}, fields F) {
	err := interfaceToError(msg, fields)
	logln(ctx, "PANIC", true, err.Error(), fields)
	panic(err)
}

func interfaceToError(msg interface{}, fields F) error {
	var err error
	switch v := msg.(type) {
	case string:
		err = errors.New(v)
		break
	case error:
		err = v
		break
	default:
		err = errors.New(fmt.Sprint(v))
	}

	for key, val := range fields {
		err = fmt.Errorf("%w: %s: %s", err, key, val)
	}

	return err
}
