package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/logic"
	"github.com/gravitl/netmaker/models"
	"github.com/gravitl/netmaker/mq"
)

// swagger:route POST /api/nodes/{network}/{nodeid}/createrelay nodes createRelay
//
// Create a relay.
//
//			Schemes: https
//
//			Security:
//	  		oauth
//
//			Responses:
//				200: nodeResponse
func createRelay(w http.ResponseWriter, r *http.Request) {
	var relay models.RelayRequest
	var params = mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&relay)
	if err != nil {
		logger.Log(0, r.Header.Get("user"), "error decoding request body: ", err.Error())
		logic.ReturnErrorResponse(w, r, logic.FormatError(err, "badrequest"))
		return
	}
	relay.NetID = params["network"]
	relay.NodeID = params["nodeid"]
	updatenodes, node, err := logic.CreateRelay(relay)
	if err != nil {
		logger.Log(0, r.Header.Get("user"),
			fmt.Sprintf("failed to create relay on node [%s] on network [%s]: %v", relay.NodeID, relay.NetID, err))
		logic.ReturnErrorResponse(w, r, logic.FormatError(err, "internal"))
		return
	}

	logger.Log(1, r.Header.Get("user"), "created relay on node", relay.NodeID, "on network", relay.NetID)
	for _, relayedNode := range updatenodes {

		err = mq.NodeUpdate(&relayedNode)
		if err != nil {
			logger.Log(1, "error sending update to relayed node ", relayedNode.ID.String(), "on network", relay.NetID, ": ", err.Error())
		}
	}

	apiNode := node.ConvertToAPINode()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiNode)
	runUpdates(&node, true)
}

// swagger:route DELETE /api/nodes/{network}/{nodeid}/deleterelay nodes deleteRelay
//
// Remove a relay.
//
//			Schemes: https
//
//			Security:
//	  		oauth
//
//			Responses:
//				200: nodeResponse
func deleteRelay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	nodeid := params["nodeid"]
	netid := params["network"]
	updatenodes, node, err := logic.DeleteRelay(netid, nodeid)
	if err != nil {
		logger.Log(0, r.Header.Get("user"), "error decoding request body: ", err.Error())
		logic.ReturnErrorResponse(w, r, logic.FormatError(err, "badrequest"))
		return
	}
	logger.Log(1, r.Header.Get("user"), "deleted relay server", nodeid, "on network", netid)
	for _, relayedNode := range updatenodes {
		err = mq.NodeUpdate(&relayedNode)
		if err != nil {
			logger.Log(1, "error sending update to relayed node ", relayedNode.ID.String(), "on network", netid, ": ", err.Error())
		}
	}
	apiNode := node.ConvertToAPINode()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiNode)
	runUpdates(&node, true)
}

// swagger:route POST /api/hosts/{hostid}/relay hosts createHostRelay
//
// Create a relay.
//
//			Schemes: https
//
//			Security:
//	  		oauth
//
//			Responses:
//				200: nodeResponse
func createHostRelay(w http.ResponseWriter, r *http.Request) {
	var relay models.HostRelayRequest
	var params = mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&relay)
	if err != nil {
		logger.Log(0, r.Header.Get("user"), "error decoding request body: ", err.Error())
		logic.ReturnErrorResponse(w, r, logic.FormatError(err, "badrequest"))
		return
	}
	relay.HostID = params["hostid"]
	relayHost, _, err := logic.CreateHostRelay(relay)
	if err != nil {
		logger.Log(0, r.Header.Get("user"),
			fmt.Sprintf("failed to create relay on host [%s]: %v", relay.HostID, err))
		logic.ReturnErrorResponse(w, r, logic.FormatError(err, "internal"))
		return
	}

	logger.Log(1, r.Header.Get("user"), "created relay on host", relay.HostID)
	// for _, relayedHost := range relayedHosts {

	// 	err = mq.PublishSingleHostUpdate(&relayedHost)
	// 	if err != nil {
	// 		logger.Log(1, "error sending update to relayed host ", relayedHost.ID.String(), ": ", err.Error())
	// 	}
	// }
	// // publish host update for relayhost
	// err = mq.PublishSingleHostUpdate(relayHost)
	// if err != nil {
	// 	logger.Log(1, "error sending update to relay host ", relayHost.ID.String(), ": ", err.Error())
	// }
	go func() {
		if err := mq.PublishPeerUpdate(); err != nil {
			logger.Log(0, "fail to publish peer update: ", err.Error())
		}
	}()

	apiHostData := relayHost.ConvertNMHostToAPI()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiHostData)
}

// swagger:route DELETE /api/hosts/{hostid}/relay hosts deleteHostRelay
//
// Remove a relay.
//
//			Schemes: https
//
//			Security:
//	  		oauth
//
//			Responses:
//				200: nodeResponse
func deleteHostRelay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	hostid := params["hostid"]
	relayHost, _, err := logic.DeleteHostRelay(hostid)
	if err != nil {
		logger.Log(0, r.Header.Get("user"), "error decoding request body: ", err.Error())
		logic.ReturnErrorResponse(w, r, logic.FormatError(err, "badrequest"))
		return
	}
	logger.Log(1, r.Header.Get("user"), "deleted relay host", hostid)
	// for _, relayedHost := range relayedHosts {
	// 	err = mq.PublishSingleHostUpdate(&relayedHost)
	// 	if err != nil {
	// 		logger.Log(1, "error sending update to relayed host ", relayedHost.ID.String(), ": ", err.Error())
	// 	}
	// }
	// err = mq.PublishSingleHostUpdate(relayHost)
	// if err != nil {
	// 	logger.Log(1, "error sending update to relayed host ", relayHost.ID.String(), ": ", err.Error())
	// }
	go func() {
		if err := mq.PublishPeerUpdate(); err != nil {
			logger.Log(0, "fail to publish peer update: ", err.Error())
		}
	}()
	apiHostData := relayHost.ConvertNMHostToAPI()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiHostData)
}
