package getversion

import "context"

//go:generate mockgen -destination=../mocks/mock_get_version.go -package=mocks -source=get_version.go

type Version struct {
	Version   string
	BuildTime string
}

// UseCaseGetVersion describes GetVersion contract.
type UseCaseGetVersion interface {
	Do(ctx context.Context) *Version
}
