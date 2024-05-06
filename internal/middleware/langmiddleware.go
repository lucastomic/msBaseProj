package middleware

import (
	"context"
	"net/http"

	"github.com/lucastomic/msBaseProj/internal/contextypes"
)

type LangMideware struct{}

func NewLangMiddleware() Middleware {
	return LangMideware{}
}

func (LangMideware) Execute(next http.HandlerFunc, errHandler errorHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		langHeader := r.Header.Get("Accept-Language")
		ctx := r.Context()
		ctx = context.WithValue(ctx, contextypes.ContextLangKey{}, langHeader)
		*r = *r.WithContext(ctx)
		next(w, r)
	})
}
