package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/hectorchu/gonano/util"
	"github.com/hectorchu/gonano/wallet"
	"github.com/hectorchu/nano-token-server/tokenchain"
)

const rpcURL = "https://mynano.ninja/api/node"

var reader = bufio.NewReader(os.Stdin)

func readLine(prompt string) (input string) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	return strings.TrimRight(input, "\r\n")
}

func initAccount(seed []byte) (a *wallet.Account, err error) {
	w, err := wallet.NewWallet(seed)
	if err != nil {
		return
	}
	w.RPC.URL = rpcURL
	return w.NewAccount(nil)
}

func initAccountFromFile() (a *wallet.Account, err error) {
	f, err := os.Open("seed.bin")
	if err != nil {
		return
	}
	defer f.Close()
	seed := make([]byte, 32)
	if _, err = f.Read(seed); err != nil {
		return
	}
	return initAccount(seed)
}

func saveSeed(seed []byte) (err error) {
	f, err := os.Create("seed.bin")
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(seed)
	return
}

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

func main() {
	var chain *tokenchain.Chain
	account, err := initAccountFromFile()
	if err != nil {
		seed := make([]byte, 32)
		rand.Read(seed)
		if account, err = initAccount(seed); err != nil {
			fmt.Println(err)
		} else if err = saveSeed(seed); err != nil {
			fmt.Println(err)
		}
	}
	for {
		for account != nil {
			balance, pending, err := account.Balance()
			if err != nil {
				fmt.Println(err)
			} else {
				if pending.Sign() > 0 {
					if err = account.ReceivePendings(); err != nil {
						fmt.Println(err)
					} else {
						continue
					}
				}
				fmt.Printf("\nAccount %s, Balance = %s\n", account.Address(), util.NanoAmount{Raw: balance})
			}
			break
		}
		if chain != nil {
			fmt.Println("\nChain", chain.Address())
			fmt.Println("Tokens:")
			for _, token := range chain.Tokens() {
				hash := strings.ToUpper(hex.EncodeToString(token.Hash()))
				fmt.Printf("%s: Hash=%s, Supply=%s\n", token.Name(), hash, amountToString(token.Supply(), token.Decimals()))
			}
		}
		fmt.Println("\nChoose option:")
		fmt.Println("(w) Set the wallet seed")
		fmt.Println("(c) Create a new token chain")
		fmt.Println("(l) Load a token chain")
		fmt.Println("(p) Parse the chain")
		fmt.Println("(g) Create a token")
		fmt.Println("(b) Get balances for a token")
		fmt.Println("(t) Transfer a token")
		fmt.Println("(q) Quit")
		switch input := readLine("> "); input {
		case "w":
			seed, err := hex.DecodeString(readLine("Enter seed: "))
			if err != nil {
				fmt.Println(err)
				continue
			}
			if account, err = initAccount(seed); err != nil {
				fmt.Println(err)
				continue
			}
			if err = saveSeed(seed); err != nil {
				fmt.Println(err)
				continue
			}
		case "c":
			chain, err = tokenchain.NewChain(rpcURL)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Initializing new chain at", chain.Address())
			if _, err = account.Send(chain.Address(), big.NewInt(1)); err != nil {
				fmt.Println(err)
				continue
			}
			if err = chain.WaitForOpen(); err != nil {
				fmt.Println(err)
				continue
			}
		case "l":
			var err error
			if chain, err = tokenchain.LoadChain(readLine("Enter address: "), rpcURL); err != nil {
				fmt.Println(err)
				continue
			}
			fallthrough
		case "p":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			if err := chain.Parse(); err != nil {
				fmt.Println(err)
				continue
			}
		case "g":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			name := readLine("Enter name: ")
			supply, ok := new(big.Int).SetString(readLine("Enter supply: "), 10)
			if !ok {
				fmt.Println("Failed reading supply")
				continue
			}
			decimals, err := strconv.Atoi(readLine("Enter decimals: "))
			if err != nil {
				fmt.Println(err)
				continue
			}
			token, err := tokenchain.TokenGenesis(chain, account, name, supply, byte(decimals))
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Token created with hash", strings.ToUpper(hex.EncodeToString(token.Hash())))
		case "b":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			hash, err := hex.DecodeString(readLine("Enter token hash: "))
			if err != nil {
				fmt.Println(err)
				continue
			}
			token, err := chain.Token(hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Balances:")
			for account, balance := range token.Balances() {
				fmt.Printf("%s = %s %s\n", account, amountToString(balance, token.Decimals()), token.Name())
			}
		case "t":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			hash, err := hex.DecodeString(readLine("Enter token hash: "))
			if err != nil {
				fmt.Println(err)
				continue
			}
			token, err := chain.Token(hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			destination := readLine("Enter destination account: ")
			amount, err := amountFromString(readLine("Enter amount: "), token.Decimals())
			if err != nil {
				fmt.Println(err)
				continue
			}
			if hash, err = token.Transfer(account, destination, amount); err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Token transferred with hash", strings.ToUpper(hex.EncodeToString(hash)))
		case "q":
			return
		default:
			fmt.Println("Unknown option", input)
		}
	}
}
