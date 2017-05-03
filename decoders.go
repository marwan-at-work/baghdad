package baghdad

import (
	"bytes"
	"encoding/gob"
)

// DecodeBuildJob is used to parse a BuildJob from a rabbitmq msg
func DecodeBuildJob(p []byte) (bj BuildJob, err error) {
	bj = BuildJob{}

	buf := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&bj)

	return
}

// DecodeBuildEvent is used to parse a BuildJob from a rabbitmq msg
func DecodeBuildEvent(p []byte) (be BuildEvent, err error) {
	be = BuildEvent{}

	buf := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&be)

	return
}
