package kratosx

import "context"

// Service 实现 service
type Service struct {
	UnimplementedUserServiceServer // 继承，实现service的时候【必须强制组合mented】，也是为了向后兼容
	Name                           string
}

func (s *Service) GetById(ctx context.Context, request *GetByIdRequest) (*GetByIdResponse, error) {
	//ctx, span := otel.Tracer("server_biz123").Start(ctx, "get_by_id123")
	//defer span.End()
	//ddl, ok := ctx.Deadline() // Deadline是一个时间点，表示某个时间点deadline，过了这个时间点，context就被取消
	//if ok {
	//	res := ddl.Sub(time.Now()) // 计算时间差
	//	log.Println(res.String())
	//}
	//time.Sleep(time.Millisecond * 100)
	return &GetByIdResponse{
		User: &User{
			Id:   123,
			Name: "hgg+" + s.Name,
		},
	}, nil
}

func (s *Service) GetById1(ctx context.Context, request *GetByIdRequest) (*GetByIdResponse, error) {
	//ctx, span := otel.Tracer("server_biz456").Start(ctx, "get_by_id456")
	//defer span.End()
	//ddl, ok := ctx.Deadline() // Deadline是一个时间点，表示某个时间点deadline，过了这个时间点，context就被取消
	//if ok {
	//	res := ddl.Sub(time.Now()) // 计算时间差
	//	log.Println(res.String())
	//}
	//time.Sleep(time.Millisecond * 100)
	return &GetByIdResponse{
		User: &User{
			Id:   456,
			Name: "hgg+" + s.Name,
		},
	}, nil
}
