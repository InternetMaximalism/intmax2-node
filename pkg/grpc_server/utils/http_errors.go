package utils

import (
	"context"
	"fmt"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gateway/http_response_modifier"
	"net/http"
	"runtime"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	badRequest          = "Bad request"
	forbidden           = "Forbidden"
	unauthorized        = "Unauthorized"
	internalServerError = "Internal server error"
)

const (
	name           = "HTTP answer"
	httpCode       = "answer_code"
	errCodeStart   = 399
	errCodeFinish  = 600
	maskOKMessage  = "[ok] in %s[%s:%d] OK"
	okDescription  = "ok.description"
	maskErrMessage = "[error] in %s[%s:%d] %v"
	errDescription = "error.description"
	callerNumber   = 1
	okAttribute    = "ok"
	errAttribute   = "error"
)

// OK sets http-header with status code equal 200.
func OK(ctx context.Context) error {
	spanCtx, span := open_telemetry.Tracer().Start(ctx, name,
		trace.WithAttributes(
			attribute.String(httpCode, strconv.Itoa(http.StatusOK)),
		))
	defer span.End()

	pc, fn, line, _ := runtime.Caller(callerNumber)
	span.SetAttributes(attribute.Key(okDescription).
		String(fmt.Sprintf(maskOKMessage, runtime.FuncForPC(pc).Name(), fn, line)))
	span.SetAttributes(attribute.Key(okAttribute).Bool(true))

	_ = grpc.SetHeader(spanCtx, metadata.Pairs(http_response_modifier.OutHTTPCode, strconv.Itoa(http.StatusOK)))
	return nil
}

// Unauthorized sets http-header with status code equal 401.
func Unauthorized(ctx context.Context, err ...error) error {
	var e error
	if len(err) == 1 {
		e = err[0]
	}

	spanCtx, span := open_telemetry.Tracer().Start(ctx, name,
		trace.WithAttributes(
			attribute.String(httpCode, strconv.Itoa(http.StatusUnauthorized)),
		))
	defer span.End()

	pc, fn, line, _ := runtime.Caller(callerNumber)
	span.SetAttributes(attribute.Key(errDescription).
		String(fmt.Sprintf(maskErrMessage, runtime.FuncForPC(pc).Name(), fn, line, e)))
	span.SetAttributes(attribute.Key(errAttribute).Bool(true))

	return Custom(spanCtx, codes.Unauthenticated, http.StatusUnauthorized, unauthorized, e)
}

// Forbidden sets http-header with status code equal 403.
func Forbidden(ctx context.Context, err ...error) error {
	var e error
	if len(err) == 1 {
		e = err[0]
	}

	spanCtx, span := open_telemetry.Tracer().Start(ctx, name,
		trace.WithAttributes(
			attribute.String(httpCode, strconv.Itoa(http.StatusForbidden)),
		))
	defer span.End()

	pc, fn, line, _ := runtime.Caller(callerNumber)
	span.SetAttributes(attribute.Key(errDescription).
		String(fmt.Sprintf(maskErrMessage, runtime.FuncForPC(pc).Name(), fn, line, e)))
	span.SetAttributes(attribute.Key(errAttribute).Bool(true))

	return Custom(spanCtx, codes.PermissionDenied, http.StatusForbidden, forbidden, e)
}

// Internal sets http-header with status code equal 500.
func Internal(ctx context.Context, log logger.Logger, format string, args ...any) error {
	log.Errorf(format, args...)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, name,
		trace.WithAttributes(
			attribute.String(httpCode, strconv.Itoa(http.StatusInternalServerError)),
		))
	defer span.End()

	pc, fn, line, _ := runtime.Caller(callerNumber)
	span.SetAttributes(attribute.Key(errDescription).
		String(fmt.Sprintf(maskErrMessage, runtime.FuncForPC(pc).Name(), fn, line, fmt.Sprintf(format, args...))))
	span.SetAttributes(attribute.Key(errAttribute).Bool(true))

	return Custom(spanCtx, codes.Internal, http.StatusInternalServerError, internalServerError, nil)
}

// BadRequest sets http-header with status code equal 400.
func BadRequest(ctx context.Context, err error) error {
	spanCtx, span := open_telemetry.Tracer().Start(ctx, name,
		trace.WithAttributes(
			attribute.String(httpCode, strconv.Itoa(http.StatusBadRequest)),
		))
	defer span.End()

	pc, fn, line, _ := runtime.Caller(callerNumber)
	span.SetAttributes(attribute.Key(errDescription).
		String(fmt.Sprintf(maskErrMessage, runtime.FuncForPC(pc).Name(), fn, line, err)))
	span.SetAttributes(attribute.Key(errAttribute).Bool(true))

	return Custom(spanCtx, codes.InvalidArgument, http.StatusBadRequest, badRequest, err)
}

// Custom code.
func Custom(ctx context.Context, code codes.Code, statusCode int, msg string, err error) error {
	spanCtx, span := open_telemetry.Tracer().Start(ctx, name,
		trace.WithAttributes(
			attribute.String(httpCode, strconv.Itoa(statusCode)),
		))
	defer span.End()

	if statusCode > errCodeStart && statusCode < errCodeFinish {
		errMsg := msg
		if err != nil {
			errMsg = err.Error()
		}
		pc, fn, line, _ := runtime.Caller(callerNumber)
		span.SetAttributes(attribute.Key(errDescription).
			String(fmt.Sprintf(maskErrMessage, runtime.FuncForPC(pc).Name(), fn, line, errMsg)))
		span.SetAttributes(attribute.Key(errAttribute).Bool(true))
	}

	_ = grpc.SetHeader(spanCtx, metadata.Pairs(http_response_modifier.OutHTTPCode, strconv.Itoa(statusCode)))
	if err != nil {
		return status.Error(code, err.Error())
	}
	return status.Error(code, msg)
}
