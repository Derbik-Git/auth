package tests // прост очто бы не заблы, попробуй сочетанеи клавишь ctrl + i

import (
	"GRPC/sso/tests/suite"
	"testing"
	"time"

	ssov1 "github.com/Derbik-Git/protos"
	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID = 0             // идентификатор не поддерживаемого приложения(тест должен вернуть ошибку)
	appID      = 1             // идентификатор поддерживаемого приложения(должен возврашать правильный ответ)
	appSecret  = "test-secret" // серетный код приложения, который определяет доверенное приложение

	passDefaultLen = 10 // максимальная длинна пароля по умолчанию(если пароль слишком короткий, система сообщит об этом)
)

/*
1. Создание фейк-данных для пользователя (почта и пароль).
2. Регистрация нового пользователя с использованием этих данных.
3. Авторизация пользователя и получение токена.
4. Парсинг полученного токена и проверка его подписи.
5. Извлечение данных из токена.
6. Привежение к типу claims, для проверки всех полученных значений/переменных.
7. Проверка всех полученных значений/переменных.
8.
*/
func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := gofakeit.Password(true, true, true, true, false, passDefaultLen)

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{ // регистрируем и создаём структуру фейкового пользователя, который выполняте запрос
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)               //  require — это пакет внутри testify, который предоставляет функции, которые останавливают выполнение теста, если условие не выполнено.  NoError проверяет, что переданная ошибка (err) равна nil. Если ошибка не равна nil, тест будет завершён с ошибкой, и будет выведено сообщение об ошибке.
	assert.NotEmpty(t, respReq.GetUserId) // - GetUserId — это автоматически созданный методом компилятором Protobuf для удобной работы с полем user_id. // assert: Это пакет, который предоставляет функции для утверждений (assertions) в тестах. Утверждения — это проверки, которые вы делаете в тестах, чтобы убедиться, что ваш код работает так, как ожидается.

	respLogin := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    emmail,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now() // время авторизации, это для проверки время действия токена

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	// это функция парсит токен, которыйм мы указали, а потом она принимаемой, анонимной функцией запрашивает таким образом у разработчика секрет, для того что бы библиотка поняла, каким именно ключом был зашифрован этот токен
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) { // Чтобы проверить подпись токена, библиотека должна знать, каким именно ключом он был зашифрован. Поскольку ключ заранее неизвестен самой библиотеке, она запрашивает его у разработчика, вызывая эту функцию
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims) // проверяем распаршеный токен на тип jwt.MapClaims, который подрузомевает наличие мапы с payload токена, payload - полезные данные в токене, например когда истекает время действия токена, или имя пользователя, которому был рисвоен токен
	require.True(t, ok)

	// все полученные значения мы проверяем на соответствие ожидаемым
	assert.Equal(t, respReq.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int64(claims["app_id"].(float64)))

	const deltaSeconds = 1

	// В отличие от простого сравнения чисел с помощью assert.Equal, метод InDelta учитывает небольшую разницу между двумя величинами. Это полезно, когда абсолютное равенство сложно обеспечить (например, из-за временных задержек сети или особенностей измерения времени).
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds) // первым делом мы ко времени с момента создания токена добавляем время жизни токена из конфига - так мы можем знать момент времени, когда токен должен истечь, благодаря .Unix() мы полученное время превращаем в секунды(так удобнее скалдывать время, смотреть и считать), сравниваем его со значением в токене(мы указывали "exp" при создании токена), deltaSeconds - это допустимое отклонение в секундах(в нашем случае 5 секунд), то есть если то что в токене не совпадает с отмеренным нами временем в пределах 5-ти секунд, то это ничего страшного, это делается для того что бы тест не ложился просто так, а просто так это потому чт опогрешнасть будет в любом случае, так как отмеряемое нами время при получении токена отличается от настоящего времени создания токена, так как loginTime - это все го литшь импровизация этого времени для теста, ибо мы не знаем когда именно токен был создан

} // к слову при запуске тестов мы указываем путь к тестовым миграциям, то есть уазываем путь не к основным миграциям при запуске, а тестовым и таблица будет отдельная и называться migrations_test

// обратный тест, пред идущему, в данном случаем мы хотим положить приложенеи, а именно мы попробуем 2 раза зарегестрироваться, что не должно допускаться в нашем сервисе
func TestRegisterLogin_DuplicationRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := gofakeit.Password(true, true, true, true, false, passDefaultLen)

	respReg, err := st.AuthClient.Register(&ssov1.RegisterRequest{ // забыл кстати сказать, особо не вдумываемся, просто переменаая называется ответ регистрации, хоть мы и создаём по сути фейковый запрос, а не ответ, но вообщем это всё так потмоу что этот самый запрос автоматически возвращает ответ, поэтому мы и называем переменную имемнем ответа
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId)

	respReg, err := st.AuthClient.Register(&ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId)
	assert.ErrorContains(t, err, "user already exists") // Эта функция проверяет, содержится ли определённая строка в тексте ошибки, полученной в результате теста. Если такая строка найдена, проверка считается успешной. Если же строка отсутствует, значит, произошла непредвиденная ошибка, и тест провалится.

}

