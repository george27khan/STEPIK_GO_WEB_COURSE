package local_storage

//
//import (
//	"sync"
//)
//
//type session struct {
//}
//
//type SessionStorage struct {
//	Session   map[string]session
//	SessionID int
//	Mu        *sync.RWMutex
//}
//
//// NewDBStorage создание хранилища данных
//func NewSessionStorage() *SessionStorage {
//	session := make(map[string]session)
//	return &SessionStorage{
//		session,
//		0,
//		&sync.RWMutex{},
//	}
//}

//// CreateUser создание пользователя в хранилище
//func (ls *MapStorage) Create(ctx context.Context, user rep.User) (string, error) {
//	ls.MuUser.Lock()
//	defer ls.MuUser.Unlock()
//	//ключем для поиска и создания будет почта
//	if _, ok := ls.Users[user.Email]; ok {
//		return "", fmt.Errorf("Пользователь с email %s уже существует", user.Email)
//	}
//	user.ID = strconv.Itoa(ls.UserID)
//	ls.UserID++
//	ls.Users[user.Email] = user
//	return user.ID, nil
//}
//
//func (ls *MapStorage) Get() {
//
//}
