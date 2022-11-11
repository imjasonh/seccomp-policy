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

package seccompprofile

import (
	"context"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	seccompprofileinformer "github.com/imjasonh/seccomp-profile/pkg/apis/injection/informers/seccomp/v1alpha1/seccompprofile"
	seccompprofilereconciler "github.com/imjasonh/seccomp-profile/pkg/apis/injection/reconciler/seccomp/v1alpha1/seccompprofile"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	informer := seccompprofileinformer.Get(ctx)

	r := &Reconciler{}
	impl := seccompprofilereconciler.NewImpl(ctx, r)
	informer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	return impl
}
