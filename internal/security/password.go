package security

import (
	"errors"
	"fmt"
	"unsafe"
)

// 1: password length must longer than 12
// 2: must contain number
// 3: must contain upper and lower case letters
// 4: must contain special symbols
// 5: continuous string must less than 5

const (
	minPasswordLen = 12
	maxContinuous  = 4
)

var errPasswordLen = fmt.Errorf("password length must longer than %d", minPasswordLen)

type rule int

const (
	containNumber rule = iota
	containUpper
	containLower
	containSpecial
)

var ruleErrors = map[rule]error{
	containNumber:  errors.New("must contain number"),
	containUpper:   errors.New("must contain upper case letter"),
	containLower:   errors.New("must contain lower case letter"),
	containSpecial: errors.New("must contain special symbol"),
}

// CheckPasswordStrength is used to check password byte slice strength.
func CheckPasswordStrength(password []byte) error {
	str := *(*string)(unsafe.Pointer(&password))
	return CheckPasswordStringStrength(str)
}

// CheckPasswordStringStrength is used to check password string strength.
func CheckPasswordStringStrength(password string) error {
	if len(password) < minPasswordLen {
		return errPasswordLen
	}
	rules := make([]bool, len(ruleErrors))
	str := make([]rune, maxContinuous+1)
	var (
		last rune
		hit  int
	)
	for _, s := range password {
		if s == last || s-1 == last || s+1 == last {
			str = append(str, s)
			hit++
			if hit > maxContinuous-1 {
				return fmt.Errorf("find continuous content: \"%s\"", string(str))
			}
		} else {
			str[0] = s
			str = str[:1]
			hit = 0
		}
		switch {
		case s >= '0' && s <= '9':
			rules[containNumber] = true
		case s >= 'A' && s <= 'Z':
			rules[containUpper] = true
		case s >= 'a' && s <= 'z':
			rules[containLower] = true
		default:
			rules[containSpecial] = true
		}
		last = s
	}
	for i := 0; i < len(rules); i++ {
		if !rules[i] {
			return ruleErrors[rule(i)]
		}
	}
	return nil
}
