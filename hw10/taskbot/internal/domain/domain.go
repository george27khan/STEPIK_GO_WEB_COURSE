package domain

type Task struct {
	ID             uint64
	Text           string
	Author         string
	AuthorChatID   int64
	Executor       string
	ExecutorChatID int64
}
