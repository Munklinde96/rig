package pipeline

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service interface {
	DryRun(ctx context.Context,
		cfg *v1alpha1.OperatorConfig,
		namespace, capsuleName string,
		spec *v1alpha2.Capsule) (*pipeline.Result, error)
}

func NewService(
	cfg config.Service,
	client client.Client,
	capSvc capabilities.Service,
	logger logr.Logger,
) Service {
	return &service{
		cfg:    cfg,
		client: client,
		capSvc: capSvc,
		logger: logger,
	}
}

type service struct {
	cfg    config.Service
	client client.Client
	capSvc capabilities.Service
	logger logr.Logger
}

// Get implements Service.
func (s *service) DryRun(
	ctx context.Context,
	cfg *v1alpha1.OperatorConfig,
	namespace, capsuleName string,
	spec *v1alpha2.Capsule,
) (*pipeline.Result, error) {
	if cfg == nil {
		cfg = s.cfg.Operator()
	}
	if spec == nil {
		spec = &v1alpha2.Capsule{}
		if err := s.client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      capsuleName,
		}, spec); err != nil {
			return nil, err
		}
	}

	steps, err := controller.GetDefaultPipelineSteps(ctx, s.capSvc)
	if err != nil {
		return nil, err
	}

	p := pipeline.New(s.client, cfg, scheme.New(), s.logger)
	for _, step := range steps {
		p.AddStep(step)
	}

	for _, step := range cfg.Steps {
		ps, err := plugin.NewStep(step, s.logger)
		if err != nil {
			return nil, err
		}

		p.AddStep(ps)
		defer ps.Stop(ctx)
	}

	return p.RunCapsule(ctx, spec, true)
}