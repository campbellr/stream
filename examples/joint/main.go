package main

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/alexander-yu/stream/joint"
)

// all metrics in the joint package must be passed through joint.Init
// if you want to push values to them
func initialize(metrics []joint.Metric) error {
	var errs []error

	for _, metric := range metrics {
		err := joint.Init(metric)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		var result *multierror.Error
		for _, err := range errs {
			result = multierror.Append(result, err)
		}
		return errors.Wrapf(result, "error initializing metrics")
	}

	return nil
}

func push(metrics []joint.Metric) error {
	var errs []error

	for _, metric := range metrics {
		for i := 0.; i < 100; i++ {
			err := metric.Push(i, i*i)
			if err != nil {
				errs = append(errs, err)
				break
			}
		}
	}

	if len(errs) != 0 {
		var result *multierror.Error
		for _, err := range errs {
			result = multierror.Append(result, err)
		}
		return errors.Wrapf(result, "error pushing to metrics")
	}

	return nil
}

func values(metrics []joint.Metric) (map[string]float64, error) {
	var errs []error

	result := map[string]float64{}
	for _, metric := range metrics {
		val, err := metric.Value()
		if err != nil {
			errs = append(errs, err)
			break
		}
		result[metric.String()] = val
	}

	if len(errs) != 0 {
		var result *multierror.Error
		for _, err := range errs {
			result = multierror.Append(result, err)
		}
		return nil, errors.Wrapf(result, "error retrieving values from metrics")
	}

	return result, nil
}

func main() {
	// tracks the global correlation
	corr := joint.NewCorrelation(5)

	// tracks the autocorrelation over a rolling window of size 10 and a lag of 4
	autocorr, err := joint.NewAutocorrelation(4, 10)
	if err != nil {
		log.Fatal(err)
	}

	// tracks the covariance over a rolling window of size 5
	cov := joint.NewCovariance(5)

	metrics := []joint.Metric{corr, autocorr, cov}

	err = initialize(metrics)
	if err != nil {
		log.Fatal(err)
	}

	err = push(metrics)
	if err != nil {
		log.Fatal(err)
	}

	values, err := values(metrics)
	if err != nil {
		log.Fatal(err)
	}

	result, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(result))
}