// завтра нужно реализовать табличные тесты, это не долго и очень легко как я понял, после этого толком проект закончен, останется посмотреть как менять свой код длдя внедрения других сервисов, с гитхаба скачаю проект url shortener, и внедрю его по видео николая тузова, пока так же по преженему учусь писать пет проекты, потом буду пытаться внедрят какие то технологии, типо кафки, редиса, графаны и так далее, только переде этим нужно посмотреть простенькие ознакомительные видео по этим технологиям, а дальше уже после понимания, что это такое, уже на практике внедрять их в свой проект если это требуется или внедрить 
// КСТАТИ НУЖНО УЗНАТЬ У ГПТ ЕСТЬ ЛИ НА ГИТХАБЕ ПРОЕКТЫ ГДЕ МОЖНО СКАЧАТЬ ИХ И ПОТРЕНИРОВАТЬСЯ ПОВНЕДРЯТЬ ВСЯКИЕ РЕДИСЫ, КАФКИ, ГРАФАНЫ ВОТ ЭТИ, Я ПОКА НЕ ЗНАЮ ДАЖЕ ЧТО ЭТО ТАКОЕ, НО ПАНТЕЛА СКАЗАЛ ЗНАТЬ НУЖНО

func TestRegister_FailCases(t *testing.T) {
	ctx, st, := suite.New(t)

	tests := []struct {
		name string
		email string
		password string
		expectedErr string // ожидаемое значение об ошибке

	}{
		{
			name: "Register with Empty Password",
			email: gofakeit.Email(),
			passwordL: "",
			expectedErr: "password is required",
		},
		{
			name: "Register with Empty Email",
			email: "",
			password: gofakeit.Password(true, true, true, true, false, passDefaultLen),
			expectedErr: "email is required",
		},
		{
			name: "Register with Both Empty Email and Password",
			email: "",
			password: "",
			expectedErr: "email is required",
		},
	}

/* 
1. Проходим по списку тестов.
2. Выбираем текущий тест (tt).
3. Пробуем зарегистрировать пользователя с указанными параметрами.
4. Проверяем, что возникла ошибка.
5. Проверяем, что ошибка содержит требуемый текст.
*/
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // Каждое обращение к t.Run() запускает новое автономное выполнение теста, записывая результат отдельно.
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email: tt.email,
				Passwoed: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Errror(), tt.expectedErr) // сравненеи ожидаемого значения об ошибке |   err.Error() — это строка ошибки, возвращаемая объектом ошибки (err). То есть, метод .Error() преобразует объект ошибки в строку, пригодную для сравнения.
		})
	}
}

func TestLogin_FailCases(t, *tesing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name string
		email string
		password string
		appID int64
		expectedErr string
	}{
		{
			name: "Login with Empty Password",
			email: gofakeit.Email(),
			password: "",
			appID: appID,
			expectedErr: "password is required",
		},
		{
			name: "Login with Empty Email",
			email: "",
			password: gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appID: appID,
			expectedErr: "email is required",
		},
		{
			name: "Login with both Empty Email and Password",
			email: "",
			password: "",
			appID: appID,
			expectedErr: "email is required",
		},
		{
			name: "Login with Non-Matching Password",
			email: gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appID: appID,
			expectedErr: "invalid email or password",
		},
		{
			name: "Login without AppID",
			email: gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appID: emptyAppID,
			expectedErr: "app_id is required"
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{ // ОТВЕЧАЮ на вопрос зачем нам регистрировать пользователя если мы проверяем логин, для того что бы проверить метод логина пользователя, нам нужно по факту зарегестрировать аккаунт, без зарегестриованного аккаунта пользователь не смог бы войти, вот и тут мы иммтируем наличие аккаунта
				Email: gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, false, passDefaultLen)
			})
			require.NoError(t, err) // применяется, когда мы проверяем наличие ошибки, с ожиданием о том, что её не произойдёт
		
			_, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email: tt.email,
				Password: tt.password,
				AppID: tt.appID,
			})
			require.Error(t, err) // применяем когда мы хотим поймать ошибку, от этого функция и называется FailCases
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
			
}
	// следующим делом скачать url-shortener, привязать его по ролику тузова, изучить сам проект, позже смотрим про бд как подключать у николая тузова и вот что я говорил про практиковаться с кафкой редисом и так далее