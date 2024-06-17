package listener

import (
	"context"

	throttle "github.com/yaronsumel/grpc-throttle"
)

func throttleFn(_ context.Context, _ string) (throttle.Semaphore, bool) {
	const throttleMax = 10
	return make(throttle.Semaphore, throttleMax), true
}
