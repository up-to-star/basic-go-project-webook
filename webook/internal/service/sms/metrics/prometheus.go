package metrics

import (
	"context"
	"github.com/basic-go-project-webook/webook/internal/service/sms"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc sms.Service) sms.Service {
	vector := prometheus.NewSummaryVec(prometheus.
		SummaryOpts{
		Namespace: "study",
		Subsystem: "webook_sms",
		Name:      "sms_resp_time",
		Help:      "统计 SMS 服务的性能数据",
	}, []string{"biz"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		svc:    svc,
		vector: vector,
	}
}

func (p *PrometheusDecorator) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		p.vector.WithLabelValues(biz).Observe(float64(time.Since(startTime).Milliseconds()))
	}()
	return p.svc.Send(ctx, biz, args, numbers...)
}
