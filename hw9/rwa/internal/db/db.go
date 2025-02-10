package db

import ls "rwa/internal/db/local_storage"

type DBStorage struct {
	User    *ls.UserStorage
	Article *ls.ArticleStorage
	Session *ls.SessionStorage
}

func NewDBStorage() *DBStorage {
	user := ls.NewUserStorage()
	article := ls.NewArticleStorage()
	session := ls.NewSessionStorage()
	return &DBStorage{user, article, session} //, article, session

}
