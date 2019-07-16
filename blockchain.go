package main

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"errors"
	"bytes"
	"crypto/ecdsa"
	"fmt"
)

const dbfile="blockchain.db"

const blockBucket = "blocks"
type blockchain struct {
	tip []byte
	db *bolt.DB
}

type blockchainiterator struct {
	currenthash []byte
	db *bolt.DB
}
func (bc *blockchain)minblock(ptransaction []*transaction)  *Block{

	for _, tx:= range ptransaction{
		if bc.verifytransaction(tx) != true{
			log.Panic("error: invalid transaction")
		}else {
			fmt.Println("verify success")
		}
	}
	var lasthash []byte
	var lastheight int32
	err:= bc.db.View(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blockBucket))
		lasthash= b.Get([]byte("l"))
		blockdata:= b.Get(lasthash)

		block := deserializeblock(blockdata)
		lastheight = block.Height
		return nil
	})
	if err!= nil {
		log.Panic(err)
	}
	pnewblock:= newblock(ptransaction,lasthash,lastheight +1 )
	bc.db.Update(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blockBucket))
		err:= b.Put(pnewblock.Hash,pnewblock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"),pnewblock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = pnewblock.Hash
		return nil
	})
	return pnewblock
}

const genisedata =  "haha blockchain"

func newblockchain(address string) *blockchain{
	var tip []byte
	db,err := bolt.Open(dbfile,0600,nil)
	if err != nil{
		log.Panic(err)
	}
	err= db.Update(func(tx *bolt.Tx) error {
		b:= tx.Bucket([]byte(blockBucket))
		if b==nil {
			println("区块链不存在，创建一个新的区块链")
			ptransaction:= newcoinbasetx(address,genisedata)

			genesis:= newgenisblock([]*transaction{ptransaction})
			b,err:= tx.CreateBucket([]byte(blockBucket))
			if err!=nil {
				log.Panic(err)
			}
			err= b.Put(genesis.Hash,genesis.Serialize())
			if err != nil{
				log.Panic(err)
			}
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		}else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc:= blockchain{tip,db}
	set:= utxoset{&bc}
	set.reindex()

	return &bc
}

func (bc *blockchain)iterator()  *blockchainiterator  {
	bci:= &blockchainiterator{bc.tip,bc.db}
	return bci
}

func (i *blockchainiterator) next()  *Block{
	var block *Block
	err:=i.db.View(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blockBucket))
		deblock:=b.Get(i.currenthash)
		block = deserializeblock(deblock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	i.currenthash = block.PrevBlockHash
	return block
}

func (bc *blockchain)printblockchain()  {
	bci:= bc.iterator()
	for ; ;  {
		block:=bci.next()
		block.String()
		if len(block.PrevBlockHash) ==0 {
			break
		}
	}
}
/*
func testblockdb()  {
	blockchain:= newblockchain()
	blockchain.minblock([]*transaction{})
	blockchain.minblock([]*transaction{})
	blockchain.printblockchain()
}
*/

func (bc *blockchain)findunspenttransaction(pubkeyhash []byte) []transaction {
	var unspenttx []transaction
	spendtxos := make(map[string][]int)
	bci:= bc.iterator()
	for {
		block:=bci.next()
		for _,tx:= range  block.Transactions{

			txid:=hex.EncodeToString(tx.ID)
		output:
			for  outidx,out := range tx.Vout {
				if spendtxos[txid] != nil {
					for _,spendout:=range spendtxos[txid] {
						if spendout == outidx{
							continue  output
						}
					}
				}
				if out.canbeunlockedwith(pubkeyhash)  {
					unspenttx = append(unspenttx,*tx)
				}
			}
			if tx.iscoinbase() == false  {
				for _,in:= range tx.Vin  {
					if in.canunlockoutputwith(pubkeyhash) {
						intxid:= hex.EncodeToString(in.Txid)
						spendtxos[intxid] = append(spendtxos[intxid],in.Voutindex)
					}
				}
			}
		}
		if len(block.PrevBlockHash) ==0 {
			break
		}
	}
	return unspenttx
}


func (bc  *blockchain)findutxo(pubkeyhash []byte) []txoutput {
	var utxos []txoutput
	unspendtransations :=bc.findunspenttransaction(pubkeyhash)
	for _, tx:= range unspendtransations{
		for _,out:= range tx.Vout{
			if out.canbeunlockedwith(pubkeyhash) {
				utxos = append(utxos,out)
			}
		}
	}
	return  utxos
}

func (bc  *blockchain) findspendableouputs(pubkeyhash []byte, amount int) (int , map[string][]int)  {
	unspentoutputs:= make(map[string][]int)
	unspenttxs:= bc.findunspenttransaction(pubkeyhash)
	accumulated := 0
	work:
	for _,tx:= range unspenttxs{
		txid:= hex.EncodeToString(tx.ID)
		for outidx, out := range tx.Vout {
			if out.canbeunlockedwith(pubkeyhash) && accumulated < amount{
				accumulated += out.Value
				unspentoutputs[txid] = append(unspentoutputs[txid], outidx)
				if accumulated >=  amount{
					break work
				}
			}
		}
	}
	return  accumulated, unspentoutputs
}

func (bc *blockchain) findtransationbyid(id []byte) (transaction, error	) {
	bci  := bc.iterator()
	for{
		block:= bci.next()
		for  _, tx:= range  block.Transactions {
			if bytes.Compare(tx.ID,id) == 0{
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return transaction{}, errors.New("Failed to not find transaction")
}
func (bc *blockchain) signtransation( tx *transaction,prikey ecdsa.PrivateKey)  {
	prevtxs:= make(map[string]transaction)
	for _, vin := range tx.Vin{
		prevtx, err:= bc.findtransationbyid(vin.Txid)
		if err!= nil {
			log.Panic(err)
		}

		prevtxs[hex.EncodeToString(prevtx.ID)] = prevtx
	}
	tx.sign(prikey, prevtxs)
}
func (bc *blockchain)verifytransaction(tx *transaction) bool  {
	prevtxs := make(map[string]transaction)
	for _,vin := range tx.Vin{
		prevtx, err:= bc.findtransationbyid(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevtxs[hex.EncodeToString(prevtx.ID)] = prevtx
	}
	return tx.verify(prevtxs)
}

func (bc *blockchain) findallutxo() map[string]txouputs  {
	utxo := make(map[string]txouputs)
	spendtxs := make(map[string][]int)
	bci := bc.iterator()

	for{

		block := bci.next()
		for _, tx:= range  block.Transactions{
			txid := hex.EncodeToString(tx.ID)
			outputs:
				for outidx, out := range tx.Vout{
					if spendtxs[txid] != nil{
						for  _, spendoutids := range  spendtxs[txid] {
							if spendoutids == outidx {
								continue  outputs
							}
						}
					}
					outs := utxo[txid]
					outs.Outputs = append(outs.Outputs, out)
					utxo[txid] = outs
				}
				if tx.iscoinbase() == false{
					for  _, in := range  tx.Vin {
						intxid := hex.EncodeToString(in.Txid)
						spendtxs[intxid] = append(spendtxs[intxid],in.Voutindex)
					}
				}
		}
		if len(block.PrevBlockHash) == 0{
			break
		}
	}
	return utxo
}

func (bc *blockchain)getblockheight() int32 {
	var lastblock Block
	err :=  bc.db.View(func(tx *bolt.Tx) error {
		b:= tx.Bucket([]byte(blockBucket))
		lasthash:= b.Get([]byte("l"))
		blockdata := b.Get(lasthash)
		lastblock = *deserializeblock(blockdata)
		return nil
	})
	if err != nil{
		log.Panic(err)
	}

	return  lastblock.Height
}
func (bc *blockchain)getblockhash() [][]byte  {
	var blocks [][]byte
	bci:= bc.iterator()
	for{
		block:= bci.next()
		blocks = append(blocks, block.Hash)
		if len(block.PrevBlockHash) == 0{
			break
		}
	}
	return blocks
}



func (bc *blockchain) getblock(blockhash []byte) (Block, error)  {
	var block Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		blockdata := b.Get(blockhash)
		if blockdata == nil {
			return  errors.New("block is not fund")
		}
		block = *deserializeblock(blockdata)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return  block, nil
}

func (bc  *blockchain) addblock( block  *Block)  {
	err :=  bc.db.Update(func(tx *bolt.Tx) error {
		b:= tx.Bucket([]byte(blockBucket))
		blockindb := b.Get(block.Hash)
		if blockindb != nil{
			return nil
		}
		blockdata:= block.Serialize()
		err :=  b.Put(block.Hash,blockdata)
		if err != nil{
			log.Panic(err)
		}
		lasthash:= b.Get([]byte("l"))
		lastblockdata := b.Get(lasthash)
		lastblock :=  deserializeblock(lastblockdata)

		if block.Height > lastblock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}
		return nil

	})
	if err != nil {
		log.Panic(err)
	}
}

