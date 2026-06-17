package resource

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/shared/test"
)

type TerminalHandlerTestSuite struct {
	suite.Suite
}

func (s *TerminalHandlerTestSuite) TestNewTerminalHandler() {
	logger := test.NewTestZapLogger()

	handler := NewTerminalHandler(logger, nil)

	s.NotNil(handler)
	s.NotNil(handler.logger)
}

func TestTerminalHandlerTestSuite(t *testing.T) {
	suite.Run(t, &TerminalHandlerTestSuite{})
}
