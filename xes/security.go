package xes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/utils/xlog"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
)

type Security interface {
	Intercept(ctx context.Context, req *http.Request) error
	Middleware(required bool) func(handler http.Handler) http.Handler
	Anonymous() func(handler http.Handler) http.Handler
}

type security struct {
	tokenAuth *jwtauth.JWTAuth
}

func (s *security) Intercept(ctx context.Context, req *http.Request) error {
	actor := es.GetActor(ctx)
	if actor == nil {
		return nil
	}

	claims := map[string]interface{}{
		"actor_id":   actor.Id.String(),
		"actor_type": actor.Type,
	}
	_, token, err := s.tokenAuth.Encode(claims)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return nil
}

func (s *security) Middleware(required bool) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := xlog.Logger(ctx)

			token, err := jwtauth.VerifyRequest(s.tokenAuth, r, jwtauth.TokenFromHeader)

			// go next if no token found and not required
			if err == jwtauth.ErrNoTokenFound && !required {
				next.ServeHTTP(w, r)
				return
			}

			if err != nil {
				log.Error("failed to verify token", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			// parse it.
			claims, err := token.AsMap(ctx)
			if err != nil {
				log.Error("failed to get claims", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			actorId, actorIdOk := claims["actor_id"]
			actorType, actorTypeOk := claims["actor_type"]
			if !actorTypeOk {
				log.Error("actor_type not found in claims")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			actorTypeStr, ok := actorType.(string)
			if !ok {
				log.Error("failed to cast actor_type to string")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			actor := &es.Actor{
				Type: actorTypeStr,
			}

			if actorIdOk {
				str, ok := actorId.(string)
				if !ok {
					log.Error("failed to cast actor_id to string")
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}

				id, err := uuid.Parse(str)
				if err != nil {
					log.Error("failed to parse actor_id", zap.Error(err))
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}
				actor.Id = id
			}

			ctx = es.SetActor(ctx, actor)

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *security) Anonymous() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = es.SetActor(ctx, &es.Actor{
				Id:   uuid.New(),
				Type: "Anonymous",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func NewSecurity(signKey string) (Security, error) {
	tokenAuth := jwtauth.New("HS256", []byte(signKey), nil, jwt.WithAcceptableSkew(30*time.Second))

	return &security{
		tokenAuth: tokenAuth,
	}, nil
}
