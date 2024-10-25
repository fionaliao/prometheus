// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tsdb

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb/chunks"
)

func BenchmarkHeadStripeSeriesCreate(b *testing.B) {
	chunkDir := b.TempDir()
	// Put a series, select it. GC it and then access it.
	opts := DefaultHeadOptions()
	opts.ChunkRange = 1000
	opts.ChunkDirRoot = chunkDir
	h, err := NewHead(nil, nil, nil, nil, opts, nil)
	require.NoError(b, err)
	defer h.Close()

	for i := 0; i < b.N; i++ {
		h.getOrCreate(uint64(i), labels.FromStrings(labels.MetricName, "test", "a", strconv.Itoa(i), "b", strconv.Itoa(i%10), "c", strconv.Itoa(i%100), "d", strconv.Itoa(i/2), "e", strconv.Itoa(i/4)))
	}
}

func BenchmarkHeadStripeSeriesCreateParallel(b *testing.B) {
	chunkDir := b.TempDir()
	// Put a series, select it. GC it and then access it.
	opts := DefaultHeadOptions()
	opts.ChunkRange = 1000
	opts.ChunkDirRoot = chunkDir
	h, err := NewHead(nil, nil, nil, nil, opts, nil)
	require.NoError(b, err)
	defer h.Close()

	var count atomic.Int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(count.Inc())
			h.getOrCreate(uint64(i), labels.FromStrings(labels.MetricName, "test", "a", strconv.Itoa(i), "b", strconv.Itoa(i%10), "c", strconv.Itoa(i%100), "d", strconv.Itoa(i/2), "e", strconv.Itoa(i/4)))
		}
	})
}

func BenchmarkHeadStripeSeriesCreate_PreCreationFailure(b *testing.B) {
	chunkDir := b.TempDir()
	// Put a series, select it. GC it and then access it.
	opts := DefaultHeadOptions()
	opts.ChunkRange = 1000
	opts.ChunkDirRoot = chunkDir

	// Mock the PreCreation() callback to fail on each series.
	opts.SeriesCallback = failingSeriesLifecycleCallback{}

	h, err := NewHead(nil, nil, nil, nil, opts, nil)
	require.NoError(b, err)
	defer h.Close()

	for i := 0; i < b.N; i++ {
		h.getOrCreate(uint64(i), labels.FromStrings(labels.MetricName, "test", "a", strconv.Itoa(i), "b", strconv.Itoa(i%10), "c", strconv.Itoa(i%100), "d", strconv.Itoa(i/2), "e", strconv.Itoa(i/4)))
	}
}

type failingSeriesLifecycleCallback struct{}

func (failingSeriesLifecycleCallback) PreCreation(labels.Labels) error                     { return errors.New("failed") }
func (failingSeriesLifecycleCallback) PostCreation(labels.Labels)                          {}
func (failingSeriesLifecycleCallback) PostDeletion(map[chunks.HeadSeriesRef]labels.Labels) {}
