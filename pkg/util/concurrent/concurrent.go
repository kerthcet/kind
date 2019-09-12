/*
Copyright 2019 The Kubernetes Authors.

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

// Package concurrent provides utilities for concurrent execution
// TODO(bentheelder): this should be internal
package concurrent

import (
	"sync"

	"sigs.k8s.io/kind/pkg/errors"
)

// UntilError runs all funcs in separate goroutines, returning the
// first non-nil error returned from funcs, or nil if all funcs return nil
func UntilError(funcs []func() error) error {
	errCh := make(chan error, len(funcs))
	for _, f := range funcs {
		f := f // capture f
		go func() {
			errCh <- f()
		}()
	}
	for i := 0; i < len(funcs); i++ {
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}

// Coalesce runs fns concurrently, returning an Errors if there are > 1 errors
func Coalesce(fns ...func() error) error {
	// run all fns concurrently
	ch := make(chan error, len(fns))
	var wg sync.WaitGroup
	for _, fn := range fns {
		wg.Add(1)
		go func(f func() error) {
			defer wg.Done()
			ch <- f()
		}(fn)
	}
	wg.Wait()
	close(ch)
	// collect up and return errors
	errs := []error{}
	for err := range ch {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 1 {
		return errors.NewAggregate(errs)
	} else if len(errs) == 1 {
		return errs[0]
	}
	return nil
}
