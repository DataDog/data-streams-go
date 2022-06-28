// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package kafka

import (
	"testing"

	"github.com/DataDog/data-streams-go/datastreams"
	"github.com/stretchr/testify/assert"
)

func assertPathwayNotEqual(t *testing.T, p1 datastreams.Pathway, p2 datastreams.Pathway) {
	decodedP1, err1 := datastreams.Decode(p1.Encode())
	decodedP2, err2 := datastreams.Decode(p2.Encode())

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.NotEqual(t, decodedP1, decodedP2)
}

func assertPathwayEqual(t *testing.T, p1 datastreams.Pathway, p2 datastreams.Pathway) {
	decodedP1, err1 := datastreams.Decode(p1.Encode())
	decodedP2, err2 := datastreams.Decode(p2.Encode())

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, decodedP1, decodedP2)
}
