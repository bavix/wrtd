package app

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/digineo/go-uci"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/config"
	"go.uber.org/zap"

	"github.com/bavix/dialer"
	"github.com/bavix/wrtd/internal/pkg/scheduler"
)

type Configure struct {
	Interface string
	Network   string
	Address   string
	Timeout   *time.Duration
	Timer     *time.Duration
}

func (c Configure) GetTimeout() time.Duration {
	if c.Timeout == nil {
		return 100 * time.Millisecond //nolint:gomnd
	}

	return *c.Timeout
}

func (c Configure) GetTimer() time.Duration {
	if c.Timer == nil {
		return 30 * time.Second //nolint:gomnd
	}

	return *c.Timer
}

type CheckList struct {
	Network []Configure
}

func NewCheckList(cfg *config.YAML) (CheckList, error) {
	var checkList CheckList

	if err := cfg.Get("").Populate(&checkList); err != nil {
		return checkList, err
	}

	return checkList, nil
}

type Checker struct {
	checkList []Configure

	log *zap.Logger
}

func NewChecker(list CheckList, logger *zap.Logger) *Checker {
	return &Checker{
		checkList: list.Network,
		log:       logger,
	}
}

//nolint:funlen
func (c *Checker) Run(ctx context.Context) {
	var wg sync.WaitGroup

	hostname, _ := os.Hostname()

	for _, configure := range c.checkList {
		wg.Add(1)

		configure := configure

		ipaddr := c.getAddress(configure)

		//nolint:gomnd
		histogram := promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dialer",
			Help:    "A histogram of normally distributed random numbers.",
			Buckets: prometheus.ExponentialBucketsRange(0.0001, 1, 250),
			ConstLabels: map[string]string{
				"interface": configure.Interface,
				"network":   configure.Network,
				"ipaddr":    ipaddr,
				"hostname":  hostname,
			},
		})

		counter := promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "dialer_total",
			Help: "Processed requests",
			ConstLabels: map[string]string{
				"interface": configure.Interface,
				"network":   configure.Network,
				"ipaddr":    ipaddr,
				"hostname":  hostname,
			},
		}, []string{"type"})

		counterOK := counter.WithLabelValues("ok")
		counterErr := counter.WithLabelValues("error")

		go func() {
			defer wg.Done()

			scheduler.RunTask(ctx, configure.GetTimer(), func(ctx context.Context, triggered time.Time) {
				ctx, cancel := context.WithTimeout(ctx, configure.GetTimeout())
				defer cancel()

				elapsed, err := dialer.Dial(ctx, configure.Network, ipaddr)
				if err != nil {
					c.log.Error("dialer error",
						zap.Error(err),
						zap.String("interface", configure.Interface),
						zap.String("ipaddr", ipaddr),
						zap.Time("triggered", triggered))

					counterErr.Inc()

					return
				}

				c.log.Info("update metrics",
					zap.String("interface", configure.Interface),
					zap.String("ipaddr", ipaddr),
					zap.Duration("elapsed", elapsed),
					zap.Time("triggered", triggered))

				counterOK.Inc()

				histogram.Observe(elapsed.Seconds())
			})
		}()
	}

	wg.Wait()
}

func (c *Checker) getAddress(configure Configure) string {
	if configure.Address != "" {
		c.log.Info("network ipaddr from config",
			zap.String("interface", configure.Interface),
			zap.String("ipaddr", configure.Address))

		return configure.Address
	}

	values, ok := uci.Get("network", configure.Interface, "ipaddr")

	if !ok {
		c.log.Error(
			"network ipaddr not found",
			zap.String("interface", configure.Interface))
	}

	for _, value := range values {
		c.log.Info("network ipaddr from uci",
			zap.String("interface", configure.Interface),
			zap.String("ipaddr", value))

		return value + ":80"
	}

	return ""
}
