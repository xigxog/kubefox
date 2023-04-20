package blocker

import "fmt"

type ErrChannelNotFound struct {
	id string
}

func (err *ErrChannelNotFound) Error() string {
	return fmt.Sprintf("not found, channel %s does not exist", err.id)
}

type ErrDuplicateChannel struct {
	id string
}

func (err *ErrDuplicateChannel) Error() string {
	return fmt.Sprintf("conflict, channel %s already exists", err.id)
}
