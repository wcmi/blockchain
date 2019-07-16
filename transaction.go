package main

import (
	"fmt"
	"strings"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"encoding/hex"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
	"math/big"
)

const subsidy = 100

type transaction struct {
	ID   []byte
	Vin  []txinput
	Vout []txoutput
}
type txinput struct {
	Txid      []byte
	Voutindex int
	Signature []byte
	Pubkey []byte
}
type txoutput struct {
	Value      int
	Pubkeyhash []byte
}

type txouputs struct{
	Outputs []txoutput
}

func (outs txouputs)Serialize() []byte  {
	var buff  bytes.Buffer
	enc := gob.NewEncoder(&buff)

	err:= enc.Encode(outs)
	fmt.Println(outs)
	if err != nil {
		log.Panic(err)
	}
	return  buff.Bytes()
}

func deserializeoutputs(data []byte) txouputs {
	var outputs txouputs
	dec := gob.NewDecoder(bytes.NewReader(data))
	err:= dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}
	return outputs
}

func (tx *transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("---transaction %x:", tx.ID))
	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Voutindex))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.Pubkeyhash))
	}

	return strings.Join(lines, "\n")
}

func (tx transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		fmt.Println("--------\n")
		log.Panic(err)
	}
	fmt.Printf("%x\n", encoded.Bytes())
	return encoded.Bytes()
}
func (tx *transaction) hash() []byte {
	txcopy := *tx
	txcopy.ID = []byte{}
	hash := sha256.Sum256(txcopy.Serialize())
	return hash[:]
}
func newtxoutput(value int, address string) *txoutput {
	txo := &txoutput{value, nil}
	//txo.Pubkeyhash = []byte(address)
	txo.lock([]byte(address))
	return txo
}
func (out *txoutput)lock(address []byte)  {
	decodeaddress := Base58Decode(address)
	pubkeyhash := decodeaddress[1:len(decodeaddress) - 4]
	out.Pubkeyhash = pubkeyhash
}
func newcoinbasetx(to,data  string) *transaction {
	txin := txinput{[]byte{}, -1, nil,[]byte(data)}
	txout := newtxoutput(subsidy, to)
	tx := transaction{nil, []txinput{txin}, []txoutput{*txout}}
	tx.ID = tx.hash()
	return &tx
}
func (out *txoutput) canbeunlockedwith(pubkeyhash []byte) bool  {
	return bytes.Compare(out.Pubkeyhash,pubkeyhash) == 0
}

func (in *txinput) canunlockoutputwith(unlockdata  []byte) bool {
	lockinghash:= hashpubkey(in.Pubkey)
	return bytes.Compare(lockinghash, unlockdata) == 0
}

func (tx transaction)iscoinbase() bool  {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Voutindex == -1
}

func newutxotransaction(from, to string, amount int , bc *blockchain) *transaction {
	var inputs []txinput
	var outputs []txoutput

	wallets, err:= newwallets()
	if err != nil {
		log.Panic(err)
	}
	wallet:= wallets.getwallet(from)
	acc, validoutputs:= bc.findspendableouputs(hashpubkey(wallet.Publickey),amount)
	if acc< amount{
		log.Panic("error: not enough funds")
	}
	for txid, outs:= range validoutputs {
		txid , err:= hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out:= range outs {
			input := txinput{txid,out, nil,wallet.Publickey}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs,*newtxoutput(amount,to))
	if acc > amount {
		outputs = append(outputs, *newtxoutput(acc-amount,from))
	}
	tx:= transaction{nil, inputs, outputs}
	tx.ID = tx.hash()
	bc.signtransation(&tx, wallet.PrivateKey)
	return  &tx
}

func (tx *transaction) trimmedcopy()  transaction  {
	var inputs []txinput
	var outputs []txoutput

	for _, vin := range  tx.Vin{
		inputs = append(inputs, txinput{vin.Txid,vin.Voutindex,nil,nil})
	}
	txcopy := transaction{tx.ID, inputs, outputs}
	return txcopy
}
func (tx *transaction) sign(privkey  ecdsa.PrivateKey, prevtxs map[string]transaction)  {
	if tx.iscoinbase() {
		return
	}
	for _, vin := range tx.Vin {
		if prevtxs[hex.EncodeToString(vin.Txid)].ID == nil{
			log.Panic("error")
		}
	}
	txcopy := tx.trimmedcopy()
	for inId, vin:= range txcopy.Vin {
		prevtx:= prevtxs[hex.EncodeToString(vin.Txid)]
		txcopy.Vin[inId].Signature = nil
		txcopy.Vin[inId].Pubkey = prevtx.Vout[vin.Voutindex].Pubkeyhash
		txcopy.ID = txcopy.hash()
		r,s, err:= ecdsa.Sign(rand.Reader, &privkey, txcopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inId].Signature = signature
		txcopy.Vin[inId].Pubkey = nil
	}
}

func (tx *transaction) verify(prevtxs map[string]transaction) bool  {
	if tx.iscoinbase() {
		return true
	}
	for _,vin := range  tx.Vin {
		if prevtxs[hex.EncodeToString(vin.Txid)].ID == nil{
			log.Panic("error")
		}
	}
	txcopy := tx.trimmedcopy()
	curve := elliptic.P256()
	for inId, vin :=  range  tx.Vin  {
		prevtx := prevtxs[hex.EncodeToString(vin.Txid)]
		txcopy.Vin[inId].Signature = nil
		txcopy.Vin[inId].Pubkey = prevtx.Vout[vin.Voutindex].Pubkeyhash
		txcopy.ID = txcopy.hash()

		r:=  big.Int{}
		s:= big.Int{}
		siglen:= len(vin.Signature)
		r.SetBytes(vin.Signature[:(siglen/2)])
		s.SetBytes(vin.Signature[(siglen/2):])

		x:= big.Int{}
		y:= big.Int{}
		keylen := len(vin.Pubkey)
		x.SetBytes(vin.Pubkey[:(keylen/2)])
		y.SetBytes(vin.Pubkey[(keylen)/2:])
		rawpubkey := ecdsa.PublicKey{curve, &x, &y}

		if ecdsa.Verify(&rawpubkey,txcopy.ID,&r, &s) == false{
			return  false
		}
		txcopy.Vin[inId].Pubkey = nil
	}
	return true
}