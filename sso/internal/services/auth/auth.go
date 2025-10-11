package auth1 // бизнес логика

import (
	"GRPC/sso/internal/app/grpc/lib/jwt"
	"GRPC/sso/internal/config/domain/models"
	"GRPC/sso/internal/lib/logger/sl"
	"GRPC/sso/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

// Register implements auth.Auth.
func (a *Auth) Register(ctx context.Context, email string, password string) (userID int64, err error) {
	panic("unimplemented")
}

// это интерфейс, который описывает методы, которые должны быть реализованы в Auth (это методы для непрямого взаимодействия с БД, ну обычный сервисный слой в GRPC(метод луковицы))

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passwordHash []byte) (uid int64, err error) // ОЧЕНЬ ВАЖНО!!!, здесь мы имеем принимаемых параметр в качестве массива байт, как это работает и для чего сделано сейчас объясню, в методах, реализующих бизнес логику приложения одна из функций принимает пароль от пользователя как строку и уже в функции конвертирует его в массив байт, далее этот массив байт сравнивается с тем, что уже есть в БД, и если такой существует, то всё хорошо и мы выдаём пользователю токен, по которому он может заходить в другие привязанные приложения без регистрации(для этого мы и создаём наш сервер аутентификации) а нужно это для безопасности, таким образом мы солим пароль и злоумышленики не смогут его прочитать, это всё что я щас сказал, работает при входа пользователя в систему, что логично, так как что бы использовать данную фичу, массив байт(пароль) уже должен имется в бд ещё с регистрации пользователя, соответственно он уже должен быть зарегестрирован, если он попытается войти, без регистрации, то система не найдёт такого пароля и возникнет ошибка
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error) // чтение данных о пользователе (GET)
	IsAdmin(ctx context.Context, userID int64) (bool, error)     // определение прав пользователя
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

