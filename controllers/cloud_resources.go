/*
Copyright 2022.

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

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/nstogner/gcp-ilb-redirect-controller/cloud"
	"github.com/nstogner/gcp-ilb-redirect-controller/cloudresources"

	clog "sigs.k8s.io/controller-runtime/pkg/log"
)

type cloudResource interface {
	Get(context.Context, cloudresources.Info) error
	Create(context.Context, cloudresources.Info) error
	Delete(context.Context, cloudresources.Info) error
	// TODO: Update/Patch (and possibly Delete followed by recreate) for resources that do not match.
}

var errNotReady = errors.New("not ready")

func remove(ctx context.Context, info cloudresources.Info, c cloudResource) error {
	log := clog.FromContext(ctx).WithValues("generatedName", info.GeneratedName, "resourceType", fmt.Sprintf("%T", c))

	log.Info("Deleting resource")
	err := c.Delete(ctx, info)
	if err != nil {
		if cloud.IsNotFound(err) {
			return nil
		}
		if cloud.IsNotReady(err) {
			return errNotReady
		}
		return err
	}

	return nil
}

func ensure(ctx context.Context, info cloudresources.Info, c cloudResource) error {
	log := clog.FromContext(ctx).WithValues("generatedName", info.GeneratedName, "resourceType", fmt.Sprintf("%T", c))

	err := c.Get(ctx, info)
	if err != nil {
		if cloud.IsNotFound(err) {
			log.Info("Creating resource")
			if err := c.Create(ctx, info); err != nil {
				if cloud.IsNotReady(err) {
					return errNotReady
				}

				return fmt.Errorf("create: %w", err)
			} else {
				return nil
			}
		}
		return fmt.Errorf("get: %w", err)
	}

	return nil
}
