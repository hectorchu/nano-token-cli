package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/hectorchu/gonano/rpc"
	"github.com/hectorchu/gonano/util"
	"github.com/hectorchu/nano-token-protocol/tokenchain"
)

const rpcURL = "https://mynano.ninja/api/node"

func hashToString(hash rpc.BlockHash) string {
	return strings.ToUpper(hex.EncodeToString(hash))
}

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
			if err := chain.Parse(); err != nil {
				fmt.Println(err)
			}
			fmt.Println("\nChain", chain.Address())
			fmt.Println("Tokens:")
			for _, token := range chain.Tokens() {
				fmt.Printf("%s: Hash=%s, Supply=%s\n",
					token.Name(), hashToString(token.Hash()),
					amountToString(token.Supply(), token.Decimals()))
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
		fmt.Println("(sp) Propose a swap")
		fmt.Println("(sa) Accept a swap")
		fmt.Println("(sc) Confirm a swap")
		fmt.Println("(sn) Cancel a swap")
		fmt.Println("(si) Get swap details")
		fmt.Println("(q) Quit")
		switch input := readString("> "); input {
		case "w":
			seed, err := readHex("Enter seed: ")
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
			index, err := readInt("Enter account index: ")
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
			if chain, err = tokenchain.LoadChain(readString("Enter address: "), rpcURL); err != nil {
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
			name := readString("Enter name: ")
			decimals, err := readInt("Enter decimals: ")
			if err != nil {
				fmt.Println(err)
				continue
			}
			supply, err := readAmountWithDecimals("Enter supply: ", byte(decimals))
			if err != nil {
				fmt.Println(err)
				continue
			}
			token, err := tokenchain.TokenGenesis(chain, wallet.account(), name, supply, byte(decimals))
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Token created with hash", hashToString(token.Hash()))
		case "b":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			hash, err := readHex("Enter token hash: ")
			if err != nil {
				fmt.Println(err)
				continue
			}
			if err := chain.Parse(); err != nil {
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
			hash, err := readHex("Enter token hash: ")
			if err != nil {
				fmt.Println(err)
				continue
			}
			token, err := chain.Token(hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			destination := readString("Enter destination account: ")
			amount, err := readAmountWithDecimals("Enter amount: ", token.Decimals())
			if err != nil {
				fmt.Println(err)
				continue
			}
			if hash, err = token.Transfer(wallet.account(), destination, amount); err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Token transferred with hash", hashToString(hash))
		case "sp":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			counterparty := readString("Enter counterparty account: ")
			hash, err := readHex("Enter token hash: ")
			if err != nil {
				fmt.Println(err)
				continue
			}
			token, err := chain.Token(hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			amount, err := readAmountWithDecimals("Enter amount: ", token.Decimals())
			if err != nil {
				fmt.Println(err)
				continue
			}
			swap, err := tokenchain.ProposeSwap(chain, wallet.account(), counterparty, token, amount)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Swap created with hash", hashToString(swap.Hash()))
		case "sa":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			hash, err := readHex("Enter swap hash: ")
			if err != nil {
				fmt.Println(err)
				continue
			}
			swap, err := chain.Swap(hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			hash, err = readHex("Enter token hash: ")
			if err != nil {
				fmt.Println(err)
				continue
			}
			token, err := chain.Token(hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			amount, err := readAmountWithDecimals("Enter amount: ", token.Decimals())
			if err != nil {
				fmt.Println(err)
				continue
			}
			hash, err = swap.Accept(wallet.account(), token, amount)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Swap accepted with hash", hashToString(hash))
		case "sc":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			hash, err := readHex("Enter swap hash: ")
			if err != nil {
				fmt.Println(err)
				continue
			}
			swap, err := chain.Swap(hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			hash, err = swap.Confirm(wallet.account())
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Swap confirmed with hash", hashToString(hash))
		case "sn":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			hash, err := readHex("Enter swap hash: ")
			if err != nil {
				fmt.Println(err)
				continue
			}
			swap, err := chain.Swap(hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			hash, err = swap.Cancel(wallet.account())
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Swap cancelled with hash", hashToString(hash))
		case "si":
			if chain == nil {
				fmt.Println("No chain loaded")
				continue
			}
			hash, err := readHex("Enter swap hash: ")
			if err != nil {
				fmt.Println(err)
				continue
			}
			if err := chain.Parse(); err != nil {
				fmt.Println(err)
				continue
			}
			swap, err := chain.Swap(hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Left side:")
			fmt.Println("  Account =", swap.Left().Account)
			fmt.Printf("  Token = %s (Hash = %s)\n", swap.Left().Token.Name(), hashToString(swap.Left().Token.Hash()))
			fmt.Println("  Amount =", amountToString(swap.Left().Amount, swap.Left().Token.Decimals()))
			fmt.Println("Right side:")
			fmt.Println("  Account =", swap.Right().Account)
			if swap.Right().Token != nil {
				fmt.Printf("  Token = %s (Hash = %s)\n", swap.Right().Token.Name(), hashToString(swap.Right().Token.Hash()))
				fmt.Println("  Amount =", amountToString(swap.Right().Amount, swap.Right().Token.Decimals()))
			}
			fmt.Println("Active =", swap.Active())
		case "q":
			return
		default:
			fmt.Println("Unknown option", input)
		}
	}
}
