package dao

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"mbook/webook/internal/domain"
	"strconv"
	"time"
)

type ArticleS3DAO struct {
	ArticleGORMDAO
	oss *s3.S3
}

func NewArticleS3DAO(db *gorm.DB, oss *s3.S3) *ArticleS3DAO {
	return &ArticleS3DAO{
		ArticleGORMDAO: ArticleGORMDAO{db: db},
		oss:            oss,
	}
}
func (a *ArticleS3DAO) Sync(ctx context.Context, art Article) (int64, error) {
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
			pubArt := PublishedArticleV2{
				Id:       art.Id,
				Title:    art.Title,
				AuthorId: art.AuthorId,
				Ctime:    now,
				Utime:    now,
				Status:   art.Status,
			}
			pubArt.Ctime = now
			pubArt.Utime = now
			err = tx.Clauses(clause.OnConflict{
				// 对MySQL不起效，但是可以兼容别的方言
				// INSERT xxx ON DUPLICATE KEY SET `title`=?
				// 别的方言：
				// sqlite INSERT XXX ON CONFLICT DO UPDATES WHERE
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"title":  pubArt.Title,
					"utime":  pubArt.Utime,
					"status": pubArt.Status,
				}),
			}).Create(&pubArt).Error
			return err
		})
	if err != nil {
		return 0, err
	}
	_, err = a.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("webook-1314583317"),
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return id, err
}
func (a *ArticleS3DAO) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	now := time.Now().UnixMilli()
	err := a.db.WithContext(ctx).
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
			return tx.Model(&PublishedArticleV2{}).
				Where("id = ?", id).
				Updates(map[string]any{
					"utime":  now,
					"status": status,
				}).Error
		})
	if err != nil {
		return err
	}
	if status == domain.ArticleStatusPrivate {
		_, err = a.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string]("webook-1314583317"),
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})
	}
	return err
}

// 去掉Content
type PublishedArticleV2 struct {
	Id       int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}
