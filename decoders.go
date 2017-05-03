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
