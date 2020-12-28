package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/hectorchu/gonano/util"
	"github.com/hectorchu/nano-token-protocol/tokenchain"
)

const rpcURL = "https://mynano.ninja/api/node"

func main() {
	var chain *tokenchain.Chain
	wallet, err := loadWallet(rpcURL)
	if err != nil {
		seed := make([]byte, 32)
		rand.Read(seed)
		if wallet, err = newWallet(seed, rpcURL); err != nil {
			fmt.Println(err)
		} else if err = wallet.save(); err != nil {
			fmt.Println(err)
		}
	}
	for {
		for wallet.account() != nil {
			balance, pending, err := wallet.account().Balance()
			if err != nil {
				fmt.Println(err)
			} else {
				if pending.Sign() > 0 {
					fmt.Println("Receiving pending amounts...")
					if err = wallet.account().ReceivePendings(); err != nil {
						fmt.Println(err)
					} else {
						continue
					}
				}
				fmt.Printf("\nAccount %s, Balance = %s\n", wallet.account().Address(), util.NanoAmount{Raw: balance})
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
		fmt.Println("(a) Change account index")
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
			if wallet, err = newWallet(seed, rpcURL); err != nil {
				fmt.Println(err)
				continue
			}
			if err = wallet.save(); err != nil {
				fmt.Println(err)
				continue
			}
		case "a":
			index, err := strconv.Atoi(readLine("Enter account index: "))
			if err != nil {
				fmt.Println(err)
				continue
			}
			if err = wallet.changeAccount(uint32(index)); err != nil {
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
			if _, err = wallet.account().Send(chain.Address(), big.NewInt(1)); err != nil {
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
			token, err := tokenchain.TokenGenesis(chain, wallet.account(), name, supply, byte(decimals))
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
				caret := " "
				if account == wallet.account().Address() {
					caret = ">"
				}
				fmt.Printf("%s %s = %s %s\n", caret, account, amountToString(balance, token.Decimals()), token.Name())
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
			if hash, err = token.Transfer(wallet.account(), destination, amount); err != nil {
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
