package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func CreateLogFile() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
	} else {
		log.SetOutput(os.Stdout)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

}

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :adress, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("adress", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		log.Println("Error adding new parcel:", err)
		return 0, err
	}

	// верните идентификатор последней добавленной записи
	lastId, err := (res.LastInsertId())
	if err != nil {
		log.Println("Error add new parcel:", err)
		return 0, err
	}
	return int(lastId), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number", sql.Named("number", number))
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		log.Printf("Error get parcel by number: parcel number %d\n%v", number, err)
		return p, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк

	// заполните срез Parcel данными из таблицы
	var res []Parcel
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client", sql.Named("client", client))
	if err != nil {
		log.Printf("Error get parsel by client: client %d\n%v", client, err)
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		p := Parcel{}
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			log.Printf("Error get parsel by client: client %d\n%v", client, err)
			return res, err
		}
		res = append(res, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error get parsel by client: client %d\n%v", client, err)
		return res, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number", sql.Named("status", status), sql.Named("number", number))
	if err != nil {
		log.Printf("Error update parcel status: parcel number %d\n%v", number, err)
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	parsel, err := s.Get(number)
	if parsel.Status == ParcelStatusRegistered {
		_, err := s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number",
			sql.Named("address", address),
			sql.Named("number", number))
		if err != nil {
			log.Printf("Error update parcel address: parcel number %d\n%v", number, err)
			return err
		}
		fmt.Printf("У посылки № %d новый адрес: %s\n", number, address)
		return nil
	}
	fmt.Printf("Невозможно изменить адрес посылки № %d, статус посылки: %s.\n", parsel.Number, parsel.Status)
	return err
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	parsel, err := s.Get(number)
	if parsel.Status == ParcelStatusRegistered {
		_, err = s.db.Exec("DELETE FROM parcel WHERE number = :number", sql.Named("number", number))
		if err != nil {
			log.Printf("Error delete parcel: parcel number %d\n%v", number, err)
			return err
		}
		fmt.Printf("Посылка № %d успешно удалена.\n", parsel.Number)
		return nil
	}
	fmt.Printf("Невозможно удалить посылку № %d, статус посылки: %s.\n", parsel.Number, parsel.Status)
	return err
}

// Чистим БД
func ClearDB(s ParcelStore) error {
	_, err := s.db.Exec("DELETE FROM parcel")
	if err != nil {
		log.Printf("Error clear database: %v", err)
		return err
	}
	_, err = s.db.Exec("UPDATE SQLITE_SEQUENCE SET seq = 0 WHERE name = 'parcel'")
	if err != nil {
		log.Printf("Error clear database: %v", err)
		return err
	}
	return nil
}
