package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

const (
	// LogFieldKeyReqID - is the logfield key for RequestID
	LogFieldKeyReqID = "requestID"
)

// NewLoggerFromReqIDStr - creates a *logrus.Entry that has requestID as a field.
// A new LogField instance will be created if log is nil
func NewLoggerFromReqIDStr(reqID string, ancestorLogger logrus.FieldLogger) logrus.FieldLogger {
	var retLogger logrus.FieldLogger = logrus.StandardLogger()

	if ancestorLogger != nil {
		retLogger = ancestorLogger
	}

	return retLogger.WithField(LogFieldKeyReqID, reqID)
}

// NewLoggerFromReqIDCtx - creates a *logrus.Entry that has requestID as a field.
// A new LogField instance will be created if log is nil
func NewLoggerFromReqIDCtx(
	ctx context.Context, ancestorLogger logrus.FieldLogger,
) logrus.FieldLogger {
	reqID := GetReqID(ctx)

	return NewLoggerFromReqIDStr(reqID, ancestorLogger)
}
