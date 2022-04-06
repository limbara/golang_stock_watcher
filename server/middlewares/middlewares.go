package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/limbara/stock-watcher/constant"
	"github.com/limbara/stock-watcher/customerrors"
	"github.com/limbara/stock-watcher/utils"
	"go.uber.org/zap"
)

func ErrorHandlerMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err, ok := recover().(error); ok {
				var responseError *customerrors.ResponseError

				if errors.As(err, &responseError) {
					responseError.SetError(err)
				} else {
					responseError = customerrors.ResponseErrorServer.SetError(err)

					logger, err := utils.Logger()
					if err != nil {
						responseError.SetError(err)
					}

					uuid := r.Context().Value(constant.ContextKeyRequestId)

					logger.Sugar().Errorf("Response ID %s Recover Error:\n %+v", uuid, responseError)
				}

				w.WriteHeader(responseError.HttpCode)
				json.NewEncoder(w).Encode(responseError)
				return
			}
		}()

		handler.ServeHTTP(w, r)
	})
}

func RequestLoggerMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		logger, err := utils.Logger()

		if err != nil {
			panic(fmt.Errorf("Request Logger Middleware Logger Error : %w", err))
		} else {
			uuid := r.Context().Value(constant.ContextKeyRequestId)

			logger.Info(
				fmt.Sprintf("Request ID %s", uuid),
				zap.String("Remote Addr", r.RemoteAddr),
				zap.String("Method", r.Method),
				zap.String("Url", r.URL.EscapedPath()),
				zap.Any("Header", r.Header),
				zap.Duration("Time", time.Since(start)*time.Millisecond),
			)
		}
	})
}

func AddContextWantJsonMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		ctx := r.Context()

		if strings.Contains(contentType, "/json") || strings.Contains(contentType, "+json") {
			ctx = context.WithValue(ctx, constant.ContextKeyWantJson, true)
		} else {
			ctx = context.WithValue(ctx, constant.ContextKeyWantJson, false)
		}

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AddContextRequestIdMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), constant.ContextKeyRequestId, uuid.New().String())

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
