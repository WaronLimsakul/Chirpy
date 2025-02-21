package auth

import "testing"

func TestHashing(t *testing.T) {
	testCases := []string{"abcdefg", "hijklmnop", "123456789"}
	for _, init := range testCases {
		hashedInit, _ := HashPassword(init)
		for _, tester := range testCases {
			unMatch := CheckPasswordHash(tester, hashedInit)
			if init == tester && unMatch != nil {
				t.Errorf("%s should match", init)
			} else if init != tester && unMatch == nil {
				t.Errorf("%s and %s shouldn't match", init, tester)
			}
		}
	}
}
