package json_rpc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"zktoro/clients"
	"zktoro/clients/ratelimiter"

	"github.com/rs/cors"

	"zktoro/clients/messaging"
	"zktoro/config"
	"zktoro/services/components/metrics"

	"zktoro/zktoro-core-go/clients/health"
	"zktoro/zktoro-core-go/ethereum"
	"zktoro/zktoro-core-go/protocol"
	"zktoro/zktoro-core-go/protocol/settings"
	"zktoro/zktoro-core-go/utils"
)

// JsonRpcProxy proxies requests from agents to json-rpc endpoint
type JsonRpcProxy struct {
	ctx         context.Context
	cfg         config.JsonRpcConfig
	server      *http.Server
	msgClient   clients.MessageClient
	rateLimiter ratelimiter.RateLimiter

	lastErr          health.ErrorTracker
	botAuthenticator clients.IPAuthenticator
}

func (p *JsonRpcProxy) Start() error {
	rpcUrl, err := url.Parse(p.cfg.Url)
	if err != nil {
		return err
	}
	rp := httputil.NewSingleHostReverseProxy(rpcUrl)

	d := rp.Director
	rp.Director = func(r *http.Request) {
		d(r)
		r.Host = rpcUrl.Host
		r.URL = rpcUrl
		for h, v := range p.cfg.Headers {
			r.Header.Set(h, v)
		}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	p.server = &http.Server{
		Addr:    ":8545",
		Handler: p.metricHandler(c.Handler(rp)),
	}
	utils.GoListenAndServe(p.server)

	go p.apiHealthChecker()

	return nil
}

func (p *JsonRpcProxy) metricHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t := time.Now()
		agentConfig, err := p.botAuthenticator.FindAgentFromRemoteAddr(req.RemoteAddr)
		if err == nil && p.rateLimiter.ExceedsLimit(agentConfig.ID) {
			writeTooManyReqsErr(w, req)
			p.msgClient.PublishProto(
				messaging.SubjectMetricAgent, &protocol.AgentMetricList{
					Metrics: metrics.GetJSONRPCMetrics(*agentConfig, t, 0, 1, 0),
				},
			)
			return
		}

		h.ServeHTTP(w, req)

		if err == nil {
			duration := time.Since(t)
			p.msgClient.PublishProto(
				messaging.SubjectMetricAgent, &protocol.AgentMetricList{
					Metrics: metrics.GetJSONRPCMetrics(*agentConfig, t, 1, 0, duration),
				},
			)
		}
	})
}

func (p *JsonRpcProxy) Stop() error {
	if p.server != nil {
		return p.server.Close()
	}
	return nil
}

func (p *JsonRpcProxy) Name() string {
	return "json-rpc-proxy"
}

// Health implements health.Reporter interface.
func (p *JsonRpcProxy) Health() health.Reports {
	return health.Reports{
		p.lastErr.GetReport("api"),
	}
}

func (p *JsonRpcProxy) apiHealthChecker() {
	p.testAPI()
	ticker := time.NewTicker(time.Minute * 5)
	for range ticker.C {
		p.testAPI()
	}
}

func (p *JsonRpcProxy) testAPI() {
	err := ethereum.TestAPI(p.ctx, p.cfg.Url)
	p.lastErr.Set(err)
}

func NewJsonRpcProxy(ctx context.Context, cfg config.Config) (*JsonRpcProxy, error) {
	jCfg := cfg.Scan.JsonRpc
	if len(cfg.JsonRpcProxy.JsonRpc.Url) > 0 {
		jCfg = cfg.JsonRpcProxy.JsonRpc
	}

	rateLimiting := cfg.JsonRpcProxy.RateLimitConfig
	if rateLimiting == nil {
		rateLimiting = (*config.RateLimitConfig)(settings.GetChainSettings(cfg.ChainID).JsonRpcRateLimiting)
	}

	msgClient := messaging.NewClient("json-rpc", fmt.Sprintf("%s:%s", config.DockerNatsContainerName, config.DefaultNatsPort))

	botAuthenticator, err := clients.NewBotAuthenticator(ctx)
	if err != nil {
		return nil, err
	}

	return &JsonRpcProxy{
		ctx:              ctx,
		cfg:              jCfg,
		botAuthenticator: botAuthenticator,
		msgClient:        msgClient,
		rateLimiter: ratelimiter.NewRateLimiter(
			rateLimiting.Rate,
			rateLimiting.Burst,
		),
	}, nil
}
