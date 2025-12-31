package service

import (
	"context"
	"errors"
	"testing"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDepartmentRepository 部门仓储的Mock实现
type MockDepartmentRepository struct {
	mock.Mock
}

func (m *MockDepartmentRepository) Create(ctx context.Context, dept domain.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *MockDepartmentRepository) FindById(ctx context.Context, deptId int64) (domain.Department, error) {
	args := m.Called(ctx, deptId)
	return args.Get(0).(domain.Department), args.Error(1)
}

func (m *MockDepartmentRepository) FindAll(ctx context.Context) ([]domain.Department, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Department), args.Error(1)
}

func (m *MockDepartmentRepository) FindByParentId(ctx context.Context, parentId int64) ([]domain.Department, error) {
	args := m.Called(ctx, parentId)
	return args.Get(0).([]domain.Department), args.Error(1)
}

func (m *MockDepartmentRepository) Update(ctx context.Context, dept domain.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *MockDepartmentRepository) Delete(ctx context.Context, deptId int64) error {
	args := m.Called(ctx, deptId)
	return args.Error(0)
}

// TestDepartmentService_CreateDepartment 测试创建部门
func TestDepartmentService_CreateDepartment(t *testing.T) {
	tests := []struct {
		name       string
		dept       domain.Department
		mockSetup  func(*MockDepartmentRepository)
		wantErr    bool
		errMessage string
	}{
		{
			name: "成功创建部门",
			dept: domain.Department{
				DeptId:      1001,
				Name:        "技术部",
				ParentId:    0,
				Description: "技术研发部门",
			},
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(dept domain.Department) bool {
					return dept.DeptId == 1001 && dept.Name == "技术部"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "创建部门失败-数据库错误",
			dept: domain.Department{
				DeptId:      1002,
				Name:        "市场部",
				ParentId:    0,
				Description: "市场营销部门",
			},
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("database error"))
			},
			wantErr:    true,
			errMessage: "database error",
		},
		{
			name: "创建子部门",
			dept: domain.Department{
				DeptId:      2001,
				Name:        "前端组",
				ParentId:    1001,
				Description: "前端开发小组",
			},
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(dept domain.Department) bool {
					return dept.ParentId == 1001
				})).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDepartmentRepository)
			tt.mockSetup(mockRepo)

			svc := NewDepartmentService(mockRepo)
			err := svc.CreateDepartment(context.Background(), tt.dept)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestDepartmentService_GetDepartment 测试获取单个部门
func TestDepartmentService_GetDepartment(t *testing.T) {
	tests := []struct {
		name       string
		deptId     int64
		mockSetup  func(*MockDepartmentRepository)
		want       domain.Department
		wantErr    bool
		errMessage string
	}{
		{
			name:   "成功获取部门",
			deptId: 1001,
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("FindById", mock.Anything, int64(1001)).
					Return(domain.Department{
						DeptId:      1001,
						Name:        "技术部",
						ParentId:    0,
						Description: "技术研发部门",
					}, nil)
			},
			want: domain.Department{
				DeptId:      1001,
				Name:        "技术部",
				ParentId:    0,
				Description: "技术研发部门",
			},
			wantErr: false,
		},
		{
			name:   "部门不存在",
			deptId: 9999,
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("FindById", mock.Anything, int64(9999)).
					Return(domain.Department{}, errors.New("record not found"))
			},
			want:       domain.Department{},
			wantErr:    true,
			errMessage: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDepartmentRepository)
			tt.mockSetup(mockRepo)

			svc := NewDepartmentService(mockRepo)
			got, err := svc.GetDepartment(context.Background(), tt.deptId)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.DeptId, got.DeptId)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.ParentId, got.ParentId)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestDepartmentService_GetAllDepartments 测试获取所有部门
