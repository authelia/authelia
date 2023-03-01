package middlewares

import (
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/random"
)

func NewTimingAttackDelayer(name string, delay, min, max time.Duration, history int) *TimingAttackDelay {
	return &TimingAttackDelay{
		log:    logging.Logger().WithFields(map[string]any{"service": "timing attack delay", "name": name}),
		random: random.NewCryptographical(),

		mu: &sync.Mutex{},

		authns: delayedAuthns(delay, history),
		n:      history,
		i:      0,

		msDelayMin: float64(min.Milliseconds()),
		msDelayMax: big.NewInt(max.Milliseconds()),
	}
}

type TimingAttackDelay struct {
	log    *logrus.Entry
	random random.Provider

	mu sync.Locker

	authns []time.Duration
	n      int
	i      int

	msDelayMin float64
	msDelayMax *big.Int
}

func (m *TimingAttackDelay) Delay(successful bool, elapsed time.Duration) {
	time.Sleep(m.actual(elapsed, m.avg(elapsed, successful), successful))
}

func (m *TimingAttackDelay) actual(elapsed time.Duration, avg float64, successful bool) time.Duration {
	additional, err := m.random.IntErr(m.msDelayMax)
	if err != nil {
		return time.Millisecond * time.Duration(avg+float64(m.msDelayMax.Int64()))
	}

	total := math.Max(avg, m.msDelayMin) + float64(additional.Int64())
	actual := math.Max(total-float64(elapsed.Milliseconds()), 1.0)

	m.log.WithFields(map[string]any{
		"successful": successful,
		"elapsed":    elapsed.Milliseconds(),
		"average":    avg,
		"random":     additional.Int64(),
		"total":      total,
		"actual":     actual,
	}).Trace("Delaying to Prevent Timing Attacks")

	return time.Millisecond * time.Duration(actual)
}

func (m *TimingAttackDelay) avg(elapsed time.Duration, successful bool) float64 {
	var sum int64

	m.mu.Lock()

	for _, authn := range m.authns {
		sum += authn.Milliseconds()
	}

	if successful {
		m.authns[m.i] = elapsed
		m.i = (m.i + 1) % m.n
	}

	m.mu.Unlock()

	return float64(sum / int64(m.n))
}

func delayedAuthns(delay time.Duration, history int) []time.Duration {
	s := make([]time.Duration, history)

	for i := range s {
		s[i] = delay
	}

	return s
}
