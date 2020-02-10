/*
Copyright 2017 The Nuclio Authors.

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

package trigger

import (
	"strconv"
	"time"

	"github.com/nuclio/nuclio/pkg/errors"
	"github.com/nuclio/nuclio/pkg/functionconfig"
	"github.com/nuclio/nuclio/pkg/processor/runtime"
	"github.com/nuclio/nuclio/pkg/processor/worker"
)

type DurationConfigField struct {
	Name    string
	Value   string
	Field   *time.Duration
	Default time.Duration
}

type AnnotationConfigField struct {
	Key         string
	ValueString *string
	ValueInt    *int
}

type Configuration struct {
	functionconfig.Trigger

	// the runtime configuration, for reference
	RuntimeConfiguration *runtime.Configuration

	// a unique trigger ID
	ID string
}

func NewConfiguration(ID string,
	triggerConfiguration *functionconfig.Trigger,
	runtimeConfiguration *runtime.Configuration) *Configuration {

	configuration := &Configuration{
		Trigger:              *triggerConfiguration,
		RuntimeConfiguration: runtimeConfiguration,
		ID:                   ID,
	}

	// set defaults
	if configuration.MaxWorkers == 0 {
		configuration.MaxWorkers = 1
	}

	return configuration
}

// allows setting configuration via annotations, for experimental settings
func (c *Configuration) PopulateConfigurationFromAnnotations(annotationConfigFields []AnnotationConfigField) error {
	var err error

	for _, annotationConfigField := range annotationConfigFields {
		if annotationValue, annotationKeyExists := c.RuntimeConfiguration.Config.Meta.Annotations[annotationConfigField.Key]; annotationKeyExists {
			if annotationConfigField.ValueString != nil {
				*annotationConfigField.ValueString = annotationValue
			} else if annotationConfigField.ValueInt != nil {
				*annotationConfigField.ValueInt, err = strconv.Atoi(annotationValue)
				if err != nil {
					return errors.Wrapf(err, "Annotation %s must be numeric", annotationConfigField.Key)
				}
			}
		}
	}

	return nil
}

// parses a duration string into a time.duration field. if empty, sets the field to the default
func (c *Configuration) ParseDurationOrDefault(durationConfigField *DurationConfigField) error {
	if durationConfigField.Value == "" {
		*durationConfigField.Field = durationConfigField.Default
		return nil
	}

	parsedDurationValue, err := time.ParseDuration(durationConfigField.Value)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse %s", durationConfigField.Name)
	}

	*durationConfigField.Field = parsedDurationValue

	return nil
}

type Statistics struct {
	EventsHandleSuccessTotal  uint64
	EventsHandleFailureTotal  uint64
	WorkerAllocatorStatistics worker.AllocatorStatistics
}

func (s *Statistics) DiffFrom(prev *Statistics) Statistics {
	workerAllocatorStatisticsDiff := s.WorkerAllocatorStatistics.DiffFrom(&prev.WorkerAllocatorStatistics)

	return Statistics{
		EventsHandleSuccessTotal:  s.EventsHandleSuccessTotal - prev.EventsHandleSuccessTotal,
		EventsHandleFailureTotal:  s.EventsHandleFailureTotal - prev.EventsHandleFailureTotal,
		WorkerAllocatorStatistics: workerAllocatorStatisticsDiff,
	}
}

type SecretRef struct {
	Name string
	Key  string
}

type Secret struct {
	Contents  string
	SecretRef SecretRef
}
