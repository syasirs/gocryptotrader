package nonce

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetAndIncrement(t *testing.T) {
	var nonce Nonce
	n1 := nonce.GetAndIncrement(time.Now().Unix)
	assert.NotZero(t, n1)
	n2 := nonce.GetAndIncrement(time.Now().Unix)
	assert.NotZero(t, n2)
	assert.NotEqual(t, n1, n2)

	var nonce2 Nonce
	n3 := nonce2.GetAndIncrement(time.Now().UnixNano)
	assert.NotZero(t, n3)
	n4 := nonce2.GetAndIncrement(time.Now().UnixNano)
	assert.NotZero(t, n4)
	assert.NotEqual(t, n3, n4)

	assert.NotEqual(t, n1, n3)
	assert.NotEqual(t, n2, n4)
}

func TestString(t *testing.T) {
	var nonce Nonce
	nonce.n = 12312313131
	got := nonce.GetAndIncrement(time.Now().Unix)
	assert.Equal(t, "12312313132", got.String())

	got = nonce.GetAndIncrement(time.Now().Unix)
	assert.Equal(t, "12312313133", got.String())
}
