package main

import (
	"os"

	"github.com/hectorchu/gonano/wallet"
)

// Wallet is a simple wallet.
type Wallet struct {
	seed []byte
	w    *wallet.Wallet
	a    *wallet.Account
}

func newWallet(seed []byte, rpcURL string) (w *Wallet, err error) {
	w = &Wallet{seed: seed}
	if w.w, err = wallet.NewWallet(seed); err != nil {
		return
	}
	w.w.RPC.URL = rpcURL
	w.a, err = w.w.NewAccount(nil)
	return
}

func loadWallet(rpcURL string) (w *Wallet, err error) {
	f, err := os.Open("wallet.bin")
	if err != nil {
		return
	}
	defer f.Close()
	seed := make([]byte, 32)
	if _, err = f.Read(seed); err != nil {
		return
	}
	return newWallet(seed, rpcURL)
}

func (w *Wallet) save() (err error) {
	f, err := os.Create("wallet.bin")
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(w.seed)
	return
}

func (w *Wallet) account() *wallet.Account {
	return w.a
}

func (w *Wallet) changeAccount(index uint32) (err error) {
	w.a, err = w.w.NewAccount(&index)
	return
}
