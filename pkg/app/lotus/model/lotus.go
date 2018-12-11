package model

import (
	lotusv1beta1 "github.com/nghialv/lotus/pkg/app/lotus/apis/lotus/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	LotusKind = "Lotus"
)

var (
	ControllerKind = schema.GroupVersionKind{
		Group:   lotusv1beta1.SchemeGroupVersion.Group,
		Version: lotusv1beta1.SchemeGroupVersion.Version,
		Kind:    LotusKind,
	}
)
