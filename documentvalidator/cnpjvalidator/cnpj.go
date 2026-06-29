package cnpjvalidator

import (
	"strconv"
	"strings"

	"github.com/juliokobayashi-ebury/lib-teste/documentvalidator/documenthelpers"
)

type CNPJValidator struct{}

func (ref CNPJValidator) SanitizeAndValidate(data string) bool {
	data = documenthelpers.SanitizeDocument(data)

	if len(data) != 14 {
		return false
	}

	return ref.Validate(data)
}

func (ref CNPJValidator) Validate(data string) bool {
	if strings.Contains(blacklist, data) || !ref.check(data) {
		return false
	}

	return true
}

const blacklist = `00000000000000
11111111111111
22222222222222
33333333333333
44444444444444
55555555555555
66666666666666
77777777777777
88888888888888
99999999999999`

func (ref CNPJValidator) stringToIntSlice(data string) (res []int) {
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

func (ref CNPJValidator) check(data string) bool {
	return ref.verify(ref.stringToIntSlice(data), 5, 12) && ref.verify(ref.stringToIntSlice(data), 6, 13)
}

func (ref CNPJValidator) verify(data []int, j int, n int) bool {
	sum := 0

	for i := 0; i < n; i++ {
		v := data[i]
		sum += v * j

		if j == 2 {
			j = 9
		} else {
			j -= 1
		}
	}

	remainder := sum % 11

	v := data[n]
	x := 0

	if remainder >= 2 {
		x = 11 - remainder
	}

	if v != x {
		return false
	}

	return true
}
