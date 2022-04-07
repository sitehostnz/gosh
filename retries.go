package gosh

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	retryAfterHeaderName      = "Retry-After"
	maintenanceModeHeaderName = "X-Maintenance-Mode"
)

// type RetryConditional func(r *resty.Response) (shouldRetry bool)
type RetryConditional resty.RetryConditionFunc

// type RetryAfter func(c *resty.Client, r *resty.Response) (time.Duration, error)
type RetryAfter resty.RetryAfterFunc

// Configures resty to
// lock until enough time has passed to retry the request as determined by the Retry-After response header.
// If the Retry-After header is not set, we fall back to value of SetPollDelay.
func configureRetries(c *Client) {
	c.resty.
		SetRetryCount(1000).
		AddRetryCondition(checkRetryConditionals(c)).
		SetRetryAfter(respectRetryAfter)
}

func checkRetryConditionals(c *Client) func(*resty.Response, error) bool {
	return func(r *resty.Response, err error) bool {
		for _, retryConditional := range c.retryConditionals {
			retry := retryConditional(r, err)
			if retry {
				log.Printf("[INFO] Received error %s - Retrying", r.Error())
				return true
			}
		}
		return false
	}
}

func tooManyRequestsRetryCondition(r *resty.Response, _ error) bool {
	return r.StatusCode() == http.StatusTooManyRequests
}

func serviceUnavailableRetryCondition(r *resty.Response, _ error) bool {
	serviceUnavailable := r.StatusCode() == http.StatusServiceUnavailable

	return serviceUnavailable
}

func requestTimeoutRetryCondition(r *resty.Response, _ error) bool {
	return r.StatusCode() == http.StatusRequestTimeout
}

func respectRetryAfter(client *resty.Client, resp *resty.Response) (time.Duration, error) {
	retryAfterStr := resp.Header().Get(retryAfterHeaderName)
	if retryAfterStr == "" {
		return 0, nil
	}

	retryAfter, err := strconv.Atoi(retryAfterStr)
	if err != nil {
		return 0, err
	}

	duration := time.Duration(retryAfter) * time.Second
	log.Printf("[INFO] Respecting Retry-After Header of %d (%s) (max %s)", retryAfter, duration, client.RetryMaxWaitTime)
	return duration, nil
}
