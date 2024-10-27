package balance_prover_service

import (
	"errors"
	"intmax2-node/internal/logger"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

type retryStrategy struct {
	log logger.Logger
}

func NewRetryStrategy(log logger.Logger) *retryStrategy {
	return &retryStrategy{
		log: log,
	}
}

func (r *retryStrategy) checkRetryCondition(response *resty.Response, err error) bool {
	if err != nil && strings.Contains(err.Error(), ErrRequestIoTimeout.Error()) {
		r.log.Debugf("retrying due to i/o timeout\n")
		return true
	}

	if response.StatusCode() == http.StatusRequestTimeout {
		r.log.Debugf("retrying due to status request timeout\n")
		return true
	}

	return false
}

func (r *retryStrategy) Condition() resty.RetryConditionFunc {
	return r.checkRetryCondition
}

func HandleGeneralError(resp *resty.Response, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), ErrRequestIoTimeout.Error()) {
			return ErrStatusRequestTimeout
		}

		return err
	}

	if resp == nil {
		return errors.New("response is nil")
	}

	if resp.StatusCode() == http.StatusRequestTimeout {
		return ErrStatusRequestTimeout
	}

	return nil
}
