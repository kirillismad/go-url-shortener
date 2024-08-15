package tests

func (s *IntegrationTestSuite) TestExample() {
	a := s.Assert()

	a.Equal(1, 1)
}
