package lotuswaiter

import (
	"context"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"

	lotusv1beta1 "github.com/lotusload/lotus/pkg/app/lotus/apis/lotus/v1beta1"
)

type LotusGetter func() (*lotusv1beta1.Lotus, error)

func wait(ctx context.Context, probePeriod time.Duration, phases []string, logger *zap.Logger, getter LotusGetter) error {
	phaseMap := make(map[string]struct{}, len(phases))
	for _, p := range phases {
		phaseMap[p] = struct{}{}
	}
	ticker := time.NewTicker(probePeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			reached, err := lotusReached(phaseMap, logger, getter)
			if err != nil {
				return err
			}
			if reached {
				logger.Info("lotus is reached")
				return nil
			}
		case <-ctx.Done():
			if err := ctx.Err(); err == context.DeadlineExceeded {
				logger.Error("timed out before waiting lotus finishes", zap.Error(err))
				return err
			}
			return nil
		}
	}
	return nil
}

func lotusReached(phases map[string]struct{}, logger *zap.Logger, getter LotusGetter) (bool, error) {
	logger.Info("start checking lotus")
	lotus, err := getter()
	if errors.IsNotFound(err) {
		logger.Info("lotus is not found, still waiting", zap.Error(err))
		return false, nil
	}
	if err != nil {
		logger.Error("failed to check lotus status", zap.Error(err))
		return false, err
	}
	phase := string(lotus.Status.Phase)
	logger.Info("lotus is found", zap.String("phase", phase))
	_, ok := phases[phase]
	return ok, nil
}
