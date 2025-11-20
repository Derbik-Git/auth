package auth // это по сути хендлер | сначала всё идёт в прото, и от прото этот хендлер берёт данные

import (
	"context"
	"errors"

	ssov1 "GRPC/protos/gen/go/tuzov.sso.v1" // импортируем наш прото файл, который мы создали, и говорим что мы импортируем его как ssov1, потому что это имя, которое мы указали в прото файле, и это имя, которое мы указали в файле main.go, и это имя будет использоваться в коде, потому что мы его указали в файле main.go

	auth1 "GRPC/sso/internal/services/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	Register(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

// файл прото описывает какие методы есть в нашем сервисе
// Например, в меню(proto) может быть указано, что есть блюдо "Login", которое принимает заказ (запрос) от клиента и возвращает ему еду (ответ).
// Когда вы создаете serverAPI, вы говорите, что этот шеф-повар будет работать в вашем ресторане и готовить заказы.

// ОЧЕНЬ ВАЖНО в кратце структура, которая обрабатывает, за счёт того что мы указали внутренную структуру из прото файла, а он уже хранит нереализованные методы, поэтому он и обработчик, потому что там методы, условно мы просто связываем метод из прото файла и в этом коде в единое целое и выходит что это одно и тоже, а тут, в этом коде эта структура нужна для того, что бы мы с помощью функции RegistrAuthServer, зарегестрировали обработчик, а с помощью самой функции Registr связали его с gRPC, потому что принимаемый параметр это тип данных *grpc.Server
type ServerAPI struct { // Структура serverAPI: Это ваша реализация сервера аутентификации. В Go вы можете создавать структуры, которые представляют собой объекты с определенными полями и методами. В данном случае serverAPI будет реализовывать методы, связанные с аутентификацией.
	ssov1.UnimplementedAuthServer // UnimplementedAuthServer из пакета gRPC — это интерфейс, определяющий обязательные методы аутентификации и авторизации. // Встраивание ssov1.UnimplementedAuthServer: Это специальная структура, сгенерированная из вашего protobuf файла. Она содержит "пустые" реализации методов интерфейса AuthServer. Встраивание этой структуры позволяет вам избежать ошибок компиляции, если вы не реализуете все методы, которые должны быть в интерфейсе AuthServer. В будущем, если вы добавите новые методы в ваш protobuf файл, вам не нужно будет обновлять вашу реализацию serverAPI, чтобы соответствовать новым требованиям.
	Auth                          Auth
} // большинство о чём тут говорится, то есть интерфейсы, нереализованые методы, это всё из файла прото, там так и будет, структура UnimplementedAuthServer, интерфейс AuthServer, и его нереализованные методы, которые мы писали ещё в ручную в файле прото (не сгенерированном), это самый первый файл, который обычно пишут когда начинают делать gRPC

// ОЧЕНЬ ВАЖНО ssov1.UnimplementedAuthServer служит как заглушка, что бы программа не требоввала сначала реализовать все методы интерфейса из прото файла, указывая эту структуру нам не обязательно реализовывать все методы, указанные в прото файле, если же мы уберём эту строку, то программа сразу выдасть ошибку что для запуска программы нужно реализовать все методы, вот для чего это делается оказывается, всё было проще чем я думал, но так же всё что я до этого писал никуда не отменяется, потому что и для тех целей это тоже нужно

func Register(gRPC *grpc.Server, auth Auth) { // Функция Register: Эта функция принимает указатель на экземпляр gRPC-сервера и регистрирует ваш сервер аутентификации (serverAPI) в этом сервере.
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{Auth: auth}) // РЕГИСТРИРУЕМ СЕРВИС(БИЗНЕС ЛОГИКУ) //  ssov1.RegisterAuthServer: Это функция, автоматически сгенерированная из вашего protobuf файла. Она принимает два аргумента: Первый аргумент — это экземпляр gRPC-сервера, который будет обрабатывать запросы. Второй аргумент — это указатель на вашу структуру serverAPI, которая реализует интерфейс аутентификации. Таким образом, gRPC знает, какие методы вы хотите использовать для обработки запросов аутентификации.
} // делаем обязательно эту функцию что бы прото файлы регестрировали наш grpc

