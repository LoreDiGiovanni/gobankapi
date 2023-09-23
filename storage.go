package main

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Storage interface {
    CreateAccount(*Account) (int64,error) 
    DeleteAccountByUUID(uuid.UUID) error
    DeleteAccountByID(int) error
    UpdateAccount(*Account) error
    GetAccountByUUID(uuid.UUID) (*Account,error)
    GetAccountByID(int64) (*Account,error)
    GetAccounts() ([]*Account,error)
}

type PostGresStore struct {
    db *sql.DB
}

func NewPostgresStore() (*PostGresStore,error){
    connStr := "user=postgres dbname=postgres password=gobankapi sslmode=disable"
	db, err := sql.Open("postgres", connStr)
    if err != nil{
        return nil,err
    }else{
        if err := db.Ping(); err != nil {
            return nil,err
        }
    }
    return &PostGresStore{db: db}, nil
}

func (s *PostGresStore) init() error{
    return s.CreateAccountTable()
}
func (s *PostGresStore) CreateAccountTable() error{
    query := `CREATE TABLE IF NOT EXISTS Account (
    id SERIAL PRIMARY KEY,
    id_uuid UUID NOT NULL,
    firstname VARCHAR(30) NOT NULL,
    lastname VARCHAR(30) NOT NULL,
    balance INT NOT NULL)`
    _,err := s.db.Exec(query);
    return err 
}

func (s *PostGresStore) CreateAccount(a * Account) (int64,error){
    query := "INSERT INTO Account(id_uuid,firstname,lastname,balance) VALUES($1,$2,$3,$4)"
    _,err := s.db.Exec(query,a.UUID,a.FirstName,a.LastName,a.Balance)
    if err!= nil {
        return 0,fmt.Errorf("Unable to create account")
    }else{
        var id int64
        row := s.db.QueryRow("SELECT id FROM Account WHERE id_uuid=$1",a.UUID)
        err := row.Scan(&id)
        if err != nil{
            return 0,fmt.Errorf("Unable to create account")
        }else {
            return id,nil
        }
    }
}


func (s *PostGresStore) DeleteAccountByID(id int64) error{
    _,err := s.db.Exec("DELETE FROM Account WHERE id=$1",id)
    if err!=nil {
        return err
    }else{
        return nil 
    }

}

func (s *PostGresStore) DeleteAccountByUUID(uuid uuid.UUID) error{
    _,err:=s.db.Exec("DELETE FROM Account WHERE id_uuid=$1",uuid)
    if err!=nil {
        return err
    }else{
        return nil 
    }
}

func (s *PostGresStore) UpdateAccount(* Account) error{
    return nil 
}

func (s *PostGresStore) GetAccountByUUID(uuid uuid.UUID) (*Account,error){
    rows,err:= s.db.Query("SELECT * FROM Account WHERE id_uuid=$1",uuid)
    if err!=nil{
        return nil, fmt.Errorf("Account %s not found",uuid.String())
    }else{
        rows.Next()
        a,err := RowToAccount(rows)
        if err != nil{
            return nil,fmt.Errorf("Account %d not found",uuid)
        }else{
            return a,nil
        }
    }
}

func (s *PostGresStore) GetAccountByID(id int64) (*Account,error){
    rows,err := s.db.Query("SELECT * FROM Account WHERE id=$1",id)
    if err!= nil {
        return nil, err
    }else{
        rows.Next()
        a,err := RowToAccount(rows)
        if err != nil{
            return nil,fmt.Errorf("Account %d not found",id)
        }else{
            return a,nil
        }
    }
}

func (s *PostGresStore) GetAccounts() ([]*Account,error){
    rows,err := s.db.Query("Select * from Account")
    if err != nil {
        return nil,err
    }else{
        accounts := []*Account{}
        for rows.Next(){
            a,err := RowToAccount(rows)
            if err != nil{
                return nil,err
            }else{
                accounts = append(accounts,a)
            }
        }
        return accounts,nil
    }
}

func RowToAccount(rows *sql.Rows) (*Account,error){
    a := &Account{}
    err:= rows.Scan(&a.ID,&a.UUID,&a.FirstName,&a.LastName,&a.Balance)
    return a,err
}
