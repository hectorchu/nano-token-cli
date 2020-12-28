package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func readString(prompt string) (input string) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	return strings.TrimRight(input, "\r\n")
}

func readHex(prompt string) (data []byte, err error) {
	return hex.DecodeString(readString(prompt))
}

func readInt(prompt string) (x int, err error) {
	return strconv.Atoi(readString(prompt))
}

func readAmount(prompt string) (x *big.Int, err error) {
	x, ok := new(big.Int).SetString(readString(prompt), 10)
	if !ok {
		err = errors.New("Failed reading amount")
	}
	return
}

func readAmountWithDecimals(prompt string, decimals byte) (x *big.Int, err error) {
	return amountFromString(readString(prompt), decimals)
}
