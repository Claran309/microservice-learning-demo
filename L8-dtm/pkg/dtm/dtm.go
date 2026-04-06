package dtm

import (
	"context"
	"errors"
	"fmt"

	"github.com/dtm-labs/dtmcli"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type DTMManager struct {
	dtmServer string
}

func NewDTMManager(dtmServer string) *DTMManager {
	zap.L().Info("√ 初始化DTM管理器成功",
		zap.String("dtm_server", dtmServer),
		zap.String("component", "dtm"),
	)
	return &DTMManager{
		dtmServer: dtmServer,
	}
}

type SagaStep struct {
	Action     string
	Compensate string
	Data       interface{}
}

func (d *DTMManager) NewSaga() *SagaBuilder {
	return &SagaBuilder{
		dtmServer: d.dtmServer,
		gid:       dtmcli.MustGenGid(d.dtmServer),
		steps:     make([]SagaStep, 0),
	}
}

type SagaBuilder struct {
	dtmServer string
	gid       string
	steps     []SagaStep
}

func (s *SagaBuilder) AddStep(actionURL string, compensateURL string, payload interface{}) *SagaBuilder {
	s.steps = append(s.steps, SagaStep{
		Action:     actionURL,
		Compensate: compensateURL,
		Data:       payload,
	})
	return s
}

func (s *SagaBuilder) WithGid(gid string) *SagaBuilder {
	s.gid = gid
	return s
}

func (s *SagaBuilder) GetGid() string {
	return s.gid
}

func (s *SagaBuilder) Submit(ctx context.Context) error {
	tracer := otel.Tracer("dtm")
	ctx, span := tracer.Start(ctx, "dtm.SagaSubmit")
	defer span.End()

	span.SetAttributes(
		attribute.String("dtm.gid", s.gid),
		attribute.Int("dtm.steps", len(s.steps)),
	)

	saga := dtmcli.NewSaga(s.dtmServer, s.gid)
	for _, step := range s.steps {
		saga.Add(step.Action, step.Compensate, step.Data)
	}

	zap.L().Info("开始提交DTM SAGA事务",
		zap.String("gid", s.gid),
		zap.Int("steps", len(s.steps)),
		zap.String("component", "dtm"),
	)

	err := saga.Submit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zap.L().Error("× DTM SAGA事务提交失败",
			zap.Error(err),
			zap.String("gid", s.gid),
			zap.String("component", "dtm"),
		)
		return errors.New("DTM SAGA事务提交失败: " + err.Error())
	}

	zap.L().Info("√ DTM SAGA事务提交成功",
		zap.String("gid", s.gid),
		zap.String("component", "dtm"),
	)

	return nil
}

func (d *DTMManager) QueryStatus(ctx context.Context, gid string) (string, error) {
	tracer := otel.Tracer("dtm")
	_, span := tracer.Start(ctx, "dtm.QueryStatus")
	defer span.End()

	span.SetAttributes(attribute.String("dtm.gid", gid))

	zap.L().Info("√ 查询DTM事务状态",
		zap.String("gid", gid),
		zap.String("component", "dtm"),
	)

	return "submitted", nil
}

func BuildURL(host string, port int, path string) string {
	return fmt.Sprintf("http://%s:%d%s", host, port, path)
}

func GetDTMGid(dtmServer string) string {
	return dtmcli.MustGenGid(dtmServer)
}
