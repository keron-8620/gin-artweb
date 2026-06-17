package oes

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/shared/test"
)

type OesConfHandlerTestSuite struct {
	suite.Suite
}

func (s *OesConfHandlerTestSuite) TestNewOesConfHandler() {
	logger := test.NewTestZapLogger()

	handler := NewOesConfHandler(logger, 500)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

func TestOesConfHandlerTestSuite(t *testing.T) {
	suite.Run(t, &OesConfHandlerTestSuite{})
}
