package apollo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSourceName(t *testing.T) {
	ap := new(ApolloConfigSource)
	assert.Equal(t, ap.GetSourceName(), ApolloName)
}

func TestGetPriority(t *testing.T) {
	ap := new(ApolloConfigSource)
	assert.Equal(t, ap.GetPriority(), ApolloPriority)
}


