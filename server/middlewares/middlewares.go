package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/limbara/stock-watcher/constant"
	"github.com/limbara/stock-watcher/customerrors"
)

func ErrorHandlerMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err, ok := recover().(error); ok {
				var responseError *customerrors.ResponseError

				if errors.As(err, &responseError) {
					w.WriteHeader(responseError.HttpCode)
					json.NewEncoder(w).Encode(responseError)
					return
				}

				responseError = &customerrors.ResponseErrorServer

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
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
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
