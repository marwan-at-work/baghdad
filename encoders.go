package baghdad

import (
	"bytes"
	"encoding/gob"
)

// EncodeBuildJob is used to serialize a BuildJob for a rabbitmq msg
func EncodeBuildJob(bj BuildJob) (p []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(bj)
	if err != nil {
		return
	}

	p = buf.Bytes()

	return
}

// EncodeBuildEvent is used to serialize a BuildEvent to a rabbitmq msg
func EncodeBuildEvent(be BuildEvent) (p []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(be)
	if err != nil {
		return
	}

	p = buf.Bytes()

	return
}
