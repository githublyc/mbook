package service

import (
	"context"
	"fmt"
	"math/rand"
	"mbook/mbook/internal/repository"
	"mbook/mbook/internal/service/sms"
)

type CodeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func (svc *CodeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generate()
	err := svc.repo.Set(ctx, biz, phone, code)
	//要开始发送验证码了
	if err != nil {
		return err
	}
	const codeTplId = "1877556"
	return svc.sms.Send(ctx, codeTplId, []string{code}, phone)

}
func (svc *CodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVerifyTooMany {
		//对外面屏蔽了验证次数过多的错误
		return false, nil
	}
	return ok, err
}
func (svc *CodeService) generate() string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
