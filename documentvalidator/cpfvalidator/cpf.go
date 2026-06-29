package cpfvalidator

import (
	"strconv"
	"strings"

	"github.com/juliokobayashi-ebury/lib-teste/documentvalidator/documenthelpers"
)

type CPFValidator struct{}

func (ref CPFValidator) SanitizeAndValidate(data string) bool {
	data = documenthelpers.SanitizeDocument(data)

	if len(data) != 11 {
		return false
	}

	return ref.Validate(data)
}

func (ref CPFValidator) Validate(data string) bool {
	if strings.Contains(blacklist, data) || !ref.check(data) {
		return false
	}
	return true
}

const blacklist = `00000000000
11111111111
22222222222
33333333333
44444444444
55555555555
66666666666
77777777777
88888888888
99999999999`

func (ref CPFValidator) stringToIntSlice(data string) (res []int) {
	for _, d := range data {
		if d >= '0' && d <= '9' {
			x, err := strconv.Atoi(string(d))
			if err != nil {
				continue
			}
			res = append(res, x)
		}
		if d >= 'A' && d <= 'Z' {
			intVal := int(d)
			res = append(res, intVal-48)
		}
	}
	return
}

func (ref CPFValidator) check(data string) bool {
	return ref.verify(ref.stringToIntSlice(data), 10, 9) && ref.verify(ref.stringToIntSlice(data), 11, 10)
}

func (ref CPFValidator) verify(data []int, j int, n int) bool {
	sum := 0

	for i := 0; i < n; i++ {
		v := data[i]
		sum += v * j

		j -= 1
	}

	remainder := (sum * 10) % 11
	if remainder == 10 {
		remainder = 0
	}

	v := data[n]

	if v != remainder {
		return false
	}

	return true
}
