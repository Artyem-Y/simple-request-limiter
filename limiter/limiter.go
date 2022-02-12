package limiter

import (
	"net/http"
	"sync"
	"time"
)

type IndicatorGetter func(r *http.Request) string

type RequestLimitService struct {
	Interval       time.Duration
	MaxCount       int
	Lock           sync.Mutex
	limitIndicator IndicatorGetter
	limitTable     map[string]*groupLimit
	errorHandler   http.HandlerFunc
}

func NewRequestLimitService(interval time.Duration, maxCnt int, limitIndicator IndicatorGetter) *RequestLimitService {
	return &RequestLimitService{
		Interval:       interval,
		MaxCount:       maxCnt,
		limitIndicator: limitIndicator,
		limitTable:     make(map[string]*groupLimit),
		errorHandler:   nil,
	}
}

func (reqLimit *RequestLimitService) isAllowed(indicator string) bool {
	reqLimit.Lock.Lock()
	defer reqLimit.Lock.Unlock()

	group, ok := reqLimit.limitTable[indicator]
	if !ok {
		group = newGroupLimit(reqLimit.Interval)
		reqLimit.limitTable[indicator] = group
	}

	if group.isStale() {
		group.refresh(reqLimit.Interval)
	}

	group.reqCount++
	return group.reqCount <= reqLimit.MaxCount
}

func (reqLimit *RequestLimitService) defaultErrorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
	}
}

func (reqLimit *RequestLimitService) OnLimitReached(errorHandler http.HandlerFunc) {
	reqLimit.errorHandler = errorHandler
}

func (reqLimit *RequestLimitService) Limit(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var indicator string

		if reqLimit.limitIndicator != nil {
			indicator = reqLimit.limitIndicator(r)
		}

		if !reqLimit.isAllowed(indicator) {
			if reqLimit.errorHandler != nil {
				reqLimit.errorHandler.ServeHTTP(w, r)
				return
			}

			reqLimit.defaultErrorHandler().ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
