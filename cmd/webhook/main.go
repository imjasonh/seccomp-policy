/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/configmaps"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"

	"github.com/imjasonh/seccomp-profile/pkg/apis/seccomp/v1alpha1"
	pwebhook "github.com/imjasonh/seccomp-profile/pkg/webhook"
)

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	// List the types to validate.
	v1alpha1.SchemeGroupVersion.WithKind("SeccompProfile"): &v1alpha1.SeccompProfile{},
}

var callbacks = map[schema.GroupVersionKind]validation.Callback{}

func NewDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return defaulting.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"defaulting.seccomp.imjasonh.dev",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			// Here is where you would infuse the context with state
			// (e.g. attach a store with configmap data)
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"validation.seccomp.imjasonh.dev",

		// The path on which to serve the webhook.
		"/resource-validation",

		// The resources to validate.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			// Here is where you would infuse the context with state
			// (e.g. attach a store with configmap data)
			return ctx
		},

		// Whether to disallow unknown fields.
		true,

		// Extra validating callbacks to be applied to resources.
		callbacks,
	)
}

func NewConfigValidationController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return configmaps.NewAdmissionController(ctx,

		// Name of the configmap webhook.
		"config.seccomp.imjasonh.dev",

		// The path on which to serve the webhook.
		"/config-validation",

		// The configmaps to validate.
		configmap.Constructors{
			logging.ConfigMapName(): logging.NewConfigFromConfigMap,
			metrics.ConfigMapName(): metrics.NewObservabilityConfigFromConfigMap,
		},
	)
}

var (
	_ resourcesemantics.SubResourceLimited = (*crdNoStatusUpdatesOrDeletes)(nil)
	_ resourcesemantics.VerbLimited        = (*crdNoStatusUpdatesOrDeletes)(nil)

	_ resourcesemantics.SubResourceLimited = (*crdEphemeralContainers)(nil)
	_ resourcesemantics.VerbLimited        = (*crdEphemeralContainers)(nil)
)

type crdNoStatusUpdatesOrDeletes struct {
	resourcesemantics.GenericCRD
}

type crdEphemeralContainers struct {
	resourcesemantics.GenericCRD
}

func (c *crdNoStatusUpdatesOrDeletes) SupportedSubResources() []string {
	// We do not want any updates that are for status, scale, or anything else.
	return []string{""}
}

func (c *crdEphemeralContainers) SupportedSubResources() []string {
	return []string{"/ephemeralcontainers", ""}
}

func (c *crdNoStatusUpdatesOrDeletes) SupportedVerbs() []admissionregistrationv1.OperationType {
	return []admissionregistrationv1.OperationType{
		admissionregistrationv1.Create,
		admissionregistrationv1.Update,
	}
}

func (c *crdEphemeralContainers) SupportedVerbs() []admissionregistrationv1.OperationType {
	return []admissionregistrationv1.OperationType{
		admissionregistrationv1.Create,
		admissionregistrationv1.Update,
	}
}

func NewMutatingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	kc := kubeclient.Get(ctx)
	validator := pwebhook.NewValidator(ctx)

	return defaulting.NewAdmissionController(ctx,
		// Name of the resource webhook.
		"mutating.seccomp.imjasonh.dev",

		// The path on which to serve the webhook.
		"/mutations",

		// The resources to validate.
		map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
			corev1.SchemeGroupVersion.WithKind("Pod"):           &crdEphemeralContainers{GenericCRD: &duckv1.Pod{}},
			appsv1.SchemeGroupVersion.WithKind("ReplicaSet"):    &crdNoStatusUpdatesOrDeletes{GenericCRD: &duckv1.WithPod{}},
			appsv1.SchemeGroupVersion.WithKind("Deployment"):    &crdNoStatusUpdatesOrDeletes{GenericCRD: &duckv1.WithPod{}},
			appsv1.SchemeGroupVersion.WithKind("StatefulSet"):   &crdNoStatusUpdatesOrDeletes{GenericCRD: &duckv1.WithPod{}},
			appsv1.SchemeGroupVersion.WithKind("DaemonSet"):     &crdNoStatusUpdatesOrDeletes{GenericCRD: &duckv1.WithPod{}},
			batchv1.SchemeGroupVersion.WithKind("Job"):          &crdNoStatusUpdatesOrDeletes{GenericCRD: &duckv1.WithPod{}},
			batchv1.SchemeGroupVersion.WithKind("CronJob"):      &crdNoStatusUpdatesOrDeletes{GenericCRD: &duckv1.CronJob{}},
			batchv1beta1.SchemeGroupVersion.WithKind("CronJob"): &crdNoStatusUpdatesOrDeletes{GenericCRD: &duckv1.CronJob{}},
		},

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			ctx = context.WithValue(ctx, kubeclient.Key{}, kc)
			ctx = duckv1.WithPodDefaulter(ctx, validator.ResolvePod)
			ctx = duckv1.WithPodSpecDefaulter(ctx, validator.ResolvePodSpecable)
			ctx = duckv1.WithCronJobDefaulter(ctx, validator.ResolveCronJob)
			return ctx
		},

		// Whether to disallow unknown fields.
		// We pass false because we're using partial schemas.
		false,
	)
}

func main() {
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "webhook",
		Port:        8443,
		SecretName:  "webhook-certs",
	})

	sharedmain.WebhookMainWithContext(ctx, "webhook",
		certificates.NewController,
		NewDefaultingAdmissionController,
		NewValidationAdmissionController,
		NewConfigValidationController,
		NewMutatingAdmissionController,
	)
}
