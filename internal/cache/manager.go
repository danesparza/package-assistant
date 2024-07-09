package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/newrelic/go-agent/v3/integrations/nrredis-v8"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"strings"
)

const PACKAGE_ASSISTANT_LOCK = "package_assistant_lock"

type Manager struct {
	rdb *redis.Client
	RS  *redsync.Redsync
}

func NewManager() (*Manager, error) {
	retval := new(Manager)

	err := *new(error)

	log.Debug().
		Str("Host", viper.GetString("redis.host")).
		Str("Port", viper.GetString("redis.port")).
		Bool("TLS", viper.GetBool("redis.tls")).
		Msg("Attempting to open redis connection")

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.host"), viper.GetString("redis.port")),
		Password: viper.GetString("redis.password"),
		DB:       0, // use default DB
	}

	rdb := redis.NewClient(opts)
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)

	//	If we need to use TLS, redefine our client
	if viper.GetBool("redis.tls") {

		log.Info().Msg("Using TLS with Redis")

		opts = &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.host"), viper.GetString("redis.port")),
			Password: viper.GetString("redis.password"),
			DB:       0, // use default DB
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
		rdb = redis.NewClient(opts)
	}

	//	Add a hook for new relic integration
	rdb.AddHook(nrredis.NewHook(opts))

	result, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return retval, fmt.Errorf("error trying to open a connection to redis: %v", err)
	}

	log.Debug().Str("result", result).Msg("Result from Redis")

	retval.rdb = rdb
	retval.RS = rs

	return retval, nil

}

// GetKey returns a key to be used in the cache system
func GetKey(entityType string, keyPart ...string) string {
	allparts := []string{}
	allparts = append(allparts, entityType)
	allparts = append(allparts, keyPart...)
	return strings.Join(allparts, ":")
}
