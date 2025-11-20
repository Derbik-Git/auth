package sqlite

import (
	"GRPC/sso/internal/config/domain/models"
	"GRPC/sso/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type Storage struct { // СМОТРИ СЮДА!!!!!!! ТЕБЕ ОЧЕНЬ ВАЖНО ПОСМОТРЕТЬ РОЛИК У НИКОЛАЯ ТУЗОВА КАК РАБОТАТЬ С SQLITE ИМЕННО В ГО, ОЧЕНЬ ПОЛЕЗНЫЙ И ВАЖНЫЙ РОЛИК
	db *sql.DB
}

// User implements auth1.UserProvider.
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	panic("unimplemented")
}

// App implements auth1.AppProvider.
func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	panic("unimplemented")
}

func NewStorage(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"

	// при создании базы данных указываем путь до самой бд
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passwordHash []byte) (int64, error) {
	const op = "storge.sqlite.SaveUSer"

	stmt, err := s.db.Prepare("INSERT INTO users (email, pass_hash) VALUSES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, email, passwordHash)
	if err != nil { // вообще эту конструкию чтоит запомнить
		var sqliteErr sqlite3.Error // еременная для проверки конкретного типа ошибки
		// 1. Посмотри на код ошибки: sqliteErr.ExtendedCode — это код ошибки, который мы получили от базы данных. 2. Сравни его с известным кодом: Мы сравниваем этот код с sqlite3.ErrConstraintUnique, который мы знаем — это ошибка уникальности. 3. Если они совпадают: Если коды совпадают, это значит, что произошла именно ошибка уникальности.
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique { // Откуда берется ExtendedCode у sqliteErr? Когда мы используем errors.As(err, &sqliteErr), мы пытаемся привести ошибку err к типу sqlite3.Error. Если это преобразование успешно, переменная sqliteErr будет содержать все поля и методы, определенные в структуре sqlite3.Error, включая поле ExtendedCode. Таким образом, если ошибка действительно является ошибкой SQLite, мы можем получить доступ к этому полю и использовать его для дальнейшей обработки. sqlite3.ErrConstraintUnique — это конкретный код ошибки, который говорит о том, что вы пытаетесь добавить или изменить запись в базе данных, но уже существует запись с таким же значением в поле, которое должно быть уникальным.
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists) // ошибка, созданная нами, о том что пользователь уже существует
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Получаем id созданной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s, %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUser(ctx context.Context, email string) (models.User, error) { // ТУТ КСТАТИ ПРОХОДИТ ВАЖНАЯ ЛОГИКА ДЛЯ ПРИЛОЖЕНИЯ, О ТОМ КАК ДАННЫЕ ИЗ РЕЗУЛЬТАТА ЗАПРОСА БУДУТ ПОПАДАТЬ И СОХРАНЯТСЯ В СТРУКТУРУ С ДАННЫМИ О ПОЛЬЗОВАТЕЛЕ
	// конструкция if email == "" {} не нужна так как мы потом приводим ошибку к типу, который отвечает за отсутствие строки и обрабатываем её
	const op = "storage.sqlite.GetUser"

	stmt, err := s.db.Prepare("SELECT id,email, pass_hash FROM users WHERE email = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email) // этот метод возвращает объект, который мы можем использовать для получения данных из результата запроса, с помощью метода row.Scan

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // sql.ErrNoRows — это ошибка, которая указывает на то, что результат запроса пуст.
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"

	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool
	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (s *Storage) GetApp(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.sqlite.GetApp"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
