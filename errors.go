package chocolateclashgoapi

import "errors"

var (
	ErrUnknownLeague     = errors.New("unknown league")
	ErrFailedToFixWarPid = errors.New("failed to fix war pid issue")
)
