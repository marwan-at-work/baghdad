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

// DecodeBuildEvent is used to parse a BuildEvent from a rabbitmq msg
func DecodeBuildEvent(p []byte) (be BuildEvent, err error) {
	be = BuildEvent{}

	buf := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&be)

	return
}

// DecodeDeployJob is used to parse a DeployJob from a rabbitmq msg
func DecodeDeployJob(p []byte) (dj DeployJob, err error) {
	dj = DeployJob{}

	buf := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&dj)

	return
}

// DecodePostDeployJob is used to parse a DeployJob from a rabbitmq msg
func DecodePostDeployJob(p []byte) (pdj PostDeployJob, err error) {
	pdj = PostDeployJob{}

	buf := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&pdj)

	return
}

// DecodeSecretsJob is used to parse a SecretsJob from a rabbitmq msg
func DecodeSecretsJob(p []byte) (sj SecretsJob, err error) {
	sj = SecretsJob{}

	buf := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&sj)

	return
}
