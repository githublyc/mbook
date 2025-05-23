package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
}

type ArticleGORMDAO struct {
	db *gorm.DB
}

func (a *ArticleGORMDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var res PublishedArticle
	err := a.db.WithContext(ctx).
		Where("id = ?", id).First(&res).Error
	return res, err
}

func (a *ArticleGORMDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := a.db.WithContext(ctx).
		Where("id=?", id).First(&art).Error
	return art, err
}

func (a *ArticleGORMDAO) GetByAuthor(ctx context.Context, uid int64,
	offset int, limit int) ([]Article, error) {
	var arts []Article
	err := a.db.WithContext(ctx).Where("author_id = ?", uid).
		Offset(offset).Limit(limit).
		Order("utime DESC").Find(&arts).Error
	return arts, err
}

func (a *ArticleGORMDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	return a.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			res := tx.Model(&Article{}).
				Where("id = ? and author_id = ?", id, uid).
				Updates(map[string]any{
					"utime":  now,
					"status": status,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return errors.New("失败。ID不对或者作者不对")
			}
			return tx.Model(&PublishedArticle{}).
				Where("id = ?", id).
				Updates(map[string]any{
					"utime":  now,
					"status": status,
				}).Error
		})
}

func (a *ArticleGORMDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := a.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			var (
				err error
			)
			dao := NewArticleGORMDAO(tx)
			if id > 0 {
				err = dao.UpdateById(ctx, art)
			} else {
				id, err = dao.Insert(ctx, art)
			}
			if err != nil {
				return err
			}
			art.Id = id
			now := time.Now().UnixMilli()
			pubArt := PublishedArticle(art)
			pubArt.Ctime = now
			pubArt.Utime = now
			//UpSert
			err = tx.Clauses(clause.OnConflict{
				// 对MySQL不起效，但是可以兼容别的方言
				// INSERT xxx ON DUPLICATE KEY SET `title`=?
				// 别的方言：
				// sqlite INSERT XXX ON CONFLICT DO UPDATES WHERE
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"title":   pubArt.Title,
					"content": pubArt.Content,
					"utime":   pubArt.Utime,
					"status":  pubArt.Status,
				}),
			}).Create(&pubArt).Error
			return err
		})
	return id, err
}

func (a *ArticleGORMDAO) SyncV1(ctx context.Context, art Article) (int64, error) {
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止后面业务panic
	defer tx.Rollback()

	var (
		id  = art.Id
		err error
	)
	dao := NewArticleGORMDAO(tx)
	if id > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		id, err = dao.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	now := time.Now().UnixMilli()
	pubArt := PublishedArticle(art)
	pubArt.Ctime = now
	pubArt.Utime = now
	err = tx.Clauses(clause.OnConflict{
		// 对MySQL不起效，但是可以兼容别的方言
		// INSERT xxx ON DUPLICATE KEY SET `title`=?
		// 别的方言：
		// sqlite INSERT XXX ON CONFLICT DO UPDATES WHERE
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   pubArt.Title,
			"content": pubArt.Content,
			"utime":   pubArt.Utime,
			"status":  pubArt.Status,
		}),
	}).Create(&pubArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, err
}

func (a *ArticleGORMDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&Article{}).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(
			map[string]any{
				"title":   art.Title,
				"content": art.Content,
				"status":  art.Status,
				"Utime":   now,
			})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("更新失败。ID不对或者作者不对")
	}
	return nil
}

func NewArticleGORMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{db: db}
}

func (a *ArticleGORMDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

type Article struct {
	Id       int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content  string `gorm:"type=BLOB" bson:"content,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}
type PublishedArticle Article
