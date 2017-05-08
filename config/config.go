package config

import (
	"os"

	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv/store/consul"
	"github.com/mangirdaz/api-skeleton/kv"
	"github.com/mangirdaz/api-skeleton/structs"
)

//default values for application
const (
	//api const
	DefaultAPIPort = "8000"
	DefaultAPIIP   = "0.0.0.0"
	//basic auth enabled or not
	DefaultBasicAuthentication = false

	//ENV
	EnvAPIPort      = "API_PORT"
	EnvAPIIP        = "API_IP"
	EnvBasicAuth    = "API_BASIC_AUTH"
	EnvConsulAddr   = "CONSUL_ADDR"
	EnvDatabasePath = "DB_PATH"
)

//Options structures for application default and configuration
type Options struct {
	Default     string
	Environment string
}

// Get - gets specified variable from either environment or default one
func Get(variable string) string {

	var config = map[string]Options{
		"EnvAPIPort": {
			Default:     DefaultAPIPort,
			Environment: EnvAPIPort,
		},
		"EnvAPIIP": {
			Default:     DefaultAPIIP,
			Environment: EnvAPIIP,
		},
		"EnvBasicAuth": {
			Default:     strconv.FormatBool(DefaultBasicAuthentication),
			Environment: EnvBasicAuth,
		},
		"EnvDefaultConsulAddr": {
			Default:     structs.DefaultConsulAddr,
			Environment: EnvConsulAddr,
		},
		"EnvDatabasePath": {
			Default:     structs.DefaultDatabasePath,
			Environment: EnvDatabasePath,
		},
	}

	for k, v := range config {
		if k == variable {
			if os.Getenv(v.Environment) != "" {
				log.WithFields(log.Fields{
					"key":   k,
					"value": v.Environment,
				}).Debug("config: setting configuration")
				return os.Getenv(v.Environment)
			}
			log.WithFields(log.Fields{
				"key":   k,
				"value": v.Default,
			}).Debug("config: setting configuration")
			return v.Default

		}
	}
	return ""
}

//init default storage client
func InitKVStorage() *kv.LibKVBackend {

	switch storagebackend := Get("EnvDefaultKVBackend"); storagebackend {
	case structs.StorageBoltDB:
		return generateDefaultStorageBackend()
	case structs.StorageConsul:
		consul.Register()
		db := Get("EnvDefaultConsulAddr")
		kv, err := kv.NewLibKVBackend(structs.StorageConsul, "default", []string{db})
		if err != nil {
			log.WithFields(log.Fields{
				"path":  db,
				"error": err,
			}).Fatal("failed to create new consult connection")
		}
		return kv
	default:
		return generateDefaultStorageBackend()
	}
}

func generateDefaultStorageBackend() *kv.LibKVBackend {
	db := Get("EnvDatabasePath")
	kv, err := kv.NewLibKVBackend(structs.DefaultStorageBackend, "default", []string{db})
	if err != nil {
		log.WithFields(log.Fields{
			"path":  db,
			"error": err,
		}).Fatal("failed to create new database")
	}
	return kv
}
