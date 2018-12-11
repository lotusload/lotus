package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Lotus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LotusSpec   `json:"spec"`
	Status LotusStatus `json:"status"`
}

type LotusSpec struct {
	TTLSecondsAfterFinished  *int32 `json:"ttlSecondsAfterFinished"`
	CheckIntervalSeconds     *int32 `json:"checkIntervalSeconds"`
	CheckInitialDelaySeconds *int32 `json:"checkInitialDelaySeconds"`

	Preparer *LotusSpecPreparer `json:"preparer"`
	Worker   *LotusSpecWorker   `json:"worker"`
	Cleaner  *LotusSpecCleaner  `json:"cleaner"`
	Checks   []LotusCheck       `json:"checks"`
}

type LotusSpecWorker struct {
	RunTime     string             `json:"runTime"`
	Replicas    *int32             `json:"replicas"`
	MetricsPort *int32             `json:"metricsPort"`
	Containers  []corev1.Container `json:"containers"`
	Volumes     []corev1.Volume    `json:"volumes"`
}

type LotusSpecPreparer struct {
	Containers []corev1.Container `json:"containers"`
	Volumes    []corev1.Volume    `json:"volumes"`
}

type LotusSpecCleaner struct {
	Containers []corev1.Container `json:"containers"`
	Volumes    []corev1.Volume    `json:"volumes"`
}

type LotusCheck struct {
	Name       string `json:"name"`
	Expr       string `json:"expr"`
	For        string `json:"for"`
	DataSource string `json:"dataSource"`
}

type LotusPhase string

const (
	LotusInit            LotusPhase = ""
	LotusPending                    = "Pending"
	LotusPreparing                  = "Preparing"
	LotusRunning                    = "Running"
	LotusCleaning                   = "Cleaning"
	LotusFailureCleaning            = "FailureCleaning"
	LotusSucceeded                  = "Succeeded"
	LotusFailed                     = "Failed"
)

type LotusStatus struct {
	PreparerStartTime      *metav1.Time `json:"preparerStartTime"`
	PreparerCompletionTime *metav1.Time `json:"preparerCompletionTime"`
	WorkerStartTime        *metav1.Time `json:"workerStartTime"`
	WorkerCompletionTime   *metav1.Time `json:"workerCompletionTime"`
	CleanerStartTime       *metav1.Time `json:"cleanerStartTime"`
	CleanerCompletionTime  *metav1.Time `json:"cleanerCompletionTime"`
	Phase                  LotusPhase   `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type LotusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Lotus `json:"items"`
}
