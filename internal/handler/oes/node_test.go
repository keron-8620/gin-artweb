package oes

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/shared/test"
)

type OesNodeHandlerTestSuite struct {
	suite.Suite
}

func (s *OesNodeHandlerTestSuite) TestNewOesNodeHandler() {
	logger := test.NewTestZapLogger()

	handler := NewOesNodeHandler(logger, nil)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

func TestOesNodeHandlerTestSuite(t *testing.T) {
	suite.Run(t, &OesNodeHandlerTestSuite{})
}
