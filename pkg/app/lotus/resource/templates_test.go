package resource

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderTemplate(t *testing.T) {
	params := &prometheusConfigParams{
		Namespace:   "default",
		ServiceName: "foo",
		RuleFiles: []string{
			"rule-file-1.yaml",
			"rule-file-2.yaml",
		},
	}
	cfg, err := renderTemplate(params, prometheusConfigTemplate)
	require.NoError(t, err)
	fmt.Println(string(cfg))
}
