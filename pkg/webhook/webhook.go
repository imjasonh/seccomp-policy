// Copyright 2022 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn/kubernetes"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	v1alpha1client "github.com/imjasonh/seccomp-profile/pkg/apis/injection/client"
	"github.com/imjasonh/seccomp-profile/pkg/apis/seccomp/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
)

type Validator struct{}

func NewValidator(ctx context.Context) *Validator {
	return &Validator{}
}

// isDeletedOrStatusUpdate returns true if the resource in question is being
// deleted, is already deleted or Status is being updated. In any of those
// cases, we do not validate the resource
func isDeletedOrStatusUpdate(ctx context.Context, deletionTimestamp *metav1.Time) bool {
	return apis.IsInDelete(ctx) || deletionTimestamp != nil || apis.IsInStatusUpdate(ctx)
}

// ResolvePodSpecable implements duckv1.PodSpecValidator
func (v *Validator) ResolvePodSpecable(ctx context.Context, wp *duckv1.WithPod) {
	// Don't mess with things that are being deleted or already deleted or
	// status update.
	if isDeletedOrStatusUpdate(ctx, wp.DeletionTimestamp) {
		return
	}

	imagePullSecrets := make([]string, 0, len(wp.Spec.Template.Spec.ImagePullSecrets))
	for _, s := range wp.Spec.Template.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	v.resolvePodSpec(ctx, &wp.Spec.Template.Spec, kubernetes.Options{
		Namespace:          getNamespace(ctx, wp.Namespace),
		ServiceAccountName: wp.Spec.Template.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	})
}

// ResolvePod implements duckv1.PodValidator
func (v *Validator) ResolvePod(ctx context.Context, p *duckv1.Pod) {
	// Don't mess with things that are being deleted or already deleted or
	// status update.
	if isDeletedOrStatusUpdate(ctx, p.DeletionTimestamp) {
		return
	}
	imagePullSecrets := make([]string, 0, len(p.Spec.ImagePullSecrets))
	for _, s := range p.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	v.resolvePodSpec(ctx, &p.Spec, kubernetes.Options{
		Namespace:          getNamespace(ctx, p.Namespace),
		ServiceAccountName: p.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	})
}

// ResolveCronJob implements duckv1.CronJobValidator
func (v *Validator) ResolveCronJob(ctx context.Context, c *duckv1.CronJob) {
	// Don't mess with things that are being deleted or already deleted or
	// status update.
	if isDeletedOrStatusUpdate(ctx, c.DeletionTimestamp) {
		return
	}

	imagePullSecrets := make([]string, 0, len(c.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets))
	for _, s := range c.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	v.resolvePodSpec(ctx, &c.Spec.JobTemplate.Spec.Template.Spec, kubernetes.Options{
		Namespace:          getNamespace(ctx, c.Namespace),
		ServiceAccountName: c.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	})
}

// For testing
var remoteResolveDigest = remote.Get

