package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

var accounts = make(map[int]decimal.Decimal)

func GetAccounsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		id, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		balance, ok := accounts[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"id": "%d", "balance": %s}`, id, balance.StringFixed(2))
		return
	}
	list := []string{}
	for accId, _ := range accounts {
		list = append(list, string(accId))
	}
	bytes, err := json.Marshal(accounts)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"accounts": %s}`, bytes)
}

func AccountHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method == "GET" {
		GetAccounsHandler(w, r)
		return
	}
	// POST
	balance := r.FormValue("balance")
	if balance == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "\"blanace\" variable is required"}`)
		return
	}
	bal, err := decimal.NewFromString(balance)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": %s}`, err)
		return
	}
	accounts[len(accounts)+1] = bal
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"id": %d, "balance": %s}`, len(accounts), bal.StringFixed(2))
}

func IndexHtmlHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("index.html")
	if err != nil {
		panic("There is no index.html file")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, string(data))
}

func SendMoneyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	amount, err := decimal.NewFromString(r.FormValue("amount"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "\"blanace\" decimal variable is required"}`)
		return
	}
	// var balanceTo decimal.Decimal
	idFrom, err := strconv.Atoi(r.FormValue("id_from"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "\"id_from\" int variable is required"}`)
		return
	}
	idTo, err := strconv.Atoi(r.FormValue("id_to"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "\"id_to\" int variable is required"}`)
		return
	}
	if _, ok := accounts[idFrom]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "id_from account doesn't exist"}`)
		return
	}
	if _, ok := accounts[idTo]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "id_to account doesn't exist"}`)
		return
	}

	if accounts[idFrom].Sub(amount).Cmp(decimal.Zero) == -1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "Not enought money to do the transfer"}`)
		return
	}
	accounts[idFrom] = accounts[idFrom].Sub(amount)
	accounts[idTo] = accounts[idTo].Add(amount)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{}")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/accounts/{id:[0-9]+}/", AccountHandler).Methods("GET")
	r.HandleFunc("/accounts/", AccountHandler).Methods("POST", "GET")
	r.HandleFunc("/send_money/", SendMoneyHandler).Methods("POST")
	r.HandleFunc("/", IndexHtmlHandler)
	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
