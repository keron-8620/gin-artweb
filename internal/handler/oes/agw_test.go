package oes

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/shared/test"
)

type AgwHandlerTestSuite struct {
	suite.Suite
}

func (s *AgwHandlerTestSuite) TestNewAgwHandler() {
	logger := test.NewTestZapLogger()

	handler := NewOesAgwHandler(logger, nil)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

func TestAgwHandlerTestSuite(t *testing.T) {
	suite.Run(t, &AgwHandlerTestSuite{})
}
