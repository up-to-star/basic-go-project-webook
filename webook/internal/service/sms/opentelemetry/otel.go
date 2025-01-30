package opentelemetry

import (
	"basic-project/webook/internal/service/sms"
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	svc    sms.Service
	tracer trace.Tracer
}

func NewService(svc sms.Service) sms.Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("webook/internal/service/sms/opentelemetry")
	return &Service{
		svc:    svc,
		tracer: tracer,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	ctx, span := s.tracer.Start(ctx, "sms_send"+tplId, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	span.AddEvent("发送短信")
	return s.svc.Send(ctx, tplId, args, numbers...)
}
