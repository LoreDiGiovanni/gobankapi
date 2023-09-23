package main

import (
	"encoding/json"
	"fmt"
	"go/token"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type APIService struct{
    listenAddr string
    store Storage 
}

type ApiError struct{
    Error string `json:"error"`
}

type apiFunc func(w http.ResponseWriter,r *http.Request) error

func WriteJSON(w http.ResponseWriter, status int, v any) error {
    w.Header().Add("Content-Type","application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(v);
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        if err := f(w,r); err!=nil {
            WriteJSON(w,http.StatusBadRequest, ApiError{Error: err.Error()}) 
        }
    }
}

func NewAPIServer(listenAddr string, store Storage) *APIService {
    return &APIService{
        listenAddr: listenAddr,
        store: store,
    }
}
func (s * APIService) Run() {
    router := mux.NewRouter()
    router.HandleFunc("/accounts",makeHTTPHandleFunc(s.accountsHandler))
    router.HandleFunc("/account",makeHTTPHandleFunc(s.accountHandler))
    router.HandleFunc("/account/{id}",makeHTTPHandleFunc(s.accountHandlerByID))
    router.HandleFunc("/account/uuid/{uuid}",makeHTTPHandleFunc(s.accountHandlerByUUID))
    router.HandleFunc("/transfer",makeHTTPHandleFunc(s.transferHendelrByID))
    router.HandleFunc("/transfer/uuid",makeHTTPHandleFunc(s.transferHandlerByUUID))
    log.Println("JSON API server runnig on port: ",s.listenAddr)
    http.ListenAndServe(s.listenAddr,router)
}

func (s *APIService) accountsHandler(w http.ResponseWriter,r *http.Request) error {
    switch r.Method {
        case "GET": return s.handleGetAccounts(w,r);
        case "POST": return s.handleCreateAccount(w,r);
    }
    return fmt.Errorf("Method not allowed %s", r.Method);
} 

func (s *APIService) accountHandler(w http.ResponseWriter,r *http.Request) error {
    switch r.Method {
        case "POST": return s.handleCreateAccount(w,r);
        case "GET": return s.handleGetAccount(w,r);
    }
    return fmt.Errorf("Method not allowed %s", r.Method);
} 


func (s *APIService) accountHandlerByUUID(w http.ResponseWriter,r *http.Request) error {
    switch r.Method {
        case "GET": return s.handleGetAccountByUUID(w,r);
        case "DELETE": return s.handleDeleteAccountByUUID(w,r);
    }
    return fmt.Errorf("Method not allowed %s", r.Method);
}

func (s *APIService) accountHandlerByID(w http.ResponseWriter,r *http.Request) error {
    switch r.Method {
        case "GET": return s.handleGetAccountByID(w,r);
        case "DELETE": return s.handleDeleteAccountByID(w,r);
    }
    return fmt.Errorf("Method not allowed %s", r.Method);
}

func (s *APIService) transferHendelrByID(w http.ResponseWriter,r *http.Request) error {
    switch r.Method {
        case "POST": return s.handleTransfereByID(w,r);
    }
    return fmt.Errorf("Method not allowed %s", r.Method);
}

func (s *APIService) transferHandlerByUUID(w http.ResponseWriter,r *http.Request) error {
    switch r.Method {
        case "POST": return s.handleTransfereByUUID(w,r);
    }
    return fmt.Errorf("Method not allowed %s", r.Method);
}


func (s *APIService) handleGetAccountByUUID(w http.ResponseWriter,r *http.Request) error {
    uuidstr := mux.Vars(r)["uuid"]
    uuid,err := uuid.Parse(uuidstr)
    if err!= nil {
        return fmt.Errorf("This %s is not a valid id ",uuidstr)
    }else{
        a,err := s.store.GetAccountByUUID(uuid) 
        if err!= nil{
            return err
        }else{
            return WriteJSON(w,http.StatusOK,a)
        }
    }
}

func (s *APIService) handleGetAccountByID(w http.ResponseWriter,r *http.Request) error {
    idstr := mux.Vars(r)["id"]
    id,err := strconv.ParseInt(idstr,10,64)
    if err!= nil {
        return fmt.Errorf("This %s is not a valid id ",idstr)
    }else{
        a,err := s.store.GetAccountByID(id) 
        if err!= nil{
            return err
        }else{
            return WriteJSON(w,http.StatusOK,a)
        }
    }
}

func (s *APIService) handleGetAccount(w http.ResponseWriter,r *http.Request) error {
    token,err := decodeJWT(r.Header.Get("x-jwt-token"))
    claims := token.Claims.(jwt.MapClaims)
    id := claims["id"].(int64) 
    a,err := s.store.GetAccountByID(id) 
    if err!= nil{
        return err
    }else{
        return WriteJSON(w,http.StatusOK,a)
    }
}

func (s *APIService) handleGetAccounts(w http.ResponseWriter,r *http.Request) error {
    accounts, err:= s.store.GetAccounts()
    if err != nil {
        return err
    }else{
        return WriteJSON(w,http.StatusOK,accounts)
    }
}

func (s *APIService) handleCreateAccount(w http.ResponseWriter,r *http.Request) error {
    CreateAccountReq := CreateAccountRequest{}
    if err := json.NewDecoder(r.Body).Decode(&CreateAccountReq); err != nil {
        return err
    }
    defer r.Body.Close()

    account := NewAccount(CreateAccountReq.FirstName,CreateAccountReq.LastName)
    id, err := s.store.CreateAccount(account)
    account.ID = id;
    if err != nil {
        return err
    }
    token, err := createJWT(account)
    if err!= nil {
        return err
    }else{
        return WriteJSON(w,http.StatusOK,struct{Token string `json:"token"`}{Token: token}) 
    }
}

func (s *APIService) handleDeleteAccountByID(w http.ResponseWriter,r *http.Request) error {
    idstr := mux.Vars(r)["id"]
    id,err := strconv.Atoi(idstr)
    if err!= nil {
        return fmt.Errorf("This %s is not a valid id ",idstr)
    }else{
        err := s.store.DeleteAccountByID(id) 
        if err!= nil{
            return err
        }else{
            return WriteJSON(w,http.StatusOK,struct{ Deleted string `json:"deleted"`}{Deleted: idstr}) 
        }
    }
}

func (s *APIService) handleDeleteAccountByUUID(w http.ResponseWriter,r *http.Request) error {
    uuidstr := mux.Vars(r)["uuid"]
    uuid,err := uuid.Parse(uuidstr)
    if err!= nil {
        return fmt.Errorf("This %s is not a valid id ",uuidstr)
    }else{
        err := s.store.DeleteAccountByUUID(uuid) 
        if err!= nil{
            return err
        }else{
            return WriteJSON(w,http.StatusOK,struct{ Deleted string `json:"deleted"`}{Deleted: uuidstr}) 
        }
    }
}

func (s *APIService) handleTransfereByID(w http.ResponseWriter,r *http.Request) error {
    TranReq := TransfereRequest{}
    if err := json.NewDecoder(r.Body).Decode(&TranReq); err!=nil{
        return fmt.Errorf("transfer failed")
    }else{
        return WriteJSON(w,http.StatusAccepted,TranReq)
    }
}

func (s *APIService) handleTransfereByUUID(w http.ResponseWriter,r *http.Request) error {
    TranReq := TransfereRequest{}
    if err := json.NewDecoder(r.Body).Decode(&TranReq); err!=nil{
        return fmt.Errorf("transfer failed")
    }else{
        return WriteJSON(w,http.StatusAccepted,TranReq)
    }
}

func middlewareAuthJWT(functionHendler http.HandlerFunc) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        tokenString := r.Header.Get("x-jwt-token")
        token,err := validateJWT(tokenString)
        
        if err!= nil{
            WriteJSON(w,http.StatusForbidden, ApiError{Error: "Invalid authorization"})
        }else{
            if !token.Valid{
                WriteJSON(w,http.StatusForbidden, ApiError{Error: "Invalid authorization"})
            }else{
                functionHendler(w,r);
            }
        }
    }
}
func createJWT(account *Account) (string, error){
    claims:= &jwt.MapClaims{
        "uuid": account.UUID.String(),
        "id": account.ID,
    }
    secret:= []byte(os.Getenv("JWT_SECRET"))
    token:= jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
    return token.SignedString(secret)
} 

func validateJWT(tokenString string) (*jwt.Token,error){
    secret := os.Getenv("JWT_SECRET")
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		    return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	    }
	    return []byte(secret), nil
    })
}
func decodeJWT(tokenString string) (*jwt.Token,error){
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil})
     
}