func (v *Validator) resolvePodSpec(ctx context.Context, ps *corev1.PodSpec, opt kubernetes.Options) {
	logger := logging.FromContext(ctx)

	client := kubeclient.Get(ctx)
	kc, err := kubernetes.New(ctx, client, opt)
	if err != nil {
		logger.Warnf("Unable to build keychain: %v", err)
		return
	}

	var desc *remote.Descriptor

	resolveContainers := func(cs []corev1.Container) {
		for i, c := range cs {
			ref, err := name.ParseReference(c.Image)
			if err != nil {
				logger.Debugf("Unable to parse reference: %v", err)
				continue
			}

			// If we are in the context of a mutating webhook, then resolve the tag to a digest.
			switch {
			case apis.IsInCreate(ctx), apis.IsInUpdate(ctx):
				desc, err = remoteResolveDigest(ref,
					remote.WithContext(ctx),
					remote.WithAuthFromKeychain(kc),
				)
				if err != nil {
					logger.Debugf("Unable to resolve digest %q: %v", ref.String(), err)
					continue
				}
				cs[i].Image = ref.Context().Digest(desc.Digest.String()).String()
			}
		}
	}

	resolveEphemeralContainers := func(cs []corev1.EphemeralContainer) {
		for i, c := range cs {
			ref, err := name.ParseReference(c.Image)
			if err != nil {
				logger.Debugf("Unable to parse reference: %v", err)
				continue
			}

			// If we are in the context of a mutating webhook, then resolve the tag to a digest.
			switch {
			case apis.IsInCreate(ctx), apis.IsInUpdate(ctx):
				desc, err = remoteResolveDigest(ref,
					remote.WithContext(ctx),
					remote.WithAuthFromKeychain(kc),
				)
				if err != nil {
					logger.Debugf("Unable to resolve digest %q: %v", ref.String(), err)
					continue
				}
				cs[i].Image = ref.Context().Digest(desc.Digest.String()).String()
			}
		}
	}

	resolveContainers(ps.InitContainers)
	resolveContainers(ps.Containers)
	resolveEphemeralContainers(ps.EphemeralContainers)

	// If there's only one container, and there isn't already a seccompProfile specified, try to extract the image's seccomp policy.
	if len(ps.InitContainers) == 0 &&
		len(ps.EphemeralContainers) == 0 &&
		len(ps.Containers) == 1 &&
		(ps.SecurityContext == nil || ps.SecurityContext.SeccompProfile == nil) {

		if desc == nil {
			logger.Warn("BUG: Expected descriptor to already be populated")
			return
		}

		b, err := desc.RawManifest()
		if err != nil {
			logger.Errorf("Unable to get raw manifest: %v", err)
			return
		}
		var mf struct {
			Annotations map[string]string `json:"annotations"`
		}
		if err := json.Unmarshal(b, &mf); err != nil {
			logger.Errorf("Unable to parse manifest: %v", err)
			return
		}
		v, ok := mf.Annotations["seccomp.imjasonh.dev/profile"]
		if !ok {
			logger.Infof("Image %s specified no seccomp profile", desc.Digest.String())
			return
		}
		logger.Infof("!!! Image %s specified a seccomp profile!", desc.Digest.String())

		var p v1alpha1.SeccompProfileJSON
		if err := json.Unmarshal([]byte(v), &p); err != nil {
			logger.Errorf("Image %s specified unparseable seccomp profile", desc.Digest.String())
			return
		}
		name := sha(v)
		if _, err := v1alpha1client.Get(ctx).SeccompV1alpha1().SeccompProfiles().Create(ctx, &v1alpha1.SeccompProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.SeccompProfileSpec{
				Contents: &p,
			},
		}, metav1.CreateOptions{}); k8serrors.IsAlreadyExists(err) {
			// Ignore.
		} else if err != nil {
			logger.Errorf("Error creating SeccompPolicy %q: %v", name, err)
			return
		}
		logger.Infof("Created or updated SeccompPolicy %q", name)

		if ps.SecurityContext == nil {
			ps.SecurityContext = &corev1.PodSecurityContext{}
		}
		ps.SecurityContext.SeccompProfile = &corev1.SeccompProfile{
			Type:             corev1.SeccompProfileTypeLocalhost,
			LocalhostProfile: pointer.String(fmt.Sprintf("profiles/%s.json", name)),
		}
	}
}

func sha(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// getNamespace tries to extract the namespace from the HTTPRequest
// if the namespace passed as argument is empty. This is a workaround
// for a bug in k8s <= 1.24.
func getNamespace(ctx context.Context, namespace string) string {
	logger := logging.FromContext(ctx)

	if namespace == "" {
		r := apis.GetHTTPRequest(ctx)
		if r != nil && r.Body != nil {
			var review admissionv1.AdmissionReview
			if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
				logger.Errorf("could not decode body: %v", err)
				return ""
			}
			return review.Request.Namespace
		}
	}
	return namespace
}
