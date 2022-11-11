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

package v1alpha1

import (
	"context"
	"fmt"

	"knative.dev/pkg/apis"
)

// Validate implements apis.Validatable
func (sp *SeccompProfile) Validate(ctx context.Context) *apis.FieldError {
	return sp.Spec.Validate(ctx).ViaField("spec")
}

type Action string

const (
	ActionLog   Action = "SCMP_ACT_LOG"
	ActionErr   Action = "SCMP_ACT_ERRNO"
	ActionAllow Action = "SCMP_ACT_ALLOW"
)

func (a Action) Valid() error {
	switch a {
	case ActionLog, ActionErr, ActionAllow:
		return nil
	default:
		return fmt.Errorf("unknown action: %s", a)
	}
}

// Validate implements apis.Validatable
func (spec *SeccompProfileSpec) Validate(ctx context.Context) *apis.FieldError {
	if spec.Contents == nil {
		return apis.ErrMissingField("contents")
	}

	if err := spec.Contents.DefaultAction.Valid(); err != nil {
		return apis.ErrInvalidValue(spec.Contents, "contents.defaultAction", fmt.Sprintf("invalid default action: %v", err))
	}
	for i, s := range spec.Contents.Syscalls {
		if err := s.Action.Valid(); err != nil {
			return apis.ErrInvalidValue(spec.Contents, "contents.syscalls.action", fmt.Sprintf("item %d: invalid action: %v", i, err))
		}
		if s.Name != "" && len(s.Names) != 0 {
			return apis.ErrInvalidValue(spec.Contents, "contents.syscalls", fmt.Sprintf("item %d: cannot specify both .name and .names", i))
		}
	}

	return nil
}
