package hasher

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	req := require.New(t)
	result := Hash("line")
	req.Equal("e2f9643543cba5c25dbbdd61ae3410ef7d0f38f81a153dcf156d121cb835f9dbdeeb2aee9f346f74ffc466e27db5abb8c7873d038da26fdb36e92ac0dbcc4175", result)
}

func TestMultiHash(t *testing.T) {
	req := require.New(t)
	result := MultiHash([]string{"testline", "testline2"})
	expected := []string{"6f624f2c02d7f0aace2a05768fefb6943822f9a8ba1245868de98586e1b061f67d828a6e2c130902ca18e1e95ed88c9a5dd5f3244aa0579236bb24cf967d96d0", "11892d55cf5060ecba7311d7208afabf8bd17fe53273edadc6bf851f475c67d98fad43f55baa1d02f8a64526d0e76a9817ece3814e0914d263a0ba3f0680f976"}
	req.Equal(expected, result)
}
