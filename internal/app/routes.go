package app

import (
	"fmt"
	"net/http"

	"net-commander-server/internal/repository"
)

const (
	prefix = "/api/v1"
)

func (a *App) loadRoutes() {
	a.router.HandleFunc(fmt.Sprintf("GET %s/subnet", prefix), http.HandlerFunc(a.getSubnets))
	a.router.HandleFunc(fmt.Sprintf("GET %s/interface", prefix), http.HandlerFunc(a.getInterfaces))
	a.router.HandleFunc(fmt.Sprintf("POST %s/subnet", prefix), http.HandlerFunc(a.createSubnet))
	// a.router.HandleFunc(fmt.Sprintf("GET %s/", prefix), http.HandlerFunc(a.createSubnet))
}

func (a *App) getInterfaces(w http.ResponseWriter, r *http.Request) {
	// repo := repository.New(a.db)
	// interfaces, err := repo.SelectInterface(r.Context())
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("failed to get interfaces: %v", err), http.StatusInternalServerError)
	// 	return
	// }

	// if err := json.NewEncoder(w).Encode(interfaces); err != nil {
	// 	http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	// 	return
	// }
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (a *App) createSubnet(w http.ResponseWriter, r *http.Request) {
	// repo := repository.New(a.db)

	// var intf repository.Interface
	// if err := json.NewDecoder(r.Body).Decode(&intf); err != nil {
	// 	http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
	// 	return
	// }

	// repo.CreateInterface(r.Context(), repository.CreateInterfaceParams{
	// 	Privatekey: intf.Privatekey,
	// 	Address:    intf.Address,
	// 	Listenport: intf.Listenport,
	// })

	w.Write([]byte(`{"status":"ok"}`))
}

func (a *App) getSubnets(w http.ResponseWriter, r *http.Request) {
	repo := repository.New(a.db)
	subnets, err := repo.GetSubnetsWithPeers(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get subnets: %v", err), http.StatusInternalServerError)
		return
	}

	responseJson := "["
	for i, subnet := range subnets {
		if i > 0 {
			responseJson += ","
		}
		responseJson += fmt.Sprintf(`{"interface_id": %d, "address": "%s", "listenPort": %d, "peer_id": %d, "name": "%s", "publicKey": "%s", "allowedIPs": "%s"}`,
			subnet.InterfaceID,
			subnet.Address,
			subnet.Listenport,
			subnet.PeerID,
			subnet.Name,
			subnet.Publickey,
			subnet.Allowedips.String,
		)

	}
	responseJson += "]"

	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(responseJson))
}
