package time

import "time"

type service struct{}

func NewService() *service {
	return &service{}
}

func (*service) Now() time.Time {
	return time.Now()
}
