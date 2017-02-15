package store

import (
	"policy-server/models"
	"time"
)

//go:generate counterfeiter -o fakes/store.go --fake-name Store . Store
type Store interface {
	Create([]models.Policy) error
	All() ([]models.Policy, error)
	Delete([]models.Policy) error
	Tags() ([]models.Tag, error)
	PoliciesWithFilter(models.PoliciesFilter) ([]models.Policy, error)
}

//go:generate counterfeiter -o fakes/metrics_sender.go --fake-name MetricsSender . metricsSender
type metricsSender interface {
	IncrementCounter(string)
	SendDuration(string, time.Duration)
}

type MetricsWrapper struct {
	Store         Store
	MetricsSender metricsSender
}

func (mw *MetricsWrapper) Create(policies []models.Policy) error {
	startTime := time.Now()
	err := mw.Store.Create(policies)
	createTimeDuration := time.Now().Sub(startTime)
	if err != nil {
		mw.MetricsSender.IncrementCounter("StoreCreateError")
	}
	mw.MetricsSender.SendDuration("StoreCreateTime", createTimeDuration)
	return err
}

func (mw *MetricsWrapper) All() ([]models.Policy, error) {
	startTime := time.Now()
	policies, err := mw.Store.All()
	allTimeDuration := time.Now().Sub(startTime)
	if err != nil {
		mw.MetricsSender.IncrementCounter("StoreAllError")
	}
	mw.MetricsSender.SendDuration("StoreAllTime", allTimeDuration)
	return policies, err
}

func (mw *MetricsWrapper) Delete(policies []models.Policy) error {
	startTime := time.Now()
	err := mw.Store.Delete(policies)
	deleteTimeDuration := time.Now().Sub(startTime)
	if err != nil {
		mw.MetricsSender.IncrementCounter("StoreDeleteError")
	}

	mw.MetricsSender.SendDuration("StoreDeleteTime", deleteTimeDuration)
	return err
}

func (mw *MetricsWrapper) Tags() ([]models.Tag, error) {
	startTime := time.Now()
	tags, err := mw.Store.Tags()
	tagsTimeDuration := time.Now().Sub(startTime)
	if err != nil {
		mw.MetricsSender.IncrementCounter("StoreTagsError")
	}
	mw.MetricsSender.SendDuration("StoreTagsTime", tagsTimeDuration)
	return tags, err
}

func (mw *MetricsWrapper) PoliciesWithFilter(policiesFilter models.PoliciesFilter) ([]models.Policy, error) {
	startTime := time.Now()
	policies, err := mw.Store.PoliciesWithFilter(policiesFilter)
	allTimeDuration := time.Now().Sub(startTime)
	if err != nil {
		mw.MetricsSender.IncrementCounter("StorePoliciesWithFilterError")
	}
	mw.MetricsSender.SendDuration("StorePoliciesWithFilterTime", allTimeDuration)
	return policies, err
}