func TestDepartmentService_GetAllDepartments(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*MockDepartmentRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name: "成功获取所有部门",
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("FindAll", mock.Anything).
					Return([]domain.Department{
						{DeptId: 1001, Name: "技术部", ParentId: 0},
						{DeptId: 1002, Name: "市场部", ParentId: 0},
						{DeptId: 2001, Name: "前端组", ParentId: 1001},
					}, nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "空部门列表",
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("FindAll", mock.Anything).
					Return([]domain.Department{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "数据库查询失败",
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("FindAll", mock.Anything).
					Return([]domain.Department(nil), errors.New("database error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDepartmentRepository)
			tt.mockSetup(mockRepo)

			svc := NewDepartmentService(mockRepo)
			got, err := svc.GetAllDepartments(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestDepartmentService_GetSubDepartments 测试获取子部门
func TestDepartmentService_GetSubDepartments(t *testing.T) {
	tests := []struct {
		name      string
		parentId  int64
		mockSetup func(*MockDepartmentRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "成功获取子部门",
			parentId: 1001,
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("FindByParentId", mock.Anything, int64(1001)).
					Return([]domain.Department{
						{DeptId: 2001, Name: "前端组", ParentId: 1001},
						{DeptId: 2002, Name: "后端组", ParentId: 1001},
					}, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "无子部门",
			parentId: 2001,
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("FindByParentId", mock.Anything, int64(2001)).
					Return([]domain.Department{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:     "父部门ID为0-获取顶级部门",
			parentId: 0,
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("FindByParentId", mock.Anything, int64(0)).
					Return([]domain.Department{
						{DeptId: 1001, Name: "技术部", ParentId: 0},
						{DeptId: 1002, Name: "市场部", ParentId: 0},
					}, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDepartmentRepository)
			tt.mockSetup(mockRepo)

			svc := NewDepartmentService(mockRepo)
			got, err := svc.GetSubDepartments(context.Background(), tt.parentId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantCount)
				// 验证所有子部门的ParentId正确
				for _, dept := range got {
					assert.Equal(t, tt.parentId, dept.ParentId)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestDepartmentService_UpdateDepartment 测试更新部门
func TestDepartmentService_UpdateDepartment(t *testing.T) {
	tests := []struct {
		name      string
		dept      domain.Department
		mockSetup func(*MockDepartmentRepository)
		wantErr   bool
	}{
		{
			name: "成功更新部门信息",
			dept: domain.Department{
				DeptId:      1001,
				Name:        "技术研发部",
				ParentId:    0,
				Description: "更新后的描述",
			},
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(dept domain.Department) bool {
					return dept.DeptId == 1001 && dept.Name == "技术研发部"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "更新部门父节点",
			dept: domain.Department{
				DeptId:      2001,
				Name:        "前端组",
				ParentId:    1002, // 调整到其他部门下
				Description: "调整部门归属",
			},
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(dept domain.Department) bool {
					return dept.DeptId == 2001 && dept.ParentId == 1002
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "更新失败-部门不存在",
			dept: domain.Department{
				DeptId:   9999,
				Name:     "不存在的部门",
				ParentId: 0,
			},
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("Update", mock.Anything, mock.Anything).
					Return(errors.New("record not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDepartmentRepository)
			tt.mockSetup(mockRepo)

			svc := NewDepartmentService(mockRepo)
			err := svc.UpdateDepartment(context.Background(), tt.dept)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestDepartmentService_DeleteDepartment 测试删除部门
func TestDepartmentService_DeleteDepartment(t *testing.T) {
	tests := []struct {
		name      string
		deptId    int64
		mockSetup func(*MockDepartmentRepository)
		wantErr   bool
	}{
		{
			name:   "成功删除部门",
			deptId: 2001,
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("Delete", mock.Anything, int64(2001)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "删除失败-部门不存在",
			deptId: 9999,
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("Delete", mock.Anything, int64(9999)).
					Return(errors.New("record not found"))
			},
			wantErr: true,
		},
		{
			name:   "删除失败-存在子部门",
			deptId: 1001,
			mockSetup: func(m *MockDepartmentRepository) {
				m.On("Delete", mock.Anything, int64(1001)).
					Return(errors.New("cannot delete department with sub-departments"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDepartmentRepository)
			tt.mockSetup(mockRepo)

			svc := NewDepartmentService(mockRepo)
			err := svc.DeleteDepartment(context.Background(), tt.deptId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestDepartmentService_EdgeCases 测试边界情况
func TestDepartmentService_EdgeCases(t *testing.T) {
	t.Run("部门名称为空", func(t *testing.T) {
		mockRepo := new(MockDepartmentRepository)
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(dept domain.Department) bool {
			return dept.Name == ""
		})).Return(nil)

		svc := NewDepartmentService(mockRepo)
		err := svc.CreateDepartment(context.Background(), domain.Department{
			DeptId:   3001,
			Name:     "",
			ParentId: 0,
		})

		assert.NoError(t, err) // Service层不做校验，由Web层或数据库约束处理
		mockRepo.AssertExpectations(t)
	})

	t.Run("部门ID为负数", func(t *testing.T) {
		mockRepo := new(MockDepartmentRepository)
		mockRepo.On("FindById", mock.Anything, int64(-1)).
			Return(domain.Department{}, errors.New("invalid dept_id"))

		svc := NewDepartmentService(mockRepo)
		_, err := svc.GetDepartment(context.Background(), -1)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("循环引用-部门ParentId指向自己", func(t *testing.T) {
		mockRepo := new(MockDepartmentRepository)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(dept domain.Department) bool {
			return dept.DeptId == dept.ParentId
		})).Return(errors.New("circular reference detected"))

		svc := NewDepartmentService(mockRepo)
		err := svc.UpdateDepartment(context.Background(), domain.Department{
			DeptId:   1001,
			Name:     "技术部",
			ParentId: 1001, // 指向自己
		})

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
