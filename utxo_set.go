package main

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
)

type utxoset struct {
	bchain * blockchain
}
const utxobucket = "chainset"

func (u utxoset) reindex()  {
	db:= u.bchain.db
	bucketname := []byte(utxobucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err1:= tx.DeleteBucket(bucketname)
		if err1 != nil && err1 != bolt.ErrBucketNotFound {
			log.Panic(err1)
		}
		_,err2:= tx.CreateBucket(bucketname)
		if err2 != nil {
			log.Panic(err2)
		}
		return nil
	})
	if err != nil{
		log.Panic(err)
	}
	utxo := u.bchain.findallutxo()

	err4 := db.Update(func(tx *bolt.Tx) error{
		b:= tx.Bucket(bucketname)

		for txID,outs := range utxo{
			key,err5:= hex.DecodeString(txID)
			if err5!=nil{
				log.Panic(err5)
			}

			err6:=b.Put(key,outs.Serialize())
			if err6 !=nil{
				log.Panic(err6)
			}
		}
		return nil
	})

	if err4!=nil{
		log.Panic(err4)
	}
}

func (u utxoset) findutxobypubkeyhash(pubkeyhash []byte) []txoutput  {
	var utxos []txoutput
	db :=  u.bchain.db
	err:= db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxobucket))
		c:= b.Cursor()

		for k, v:= c.First();k!= nil;k,v = c.Next()  {
			outs := deserializeoutputs(v)
			for _,out := range outs.Outputs{
				if out.canbeunlockedwith(pubkeyhash) {
					utxos = append(utxos, out)
				}
			}
		}
		return nil
	})
	if err!= nil {
		log.Panic(err)
	}
	return utxos
}

func (u utxoset) update(block *Block)  {
	db := u.bchain.db
	err:= db.Update(func(tx *bolt.Tx) error {
		b:= tx.Bucket([]byte(utxobucket))
		for _, tx:= range block.Transactions{
			if tx.iscoinbase() == false{
				for _,vin := range  tx.Vin{
					updateouts := txouputs{}
					outsbytes:=  b.Get(vin.Txid)
					outs := deserializeoutputs(outsbytes)

					for outidx, out := range  outs.Outputs{
						if outidx != vin.Voutindex{
							updateouts.Outputs = append(updateouts.Outputs,out)
						}
					}
					if len(updateouts.Outputs) == 0{
						err:= b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					}else {
						err:= b.Put(vin.Txid,updateouts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}
			newoutputs := txouputs{}
			for _, out:= range  tx.Vout{
				newoutputs.Outputs = append(newoutputs.Outputs,out)
			}

			err:=  b.Put(tx.ID, newoutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}