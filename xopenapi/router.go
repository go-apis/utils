package xopenapi

import (
	"net/http"
	"reflect"

	"github.com/contextcloud/goutils/xservice"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest/response/gzip"
	"github.com/swaggest/rest/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/swaggest/swgui/v5emb"
)

type HttpMiddleware = func(http.Handler) http.Handler

func New(cfg *xservice.ServiceConfig, middlewares ...HttpMiddleware) *web.Service {
	uuidDef := jsonschema.Schema{}
	uuidDef.AddType(jsonschema.String)
	uuidDef.WithFormat("uuid")
	uuidDef.WithExamples("248df4b7-aa70-47b8-a036-33ac447e668d")

	reflector := openapi3.NewReflector()
	reflector.AddTypeMapping(uuid.UUID{}, uuidDef)
	reflector.InlineDefinition(uuid.UUID{})
	reflector.DefaultOptions = append(reflector.DefaultOptions, jsonschema.InterceptNullability(func(params jsonschema.InterceptNullabilityParams) {
		if params.Type.Kind() != reflect.Ptr && params.Schema.HasType(jsonschema.Null) && params.Schema.HasType(jsonschema.Array) {
			*params.Schema.Type = jsonschema.Array.Type()
		}
	}))

	s := web.NewService(reflector)
	s.OpenAPISchema().SetTitle(cfg.Service)
	s.OpenAPISchema().SetVersion(cfg.Version)

	spanNameFormatter := func(operation string, r *http.Request) string {
		rctx := chi.NewRouteContext()
		if s.Match(rctx, r.Method, r.URL.Path) {
			return rctx.RoutePattern()
		}
		return operation
	}

	// Setup middlewares.
	s.Wrap(
		gzip.Middleware,
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		otelhttp.NewMiddleware(cfg.Service, otelhttp.WithSpanNameFormatter(spanNameFormatter)),
	)
	s.Wrap(
		middlewares...,
	)

	s.Docs("/docs", v5emb.New)
	s.Mount("/debug", middleware.Profiler())

	return s
}
