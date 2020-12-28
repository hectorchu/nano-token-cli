package main

import (
	"errors"
	"math/big"
)

func amountToString(amount *big.Int, decimals byte) string {
	x := big.NewInt(10)
	exp := x.Exp(x, big.NewInt(int64(decimals)), nil)
	r := new(big.Rat).SetFrac(amount, exp)
	return r.FloatString(int(decimals))
}

func amountFromString(s string, decimals byte) (amount *big.Int, err error) {
	x := big.NewInt(10)
	exp := x.Exp(x, big.NewInt(int64(decimals)), nil)
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, errors.New("Unable to parse amount")
	}
	r = r.Mul(r, new(big.Rat).SetInt(exp))
	if !r.IsInt() {
		return nil, errors.New("Unable to parse amount")
	}
	return r.Num(), nil
}
