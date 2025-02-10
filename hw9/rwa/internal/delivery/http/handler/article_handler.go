package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	dm "rwa/internal/domain"
	"rwa/internal/dto"
)

var (
// errCreateUser    error = errors.New("Create user error. %s")
// errLoginUser     error = errors.New("Login user error. %s")
// errRequestBody   error = errors.New("Unmarshal body error. %s")
// errHTTPMethod    error = errors.New("Bad HTTP method. %s")
// errCreateSession error = errors.New("Create session error. %s")
// errGetUser       error = errors.New("Get user error. %s")
// errMarshalResp   error = errors.New("Marshal response error. %s")
// errUserUpd       error = errors.New("Update user error. %s")
)

// интерфейс для обращения в usecase
type ArticleUseCase interface {
	Create(ctx context.Context, userID string, article *dm.Article) (*dm.Article, error)
	GetAll(ctx context.Context) ([]*dm.Article, error)
	GetByAuthor(ctx context.Context, username string) ([]*dm.Article, error)
	GetByTag(ctx context.Context, tag string) ([]*dm.Article, error)
}

type ArticleHandler struct {
	articleCase ArticleUseCase
}

func NewArticleHandler(articleCase ArticleUseCase) *ArticleHandler {
	return &ArticleHandler{articleCase}
}

// postArticle создание статьи
func (a *ArticleHandler) postArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := ctx.Value("session").(*dm.Session)

	// получаем тело запроса
	article := &dto.Article{&dto.ArticleInfo{}}
	if err := json.NewDecoder(r.Body).Decode(article); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf(errRequestBody.Error(), err.Error()))
		return
	}
	//маппинг структуры
	articleDM := &dm.Article{
		Body:        article.Info.Body,
		Description: article.Info.Description,
		TagList:     article.Info.TagList,
		Title:       article.Info.Title,
	}

	// создание статьи
	articleDM, err := a.articleCase.Create(ctx, session.UserID, articleDM)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
		return
	}
	articleResp := dto.ArticleResp{dto.ArticleRespInfo{
		Author: dto.Author{
			Bio:      articleDM.Author.BIO,
			Username: articleDM.Author.Username,
		},
		Body:        articleDM.Body,
		Title:       articleDM.Title,
		Description: articleDM.Description,
		CreatedAt:   articleDM.CreatedAt,
		UpdatedAt:   articleDM.UpdatedAt,
		TagList:     articleDM.TagList,
	}}
	resp, err := json.Marshal(articleResp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
	return
}

// postArticle создание статьи
func (a *ArticleHandler) getAllArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// создание статьи
	articlesDM, err := a.articleCase.GetAll(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
		return
	}
	articles := make([]*dto.ArticleRespInfo, 0, len(articlesDM))
	for _, article := range articlesDM {
		articles = append(articles, &dto.ArticleRespInfo{
			Author: dto.Author{
				Bio:      article.Author.BIO,
				Username: article.Author.Username,
			},
			Body:        article.Body,
			Title:       article.Title,
			Description: article.Description,
			CreatedAt:   article.CreatedAt,
			UpdatedAt:   article.UpdatedAt,
			TagList:     article.TagList,
		})
	}
	articlesResp := dto.ArticleAllResp{
		Articles:      articles,
		ArticlesCount: len(articles),
	}
	resp, err := json.Marshal(articlesResp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	return
}

// getArticleByAuthor получение статей по автору
func (a *ArticleHandler) getArticleByAuthor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// создание статьи
	username := r.URL.Query().Get("author")
	articlesDM, err := a.articleCase.GetByAuthor(ctx, username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
		return
	}
	articles := make([]*dto.ArticleRespInfo, 0, len(articlesDM))
	for _, article := range articlesDM {
		articles = append(articles, &dto.ArticleRespInfo{
			Author: dto.Author{
				Bio:      article.Author.BIO,
				Username: article.Author.Username,
			},
			Body:        article.Body,
			Title:       article.Title,
			Description: article.Description,
			CreatedAt:   article.CreatedAt,
			UpdatedAt:   article.UpdatedAt,
			TagList:     article.TagList,
		})
	}
	articlesResp := dto.ArticleAllResp{
		Articles:      articles,
		ArticlesCount: len(articles),
	}
	resp, err := json.Marshal(articlesResp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	return
}

// getArticleByTag получение статей по тэгу
func (a *ArticleHandler) getArticleByTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// создание статьи
	tag := r.URL.Query().Get("tag")
	articlesDM, err := a.articleCase.GetByTag(ctx, tag)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
		return
	}
	articles := make([]*dto.ArticleRespInfo, 0, len(articlesDM))
	for _, article := range articlesDM {
		articles = append(articles, &dto.ArticleRespInfo{
			Author: dto.Author{
				Bio:      article.Author.BIO,
				Username: article.Author.Username,
			},
			Body:        article.Body,
			Title:       article.Title,
			Description: article.Description,
			CreatedAt:   article.CreatedAt,
			UpdatedAt:   article.UpdatedAt,
			TagList:     article.TagList,
		})
	}
	articlesResp := dto.ArticleAllResp{
		Articles:      articles,
		ArticlesCount: len(articles),
	}
	resp, err := json.Marshal(articlesResp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	return
}

func (a *ArticleHandler) Articles(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("r.Method", r.Method)
	//fmt.Println("r.URL", r.URL)
	//fmt.Println("r.URL.Path", r.URL.Path)
	//fmt.Println("r.URL.Query()", r.URL.Query())
	if r.Method == http.MethodGet && len(r.URL.Query()) == 0 {
		a.getAllArticle(w, r)
		return
	} else if r.Method == http.MethodGet && r.URL.Query().Get("author") != "" {
		a.getArticleByAuthor(w, r)
		return
	} else if r.Method == http.MethodGet && r.URL.Query().Get("tag") != "" {
		a.getArticleByTag(w, r)
		return
	} else if r.Method == http.MethodPost {
		a.postArticle(w, r)
		return
	}
	writeError(w, http.StatusBadRequest, errHTTPMethod)
}
