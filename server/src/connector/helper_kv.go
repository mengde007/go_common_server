package connector

import (
	gp "code.google.com/p/goprotobuf/proto"
	"dbclient"
)

func KVQueryBase(table, uid string, value gp.Message) (exist bool, err error) {
	return dbclient.KVQueryBase(table, uid, value)
}

func KVWriteBase(table, uid string, value gp.Message) (result bool, err error) {
	return dbclient.KVWriteBase(table, uid, value)
}

func KVDeleteBase(table, uid string) (result bool, err error) {
	return dbclient.KVDeleteBase(table, uid)
}

func KVQueryExt(table, uid string, value gp.Message) (exist bool, err error) {
	return dbclient.KVQueryExt(table, uid, value)
}

func KVWriteExt(table, uid string, value gp.Message) (result bool, err error) {
	return dbclient.KVWriteExt(table, uid, value)
}

func KVDeleteExt(table, uid string) (result bool, err error) {
	return dbclient.KVDeleteExt(table, uid)
}
