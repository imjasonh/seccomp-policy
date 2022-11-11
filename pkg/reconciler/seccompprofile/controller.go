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
	"fmt"
	"log"
	"os"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	seccompprofileinformer "github.com/imjasonh/seccomp-profile/pkg/apis/injection/informers/seccomp/v1alpha1/seccompprofile"
	seccompprofilereconciler "github.com/imjasonh/seccomp-profile/pkg/apis/injection/reconciler/seccomp/v1alpha1/seccompprofile"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	// Do a quick check that we can list the local directory.
	if err := listFiles(ctx); err != nil {
		log.Fatalf("Failed to list files: %v", err)
	}

	informer := seccompprofileinformer.Get(ctx)

	r := &Reconciler{}
	impl := seccompprofilereconciler.NewImpl(ctx, r)
	informer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	return impl
}

func listFiles(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	fis, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}
	for _, fi := range fis {
		logger.Infof("- %s", fi.Name())
	}
	return nil
}
