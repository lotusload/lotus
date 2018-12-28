package lotuswaiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	lotusv1beta1 "github.com/lotusload/lotus/pkg/app/lotus/apis/lotus/v1beta1"
)

func TestWaitTimedOut(t *testing.T) {
	phases := []string{
		string(lotusv1beta1.LotusFailed),
		string(lotusv1beta1.LotusSucceeded),
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)
	err := wait(ctx, time.Second, phases, zap.NewNop(), func() (*lotusv1beta1.Lotus, error) {
		return &lotusv1beta1.Lotus{
			Status: lotusv1beta1.LotusStatus{
				Phase: lotusv1beta1.LotusPending,
			},
		}, nil
	})
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestLotusReached(t *testing.T) {
	phases := map[string]struct{}{
		string(lotusv1beta1.LotusFailed):    struct{}{},
		string(lotusv1beta1.LotusSucceeded): struct{}{},
	}
	testcases := []struct {
		phases      map[string]struct{}
		getterLotus *lotusv1beta1.Lotus
		getterError error
		reached     bool
		hasError    bool
	}{
		{
			phases:      phases,
			getterError: fmt.Errorf("unknown error"),
			reached:     false,
			hasError:    true,
		},
		{
			phases:      phases,
			getterError: errors.NewNotFound(schema.GroupResource{}, "foo"),
			reached:     false,
			hasError:    false,
		},
		{
			getterLotus: &lotusv1beta1.Lotus{},
			reached:     false,
			hasError:    false,
		},
		{
			phases:      phases,
			getterLotus: &lotusv1beta1.Lotus{},
			reached:     false,
			hasError:    false,
		},
		{
			phases: phases,
			getterLotus: &lotusv1beta1.Lotus{
				Status: lotusv1beta1.LotusStatus{
					Phase: lotusv1beta1.LotusPending,
				},
			},
			reached:  false,
			hasError: false,
		},
		{
			phases: phases,
			getterLotus: &lotusv1beta1.Lotus{
				Status: lotusv1beta1.LotusStatus{
					Phase: lotusv1beta1.LotusFailed,
				},
			},
			reached:  true,
			hasError: false,
		},
		{
			phases: phases,
			getterLotus: &lotusv1beta1.Lotus{
				Status: lotusv1beta1.LotusStatus{
					Phase: lotusv1beta1.LotusSucceeded,
				},
			},
			reached:  true,
			hasError: false,
		},
	}
	logger := zap.NewNop()
	for i, tc := range testcases {
		des := fmt.Sprintf("index = %d", i)
		reached, err := lotusReached(tc.phases, logger, func() (*lotusv1beta1.Lotus, error) {
			return tc.getterLotus, tc.getterError
		})
		assert.Equal(t, tc.reached, reached, des)
		assert.Equal(t, tc.hasError, err != nil, des)
	}
}
