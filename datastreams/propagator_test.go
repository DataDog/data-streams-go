// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package datastreams

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testPathway() Pathway {
	now := time.Now().Local().Truncate(time.Millisecond)
	return Pathway{
		hash:         234,
		pathwayStart: now.Add(-time.Hour),
		edgeStart:    now,
	}
}

func TestEncode(t *testing.T) {
	time := time.Unix(1680033770, 0)
	p := newPathway(time, "type:kafka")
	fmt.Println(p.hash)
	encoded := p.Encode()
	fmt.Println(encoded)
}

func TestEncodeStr(t *testing.T) {
	p := Pathway{
		hash:         234,
		pathwayStart: time.Unix(0, 0),
		edgeStart:    time.Unix(0, 0),
	}
	encoded := p.EncodeStr()
	fmt.Println("before")
	fmt.Println(encoded)
	fmt.Println("after")
	decoded, err := DecodeStr(encoded)
	assert.Nil(t, err)
	assert.Equal(t, p, decoded)
}
