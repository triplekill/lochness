package cli_test

import (
	"os"
	"strings"
	"testing"

	"github.com/mistifyio/lochness/internal/cli"
	"github.com/stretchr/testify/suite"
)

type JMapTestSuite struct {
	suite.Suite
}

func TestJMapTestSuite(t *testing.T) {
	suite.Run(t, new(JMapTestSuite))
}

func (s *JMapTestSuite) TestID() {
	j := &cli.JMap{}
	s.Empty(j.ID())

	j = &cli.JMap{"id": "asdf"}
	s.Equal("asdf", j.ID())
}

func (s *JMapTestSuite) TestString() {
	j := &cli.JMap{"id": "asdf", "foo": "bar"}
	s.Equal(`{"foo":"bar","id":"asdf"}`, j.String())
}

func (s *JMapTestSuite) TestPrint() {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		w.Close()
		os.Stdout = stdout
	}()

	j := &cli.JMap{"id": "asdf", "foo": "bar"}

	j.Print(false) // id only
	j.Print(true)  // full json

	buf := make([]byte, 32)
	_, _ = r.Read(buf)
	results := strings.Split(string(buf), "\n")

	s.Equal(j.ID(), results[0])
	s.Equal(j.String(), results[1])
}

func (s *JMapTestSuite) TestLen() {
	jms := cli.JMapSlice{cli.JMap{}, cli.JMap{}}
	s.Equal(2, jms.Len())
}

func (s *JMapTestSuite) TestLess() {
	jms := cli.JMapSlice{
		cli.JMap{"id": "a"},
		cli.JMap{"id": "b"},
	}

	s.True(jms.Less(0, 1))
	s.False(jms.Less(1, 0))
}

func (s *JMapTestSuite) TestSwap() {
	j0 := cli.JMap{"id": "a"}
	j1 := cli.JMap{"id": "b"}
	jms := cli.JMapSlice{
		j0,
		j1,
	}

	jms.Swap(0, 1)
	s.Equal(j1, jms[0])
	s.Equal(j0, jms[1])
}
