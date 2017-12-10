package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	_ "github.com/mattn/go-sqlite3"
)

type IngosHandler struct {
	DB *sql.DB
}

type RegistrationRequest struct {
	Vin string `json:"vinCode"`
}

type RegistrationResponse struct {
	Address string `json:"address"`
	Tx      string `json:"tx"`
}

type OperationRequest struct {
	Sender     string `json:"from"`
	Recipient  string `json:"to"`
	Attachment string `json:"attribute"`
}

type OperationResponse struct {
	TxID       string `json:"tx"`
	Recipient  string `json:"to"`
	Sender     string `json:"from"`
	Attachment string `json:"attribute"`
}

// 200 - ok
// 201 - Create
// 400 - bad request
// 500 -

// присваиваем кошелек автомобиля для совершения различных операций
// ?VIN=98789799809809
// curl -v -X POST -H "Content-Type: application/json" -d '{"vin": "90238049832098409238"}' http://127.0.0.1:8080/api/v1/registration?VIN=09839280492830909
// curl -v -X POST -H "Content-Type: application/json" -d '{"vin": "90238049832098409238"}' http://77.244.213.45:8080/api/v1/registration

// POST JSON {vin: string}
// RESP {tx: string, address: string}
func (handler *IngosHandler) Registration(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	var params RegistrationRequest

	json.Unmarshal(body, &params)

	if params.Vin == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok, err := handler.checkVin(params.Vin)
	__err_panic(err)

	var wallet Wallet

	if !ok {
		wallet = *getAddress()
		err := handler.registrationVin(params.Vin, wallet.address)
		__err_panic(err)

		w.WriteHeader(http.StatusCreated)

	} else {
		address, _ := handler.getAddressByVin(params.Vin)
		wallet = *NewWallet(address)
		w.WriteHeader(http.StatusOK)

	}

	resultReg, err := wallet.transaction("CREATED")

	__err_panic(err)

	txParam := resultReg.(map[string]interface{})
	//txParam := resultParam["tx"].(map[string]interface{})

	log.Println(txParam)

	respParam := RegistrationResponse{
		Address: wallet.address,
		Tx:      txParam["id"].(string),
	}

	result, _ := json.Marshal(respParam)

	w.Write(result)

}

func JsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		w.Header().Add("Content-Type", "application/json")
		//		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		//		w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		next.ServeHTTP(w, r)
	})
}

// присваиваем кошелек автомобиля для совершения различных операций
// ?VIN=98789799809809
// curl -v -X POST -H "Content-Type: application/json" -d '{"to": "3PJp6xRMmxF65qs5CZkPauyM66tKBs6tp1r", "attribute": "test"}' http://127.0.0.1:8080/api/v1/operation
// POST JSON {from: sting, to: string, attribute: string}
// RESP {tx: string, address: string}
func (handler *IngosHandler) Operation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	var params OperationRequest

	json.Unmarshal(body, &params)

	if params.Recipient == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	wallet := NewWallet(params.Recipient)

	result, err := wallet.transaction(params.Attachment)

	__err_panic(err)

	txParam := result.(map[string]interface{})
	//txParam := resultParam["id"].(map[string]interface{})

	log.Println(txParam)

	respParam := OperationResponse{
		TxID:       txParam["id"].(string),
		Recipient:  txParam["recipient"].(string),
		Sender:     txParam["sender"].(string),
		Attachment: params.Attachment}
	w.WriteHeader(http.StatusOK)

	resultJson, _ := json.Marshal(respParam)

	//w.Header().Add("Content-Length", strconv.Itoa(len(result)))

	w.Write(resultJson)

}

// запрашиваем историю изменений по автомобилю
// ?VIN=98789799809809
// ?address=3PAUs5TpuHBBhZ6iNtBtXX1RjAsT84Tfu5H
func (handler *IngosHandler) Events(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// определяем параметры vin номера
	address := r.FormValue("address")

	if address == "" {
		vin := r.FormValue("VIN")
		address, _ = handler.getAddressByVin(vin)
	}

	wallet := &Wallet{
		address: address,
	}

	wallet.events()
}

func NewWaves(db *sql.DB) (http.Handler, error) {

	// определяем хендлег
	ingosHandler := &IngosHandler{
		DB: db,
	}

	router := httprouter.New()
	router.POST("/api/v1/registration", ingosHandler.Registration)
	router.POST("/api/v1/events", ingosHandler.Events)
	router.POST("/api/v1/operation", ingosHandler.Operation)

	return router, nil

}

func main() {
	db, err := sql.Open("sqlite3", "./tasks.db")
	err = db.Ping() // вот тут будет первое подключение к базе
	__err_panic(err)

	handler, err := NewWaves(db)
	if err != nil {
		panic(err)
	}

	cors := JsonMiddleware(handler)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", cors)

}

// не используйте такой код в прошакшене
// ошибка должна всегда явно обрабатываться
func __err_panic(err error) {

	if err != nil {
		panic(err)
	}
}
