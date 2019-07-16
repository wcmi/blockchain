package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"encoding/gob"
	"crypto/elliptic"
	"bytes"
)

const walletFile = "wallet.dat"

type  Wallets struct{
	Walletsstore map[string] *Wallet
}

func newwallets() ( *Wallets, error)  {
	wallets := Wallets{}

	wallets.Walletsstore = make(map[string]*Wallet)

	err:= wallets.loadfromfile()
	return &wallets, err
}

func (ws  *Wallets)createwallet() string  {
	wallet:= newwallet()
	address:= fmt.Sprintf("%s",wallet.GetAddress())
	ws.Walletsstore[address] = wallet
	return  address
}
func (ws *Wallets)getwallet(address string)  Wallet{
	return *ws.Walletsstore[address]
}
func (ws * Wallets) getwalletsadress() []string {
	var arraddress []string
	for address,_:= range  ws.Walletsstore  {
		arraddress = append(arraddress,address )
	}
	return  arraddress
}

func (ws *Wallets) loadfromfile()  error {
	if _, err := os.Stat(walletFile);os.IsNotExist(err){
		return  err
	}
	filecontent ,err:= ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(filecontent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}
	ws.Walletsstore = wallets.Walletsstore
	return nil
}

func (ws *Wallets) savetofile()  {
	var  content  bytes.Buffer
	gob.Register(elliptic.P256())
	encoder:= gob.NewEncoder(&content)
	err:= encoder.Encode(ws)

	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile,content.Bytes(),0777)
	if err != nil {
		log.Panic(err)
	}
}