// New retursn a new instance of the Auth service.
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	//New(log, storage, storage, storage, time.Hour) (я так понял это вызов метода New)

	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) { // ТЫК!!! ТЫК!!!   Передача email в метод User из интерфейса userProvider необходима для того, чтобы получить данные о пользователе, который пытается войти в систему. Эта функция не должна соответствовать интерфейсу аутентификации, поскольку она выполняет более низкоуровневую операцию — получение данных о пользователе. Ваша функция Login может использовать другие методы или структуры для выполнения аутентификации, но для получения информации о пользователе она обращается к слою работы с базой данных через интерфейс userProvider.
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email)) // опять же повторюсь, если приложение пойдёт в прод, то личные данные пользователя лучше скрыть

	user, err := a.userProvider.User(ctx, email) // бизнес логика(то есть наш сервис) туту ничего не делает, мы просто передаём значения в репозиторий и он уже по этим данниым ищеть и по GET запросу выдаёт пользоватлемя, опять же данные передаются через интерфейсы
	if err != nil {                              // короче таким образом мы смотрим вернул ли репозиторий ошибку "пользователь не найден", если вернул то обрабатываем по if дальше (это такое более простое объяснение тому, что я написал ниже)
		if errors.Is(err, storage.ErrUserNotFound) { // у нас в storage есть несколько ошибок, которые может возвращать репозиторий, если репозиторий возвращает ошибку ErrUserNotFound, то наша функция errors.Is буквально говорит: "если ошибка err равна storage.ErrUserNotFound, то верни true, иначе false" про тру и фолз, это уже аспекты функции(если углубляться, просто эта функция возвращает булевые значения), ну и там дальше если true, то блок if выполняется
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials) // !!!!!!!ВОТ ЭТО ТОЖЕ ОЧЕНЬ ВАЖНО ПОНЯТЬ КОГДА БУДЕТ НЕ ПОНЯТНА ОБРАБОТАКА ОШИБОК В ХЕНДЛЕРЕ Если учетные данные неверны, функция возвращает ошибку ErrInvalidCredentials. Эта ошибка затем будет передана обратно в вызывающий код, то есть в HANDLER, а собирается и вызывается это, то есть как програмам понимает что handler вызывающий код, она понимает так как это собирается в main.go
		}

		a.log.Error("failed to get user")

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil { // программа понимает что user - это структура, которая совсем в другом фавйлу, таким образом, что когда мы передавали данные в репозиторий через интерфейс, этот интерфейс должен возвращать эту структуру, и когда мы эту опреацию(передача данных в репозиторий через интерфейс) суём с переменную user, то ей автоматически присваивается тип возвращаемый интерфейсом? Верно! Когда функция, реализующая интерфейс UserProvider, возвращает данные, она обязана создать и заполнить объект типа models.User. Именно этот объект потом присваивается переменной user.
		a.log.Info("invalid credentionals", sl.Err(err)) // а так эта функция сравнивает пароли, первый параметр, это тот, что мы сохранили в бд, а второй параметр, это пароль, который пришёл при входе в систему пользователем

		return "", fmt.Errorf("%s: %w", err, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, int(appID)) // то есть мы передаём id приложения и по этим данным бд ищет информацию о приложении, правильно? как и с интерфесом выше схема. только други цели, тут мы делаем это для того, что бы в будущем получить информацию о приложении, а выше делали что бы получить информацию о пользователе
	if err != nil {                                // АААААААААААААААААААААААААААААААААААА получается мы идём в appProvider пполучаем по id приложения(appID) информацию о приложении, и суём ответ(возвращаемое значение интферфейса) в переменную app | вот так вот всё просто работает
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil

}

// RegisterNewUser registers new user in the system and teturns user ID.
// If user with given username already exists, returns error.
func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	const op = "auth.RegisterNewServer"

	log := a.log.With(slog.String("op", op), slog.String("email", email)) // так же сюда можно впихнуть email, но в проде такую информацию(email или password) нужно уметь удалять при запросе, так как это может увидеть тот, кому не следует это видеть(злоумышленники)

	log.Info("registering new user")

	// хешируем пароль("солим")
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate pssword hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash) // Repo := &repository.Repository{db: db} (это пример(из названий всё понятно))вот с помощью такого обозначения предположительно в main.go слой бд понимает что в сервисном слое в структуре есть интерфейс UserSaver, который мы должны реализовать и можем доставать из него переданные параметры(собственно из за которых я пишу эту пояснительную записку, так как мне самому не понятно как слой работы с бд, находит данные. которые мы передали в этой функции)
	if err != nil {                                       // authService := &auth.Auth{userSaver: Repo} // вот так узнаёт что существует такая то структура
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save user", sl.Err(err)) // а так без замудрений для более продвинутых, мы просто передём все нужные значения в интерфейс, а в слое с бд мы за счёт реализации интерфейса достаём от туда данные, которые нужно записать в бд, а собственно говоря выше я объяснял как слой с бд понимает что он может реализовывать интерфейс UserSaver и доставать из него данные, проще говоря так бд слой просто напросто может понять какой метод он должен реализовать, а там всю работу с передаваемымми значениями уже сделал сервисный слой и слою бд просто остаётся сохранить данные в бд
		// если хотим передать (условно) все интерфейсы из слоя с бд то передаём структуру слоя бд(репозитория) в функцию конструктор New: authService := auth.New(Repo), а если хотим передать какой то отдельный интерфейс, то указываем структуру Auth: authService := auth.Auth{userSaver: Repo} к примеру если в репозитории один метод SaveUser только
		return 0, fmt.Errorf("%s: %w", op, err) // ЭТО ОЧЕНЬ СУПЕР МЕГА ВАЖНАЯ БАЗОВАЯ ВЕЩЬ, КОТОРУЮ Я ДОЛЖЕН КОНЕЧНО ЖЕ КАК SENIOR ЗАПОМНИТЬ!!!!!!!! это не сарказм (и к слову объяснения того, что я написал выше, это работа уже в main.go идёт, по связи слоев(метода луковицы) по сути)
	}

	log.Info("user registered")

	return id, nil // возвращаем id пользователя
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(slog.Int64("userID", userID), slog.String("op", op)) // опять же скажу что в продакшене, такую информацию лучше не логировать, но так как это учебный проект, то допускается

	log.Info("chrecking if user is admin by id")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) { // репозиторий будет возвращать ошибку ErrAppNotFound, а сервис сверяет(с помошью if errors.Is) если err является этой ошибкой, то выполняется соответствующий блок кода  В ДРУГИХ ОСНОВНЫХ ФУНКЦИЯХ ТАК ЖЕ
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}

		log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))
	}

	return isAdmin, nil

}
