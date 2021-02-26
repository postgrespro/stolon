// Copyright (c) 2021, Postgres Professional

package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
)

type TeeLogger struct {
	pctx context.Context
	p *PostgresKeeper
	host string
	base *zap.SugaredLogger
}

func (s *TeeLogger) Warn(args ...interface{}) {
	s.base.Warn(args...)
}

func (s *TeeLogger) Errorw(msg string, keysAndValues ...interface{}) {
	s.p.e.PutKeeperFeedback(s.pctx, filepath.Join("remoteLogs", s.host, s.p.cfg.ClusterName, s.p.keeperLocalState.UID, "error"), fmt.Sprintf("%s: %v", msg, s.mergeFields(keysAndValues)))
	s.base.Errorw(msg, keysAndValues...)
}

func (s *TeeLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	s.p.e.PutKeeperFeedback(s.pctx, filepath.Join("remoteLogs", s.host, s.p.cfg.ClusterName, s.p.keeperLocalState.UID, "fatal"), fmt.Sprintf("%s: %v", msg, s.mergeFields(keysAndValues)))
	s.base.Errorw(msg, keysAndValues...)
}

func (s *TeeLogger) Warnw(msg string, keysAndValues ...interface{}) {
	s.base.Warnw(msg, keysAndValues...)
}

func (s *TeeLogger) Infow(msg string, keysAndValues ...interface{}) {
	s.base.Infow(msg, keysAndValues...)
}

func (s *TeeLogger) Debugw(msg string, keysAndValues ...interface{}) {
	s.base.Debugw(msg, keysAndValues...)
}

func (s *TeeLogger) Infof(template string, args ...interface{}) {
	s.base.Infof(template, args...)
}

func (s *TeeLogger) Debugf(template string, args ...interface{}) {
	s.base.Debugf(template, args...)
}

func (s *TeeLogger) mergeFields(args []interface{}) string {
	if len(args) == 0 {
		return ""
	}
	res := ""
	var invalid []invalidPair

	for i := 0; i < len(args); {
		// This is a strongly-typed field. Consume it and move on.
		if f, ok := args[i].(zap.Field); ok {
			if res != "" {
				res += ", "
			}
			if f.Key == "error" {
				res = res + fmt.Sprintf("%v", f.Interface)
			} else {
				res = res + fmt.Sprintf("%s: %v, ", f.Key, f.Interface)
			}

			i++
			continue
		}

		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			s.base.DPanic("Ignored key without a value.", zap.Any("ignored", args[i]))
			break
		}

		// Consume this value and the next, treating them as a key-value pair. If the
		// key isn't a string, add this pair to the slice of invalid pairs.
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			// Subsequent errors are likely, so allocate once up front.
			if cap(invalid) == 0 {
				invalid = make([]invalidPair, 0, len(args)/2)
			}
			invalid = append(invalid, invalidPair{i, key, val})
		} else {
			if res != "" {
				res += ", "
			}
			res = res + fmt.Sprintf("%s: %v, ", keyStr, val)
		}
		i += 2
	}

	// If we encountered any invalid key-value pairs, log an error.
	if len(invalid) > 0 {
		s.base.DPanicf("Ignored key-value pairs with non-string keys.", zap.Any("invalid", invalid))
	}
	return res
}

type invalidPair struct {
	position   int
	key, value interface{}
}
