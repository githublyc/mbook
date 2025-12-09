package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"mbook/webook/internal/domain"
	"mbook/webook/internal/repository/cache"
	"mbook/webook/internal/repository/dao"
	"time"
)

// ArticleRepository 接口层面上完全没有体现要不要缓存，如何缓存

//go:generate mockgen -source=./article.go -package=repomocks -destination=./mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}
type CachedArticleRepository struct {
	dao   dao.ArticleDAO
	cache cache.ArticleCache
	// 因为如果你直接访问 UserDAO，你就绕开了 repository，
	// 而repository 一般都有一些缓存机制
	userRepo UserRepository
	//SyncV1用
	readerDAO dao.ArticleReaderDAO
	authorDAO dao.ArticleAuthorDAO
	//SyncV2用
	db *gorm.DB
}

func (c *CachedArticleRepository) Cache() cache.ArticleCache {
	return c.cache
}
func (c *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	arts, err := c.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.PublishedArticle, domain.Article](arts,
		func(idx int, src dao.PublishedArticle) domain.Article {
			return c.ToDomain(dao.Article(src))
		}), nil
}

func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.ToDomain(dao.Article(art))
	author, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
		//要额外记录日志，因为你吞掉了错误信息
		//return res, nil
	}
	res.Author.Name = author.Nickname
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	go func() {
		er := c.cache.Set(ctx, c.ToDomain(art))
		if er != nil {
			//记录日志
		}
	}()
	return c.ToDomain(art), nil
}

func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 首先第一步，判定要不要查询缓存
	// 事实上， limit <= 100 都可以查询缓存
	if offset == 0 && limit == 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res, err
		} else {
			// 要考虑记录日志
			// 缓存未命中，你是可以忽略的
		}
	}
	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts,
		func(idx int, src dao.Article) domain.Article {
			return c.ToDomain(src)
		})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			// 缓存回写失败，不一定是大问题，但有可能是大问题
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 记录日志
				// 我需要监控这里
			}
		}
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, res)
	}()
	return res, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context,
	uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	if err == nil {
		er := c.cache.DelFirstPage(ctx, uid)
		if er != nil {
			// 也要记录日志
		}
	}
	return err
}

func NewCachedArticleRepositoryV2(authorDAO dao.ArticleAuthorDAO,
	readerDAO dao.ArticleReaderDAO) *CachedArticleRepository {
	return &CachedArticleRepository{
		authorDAO: authorDAO,
		readerDAO: readerDAO,
	}
}

// Sync DAO 层同步数据
func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	//在这里尝试缓存：一发表就缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		//可以灵活设置过期时间
		user, er := c.userRepo.FindById(ctx, art.Author.Id)
		if er != nil {
			//记录日志
			return
		}
		art.Author.Name = user.Nickname
		er = c.cache.SetPub(ctx, art)
		if er != nil {
			//记录日志
		}
	}()
	return id, err
}

func (c *CachedArticleRepository) Update(ctx context.Context,
	art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	return err
}

func NewCachedArticleRepository(dao dao.ArticleDAO,
	cache cache.ArticleCache, userRepo UserRepository) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	return id, err
}

// SyncV1 SyncV2 Repository 层同步数据
// SyncV1 非事务实现
func (c *CachedArticleRepository) SyncV1(ctx context.Context,
	art domain.Article) (int64, error) {
	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = c.authorDAO.Update(ctx, artn)
	} else {
		id, err = c.authorDAO.Create(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = c.readerDAO.Upsert(ctx, artn)
	return id, err
}

// SyncV2 事务实现
// 这种写法，逼迫 Repository 引入了 gorm.DB 的依赖，
// 也就是相当于 Repository 强依赖于 DAO 依赖的东西。
// • 没有坚持面向接口原则。
// • 跨层依赖，导致后续如果切换 gorm.DB 到别的实现，Repository 也会受到影响。
func (c *CachedArticleRepository) SyncV2(ctx context.Context,
	art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止后面业务panic
	defer tx.Rollback()

	authorDAO := dao.NewArticleGORMAuthorDAO(tx)
	readerDAO := dao.NewArticleGORMReaderDAO(tx)
	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = authorDAO.Update(ctx, artn)
	} else {
		id, err = authorDAO.Create(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = readerDAO.UpsertV2(ctx, dao.PublishedArticle(artn))
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, err
}
func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}
func (c *CachedArticleRepository) ToDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			//记录缓存
		}
	}
}