// функция реигстр - обязательна функция что бы зарегестрировать/связать с gRPC сервером наш сервис и его методы которые находятся в файле прото, а понимает он что эти методы из прото файла таким образом что в функцию RegisterAuthServer мы указываем указатель на структуру serverAPI, которая хранит в себе структуру из файла прото со всеми нашими пустыми методами

const (
	emptyValue = 0
)

func (s *ServerAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) { // для обращения к запросу использовать req
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is empty") // status.Error Этот метод позволяет создавать ошибки, которые могут быть интерпретированы клиентом gRPC.
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is empty") // codes.InvalidArgument позволяет клиенту понять, что ошибка возникла из за неправильного пароля или email(к примеру)
	}

	// ВОТ ЭТО СУПЕР МЕГА ВАЖНО!!!!!!! корчоче мы создаём интерфейс, который реализует методы структуры

	if req.GetAppId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "app_id is empty")
	}

	// TODO: implement login via auth service

	token, err := s.Auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId())) // получаем токен из нашего сервиса аутентификации, который мы передали в функцию Register
	if err != nil {
		if errors.Is(err, auth1.ErrInvalidCredentials) { // ДАЮ ВАЖНОЕ ПОЯСНЕНИЕ, о том что, метод из хендлера вызывает метод из сервиса(бизнес логики), и в случае ошибки со стороны сервиса он возвращает ошибку, нами придуманного типа, и тут в хедлере он проверяет ошибку err вызова функции на тип ошибки созданный в сервисе и если она совпадает мы возвращаем сответственное пояснение return nil, status.Error(codes.InvalidArgument, "invalid credentials")
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error") // codes.Internal используется для обозначения внутренних ошибок, которые не связаны с клиентом или сервером. (клиенту лучше не показывать, если будет выкатываться в продакшн, ЛУЧШЕ ЗАМЕНИТЬ)
	}

	return &ssov1.LoginResponse{ //Геттеры (или методы доступа) — это специальные методы в объектно-ориентированном программировании, которые используются для получения значений полей (свойств) объекта. В контексте вашего кода, геттеры генерируются автоматически при компиляции файла .proto, который описывает ваши gRPC-сервисы и сообщения.
		Token: token,
	}, nil
}

func (s *ServerAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.ReqisterResopnse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}
	// без структуры s *ServerAPi(который реализует интерфейс Auth) мы не сможем вызвать метод Register из интерфейса Auth, методы которого реализуются в бизнес логике, а геттеры это передаваемые значения достанные из прото файла и переданные в бизнес логику
	userId, err := s.Auth.Register(ctx, req.GetEmail(), req.GetPassword()) // // здесь выполняется вызов бизнес-логики // это очень важно глубоко понять и разобраться в работе с интерфейсами и как за счёт них мы дёргаем интерфейсы
	if err != nil {
		if errors.Is(err, auth1.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.ReqisterResopnse{
		UserId: userId,
	}, nil

}

func (s *ServerAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdmineResponse, error) {
	if err := validateIsAdminRequest(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.Auth.IsAdmin(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, auth1.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdmineResponse{
		IsAdmin: isAdmin,
	}, nil
}

// выношу валидацию (проверку запроса перед использованием), в отдельную функцию для каждого метода (единственное не сделал для login потому что хочу попробовать и так и так, для наглядного примера как можно делать валидацию, но в отдельной функции выглядит покрасивее и в самом методе не занимает лишнего места)

func validateRegisterRequest(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is empty")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}

	return nil

}

func validateIsAdminRequest(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user_id is empty")
	}

	return nil

}
