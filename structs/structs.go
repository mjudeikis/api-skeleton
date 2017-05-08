package structs

const (

	//default KVstorage configuration
	DefaultStorageBackend = StorageBoltDB
	StorageBoltDB         = "boltdb"
	StorageConsul         = "consul"
	StorageETCD           = "etcd"
	DefaultConsulAddr     = "consul:8500"
	DefaultDatabasePath   = "/tmp/bolt.db"
)
