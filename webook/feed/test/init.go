package test

import (
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	feedv1 "mbook/webook/api/proto/gen/feed/v1"
	followMocks "mbook/webook/api/proto/gen/follow/v1/mocks"
	"mbook/webook/feed/grpc"
	"mbook/webook/feed/ioc"
	"mbook/webook/feed/repository"
	"mbook/webook/feed/repository/cache"
	"mbook/webook/feed/repository/dao"
	"mbook/webook/feed/service"
	"testing"
)

func InitGrpcServer(t *testing.T) (feedv1.FeedSvcServer, *followMocks.MockFollowServiceClient, *gorm.DB) {
	loggerV1 := ioc.InitLogger()
	db := ioc.InitDB(loggerV1)
	feedPullEventDAO := dao.NewFeedPullEventDAO(db)
	feedPushEventDAO := dao.NewFeedPushEventDAO(db)
	cmdable := ioc.InitRedis()
	feedEventCache := cache.NewFeedEventCache(cmdable)
	feedEventRepo := repository.NewFeedEventRepo(feedPullEventDAO, feedPushEventDAO, feedEventCache)
	mockCtrl := gomock.NewController(t)
	followClient := followMocks.NewMockFollowServiceClient(mockCtrl)
	v := ioc.RegisterHandler(feedEventRepo, followClient)
	feedService := service.NewFeedService(feedEventRepo, v)
	feedEventGrpcSvc := grpc.NewFeedEventGrpcSvc(feedService)
	return feedEventGrpcSvc, followClient, db
}
