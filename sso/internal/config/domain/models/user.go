package models // пакет с с данными, к которому имеют доступ сервисный слой и слой раюоты с базой данных, это нужно для упрощения кода

type User struct {
	ID       int64
	Email    string
	PassHash []byte
}
