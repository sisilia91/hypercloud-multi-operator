package controllers

import "time"

const (
	requeueAfter10Second = 10 * time.Second
	requeueAfter20Second = 20 * time.Second
	requeueAfter30Second = 30 * time.Second
	requeueAfter1Minute  = 1 * time.Minute
)

const (
	// upgrade template
	CAPI_VSPHERE_UPGRADE_TEMPLATE = "capi-vsphere-upgrade-template"

	// machine을 구분하기 위한 label key
	CAPI_CLUSTER_LABEL_KEY      = "cluster.x-k8s.io/cluster-name"
	CAPI_CONTROLPLANE_LABEL_KEY = "cluster.x-k8s.io/control-plane"
	CAPI_WORKER_LABEL_KEY       = "cluster.x-k8s.io/deployment-name"
)
