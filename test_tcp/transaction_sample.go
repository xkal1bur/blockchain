package main

type Tx struct {
	Version   uint32
	TxIns     []TxIn
	TxOuts    []TxOut
	Locktimei uint32
}
type TxIn struct {
	PrevTx    []byte
	PrevIndex uint32
	Sequence  uint32
	Witness   [][]byte
	Net       string
}
type TxOut struct {
	Amount uint64
}
