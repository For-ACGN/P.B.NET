package security

import (
	"errors"
	"unsafe"
)

// 1: password length must longer than 12
// 2: must contain number
// 3: must contain upper and lower case letters
// 4: must contain special symbols

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
	if len(password) < 12 {
		return errors.New("password length must longer than 12")
	}
	rules := make([]bool, len(ruleErrors))
	var (
		last rune
		str  []rune
		hit  int
	)
	for _, s := range password {
		if s == last || s-1 == last || s+1 == last {
			str = append(str, s)
			hit++
		} else {
			str = nil
		}
		if hit > 4 {
			return errors.New("find continuous content: " + string(str))
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
