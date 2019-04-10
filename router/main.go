package router

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	cfg "github.com/vulpemventures/nigiri-chopsticks/config"
	"github.com/vulpemventures/nigiri-chopsticks/faucet"
	"github.com/vulpemventures/nigiri-chopsticks/faucet/liquid"
	"github.com/vulpemventures/nigiri-chopsticks/faucet/regtest"
	"github.com/vulpemventures/nigiri-chopsticks/helpers"
	"github.com/vulpemventures/nigiri-chopsticks/router/middleware"
)

// Router extends gorilla Router
type Router struct {
	*mux.Router
	Config    cfg.Config
	RPCClient *helpers.RpcClient
	Faucet    faucet.Faucet
}

var faucets = map[string]func(chain string, client *helpers.RpcClient) faucet.Faucet{
	"bitcoin": regtestfaucet.NewFaucet,
	"liquid":  liquidfaucet.NewFaucet,
}

// NewRouter returns a new Router instance
func NewRouter(config cfg.Config) *Router {
	router := mux.NewRouter().StrictSlash(true)
	rpcClient, _ := helpers.NewRpcClient(config.RPCServerURL(), false, 10)

	r := &Router{router, config, rpcClient, nil}

	if r.Config.IsFaucetEnabled() {
		url := r.Config.RPCServerURL()
		chain := r.Config.Chain()
		r.Faucet = faucets[chain](url, rpcClient)
		r.HandleFunc("/faucet", r.HandleFaucetRequest).Methods("POST")

		status, blockHashes, err := r.Faucet.Fund()
		if err != nil {
			log.WithError(err).WithField("status", status).Warning("Could not be able to fund faucet, please do it manually")
		} else {
			if len(blockHashes) > 0 {
				log.WithField("num_blocks", len(blockHashes)).Info("Faucet has been funded mining some blocks")
			}
		}
	}

	if config.IsLoggerEnabled() {
		r.Use(middleware.Logger)
	}
	r.HandleFunc("/broadcast", r.HandleBroadcastRequest).Methods("POST")
	r.PathPrefix("/").HandlerFunc(r.HandleElectrsRequest)

	return r
}
