package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	"alongtu/backend/internal/auth"
	"alongtu/backend/internal/navigation"
	"alongtu/backend/internal/places"
	"alongtu/backend/internal/platform/config"
	"alongtu/backend/internal/platform/cursor"
	"alongtu/backend/internal/platform/httpx"
	"alongtu/backend/internal/platform/observability"
	"alongtu/backend/internal/providers/amap"
	"alongtu/backend/internal/providers/resilient"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

type App struct {
	server      *http.Server
	adminServer *http.Server
	db          *pgxpool.Pool
	redis       *redis.Client
}

const startupDependencyTimeout = 5 * time.Second

func New(ctx context.Context, c config.Config, log *slog.Logger) (*App, error) {
	startupCtx, cancelStartup := context.WithTimeout(ctx, startupDependencyTimeout)
	defer cancelStartup()
	db, err := pgxpool.New(ctx, c.DatabaseURL)
	if err != nil {
		return nil, err
	}
	metrics := observability.NewMetrics()
	if err = metrics.RegisterDatabasePool(db); err != nil {
		db.Close()
		return nil, err
	}
	if err = db.Ping(startupCtx); err != nil {
		db.Close()
		return nil, err
	}
	opt, err := redis.ParseURL(c.RedisURL)
	if err != nil {
		db.Close()
		return nil, err
	}
	rc := redis.NewClient(opt)
	rc.AddHook(metrics.RedisHook())
	if err = rc.Ping(startupCtx).Err(); err != nil {
		db.Close()
		_ = rc.Close()
		return nil, err
	}
	var sender auth.Sender = auth.DisabledSender{}
	if c.Environment == "development" {
		sender = auth.NewDevelopmentSender(log)
	} else if c.Environment == "production" {
		sender, err = auth.NewAliyunSender(c.AliyunAccessKeyID, c.AliyunAccessKeySecret, c.AliyunSMSSignName, c.AliyunSMSTemplateCode)
		if err != nil {
			db.Close()
			_ = rc.Close()
			return nil, err
		}
	}
	authHandler := auth.NewHandler(auth.NewService(auth.NewPostgresRepository(db), rc, sender, c.JWTSecret, c.AccessTTL, c.RefreshTTL, metrics))
	placeRepo := places.NewPostgresRepository(db)
	placeProvider := amap.New(c.AmapBaseURL, c.AmapKey, c.ProviderTimeout)
	resilientProvider := resilient.NewWithTimeout(placeProvider, rc, c.ProviderTimeout, metrics)
	placeHandler := places.NewHandler(places.NewService(placeRepo, resilientProvider, cursor.New(c.CursorSecret, 15*time.Minute), metrics))
	navigationHandler := navigation.NewHandler(navigation.NewService(resilientProvider, c.RideHailingURLTemplate))
	trusted, err := trustedPrefixes(c.TrustedProxyCIDRs)
	if err != nil {
		db.Close()
		_ = rc.Close()
		return nil, err
	}
	ready := func(checkCtx context.Context) error {
		return errors.Join(db.Ping(checkCtx), rc.Ping(checkCtx).Err())
	}
	r := newHTTPHandler(log, trusted, c.AllowedOrigins, authHandler, placeHandler, navigationHandler, metrics, ready)
	server := &http.Server{Addr: c.HTTPAddr, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	adminMux := http.NewServeMux()
	adminMux.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))
	adminServer := &http.Server{Addr: c.AdminAddr, Handler: adminMux, ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	return &App{server: server, adminServer: adminServer, db: db, redis: rc}, nil
}

func newHTTPHandler(log *slog.Logger, trusted []netip.Prefix, allowedOrigins []string, authHandler *auth.Handler, placeHandler *places.Handler, navigationHandler *navigation.Handler, metrics *observability.Metrics, ready func(context.Context) error) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return httpx.Recover(log, next) })
	r.Use(func(next http.Handler) http.Handler { return httpx.RequestContext(log, next) })
	if metrics != nil {
		r.Use(metrics.HTTPMiddleware)
	}
	r.Use(func(next http.Handler) http.Handler { return httpx.ClientIPMiddleware(trusted, next) })
	r.Use(func(next http.Handler) http.Handler { return httpx.CORS(allowedOrigins, next) })
	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) { httpx.JSON(w, 200, map[string]string{"status": "ok"}) })
	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		checkCtx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if ready == nil || ready(checkCtx) != nil {
			httpx.Fail(w, r, 503, "not_ready", "dependencies unavailable", nil)
			return
		}
		httpx.JSON(w, 200, map[string]string{"status": "ready"})
	})
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/sms/send", authHandler.Send)
		api.Post("/auth/sms/verify", authHandler.Verify)
		api.Post("/auth/refresh", authHandler.Refresh)
		api.Post("/auth/logout", authHandler.Logout)
		api.Group(func(scoped chi.Router) {
			scoped.Use(authHandler.Optional)
			scoped.Mount("/places", placeHandler.Routes())
			scoped.Post("/navigation/routes", navigationHandler.Route)
			scoped.Get("/navigation/external-links", navigationHandler.ExternalLinks)
		})
	})
	return r
}
func (a *App) Run(ctx context.Context, shutdownTimeout time.Duration) error {
	errCh := make(chan error, 2)
	go func() { errCh <- a.server.ListenAndServe() }()
	go func() { errCh <- a.adminServer.ListenAndServe() }()
	var cause error
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			cause = err
		}
	case <-ctx.Done():
		cause = nil
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		apiErr := a.server.Shutdown(shutdownCtx)
		adminErr := a.adminServer.Shutdown(shutdownCtx)
		return errors.Join(cause, apiErr, adminErr)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	apiErr := a.server.Shutdown(shutdownCtx)
	adminErr := a.adminServer.Shutdown(shutdownCtx)
	return errors.Join(cause, apiErr, adminErr)
}
func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
	if a.redis != nil {
		_ = a.redis.Close()
	}
}
func trustedPrefixes(values []string) ([]netip.Prefix, error) {
	out := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		p, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
