package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/metrics"
)

// metricsRouter mounts the Prometheus endpoint.
//
// Any signed-in role may read it: it holds no secret. Prometheus authenticates
// with an API key in its authorization block, which sends
// Authorization: Bearer <key>, and CheckToken already accepts that.
//
//	scrape_configs:
//	  - job_name: ironkvm
//	    scheme: https
//	    metrics_path: /api/metrics
//	    authorization:
//	      credentials: <api key>
//	    static_configs:
//	      - targets: ['<kvm>']
func metricsRouter(r *gin.Engine) {
	r.GET("/api/metrics", middleware.CheckToken(), metrics.Handler)
}
