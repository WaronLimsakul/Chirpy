package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

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

func TestJWTFunc(t *testing.T) {
	testSecrets := []string{"nerhgpodfhgnaorg23902l", "uon;ldsknfhsrdofgj", "ashifoas320r467eryl"}
	testUUIDs := uuid.UUIDs{uuid.New(), uuid.New(), uuid.New()}

	// normal test
	for i, testSecret := range testSecrets {
		tokenString, _ := MakeJWT(testUUIDs[i], testSecret, time.Minute*5)
		id, err := ValidateJWT(tokenString, testSecret)
		if err != nil {
			t.Errorf("test case %d: %s", i+1, err)
		} else if id != testUUIDs[i] {
			t.Errorf("test case %d: uuid not match: %s vs %s", i+1, id, testUUIDs[i])
		} else {
			t.Logf("normal test %d pass\n", i)
		}
	}
}

func TestJWTExpiration(t *testing.T) {
	// expiration test
	expirationTestSecret := "eriong23arsdlj"
	expiredToken, _ := MakeJWT(uuid.New(), expirationTestSecret, time.Millisecond*100)
	time.Sleep(time.Second)
	_, err := ValidateJWT(expiredToken, expirationTestSecret)
	if err == nil {
		t.Errorf("expiration test fail")
	} else {
		t.Logf("expiration test pass\n")
	}
}
