package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDepartmentDb DAO层Mock
type MockDepartmentDb struct {
	mock.Mock
}

func (m *MockDepartmentDb) Create(ctx context.Context, dept dao.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *MockDepartmentDb) FindById(ctx context.Context, deptId int64) (dao.Department, error) {
	args := m.Called(ctx, deptId)
	return args.Get(0).(dao.Department), args.Error(1)
}

func (m *MockDepartmentDb) FindAll(ctx context.Context) ([]dao.Department, error) {
	args := m.Called(ctx)
	return args.Get(0).([]dao.Department), args.Error(1)
}

func (m *MockDepartmentDb) FindByParentId(ctx context.Context, parentId int64) ([]dao.Department, error) {
	args := m.Called(ctx, parentId)
	return args.Get(0).([]dao.Department), args.Error(1)
}

func (m *MockDepartmentDb) Update(ctx context.Context, dept dao.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *MockDepartmentDb) Delete(ctx context.Context, deptId int64) error {
	args := m.Called(ctx, deptId)
	return args.Error(0)
}

// TestDepartmentRepository_Create 测试创建部门
func TestDepartmentRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		dept      domain.Department
		mockSetup func(*MockDepartmentDb)
		wantErr   bool
	}{
		{
			name: "成功创建部门",
			dept: domain.Department{
				DeptId:      1001,
				Name:        "技术部",
				ParentId:    0,
				Description: "技术研发部门",
			},
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(dept dao.Department) bool {
					return dept.DeptId == 1001 &&
						dept.Name == "技术部" &&
						dept.Description.Valid &&
						dept.Description.String == "技术研发部门"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "创建部门-空描述",
			dept: domain.Department{
				DeptId:      1002,
				Name:        "市场部",
				ParentId:    0,
				Description: "",
			},
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(dept dao.Department) bool {
					return dept.DeptId == 1002 && !dept.Description.Valid
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "创建子部门",
			dept: domain.Department{
				DeptId:      2001,
				Name:        "前端组",
				ParentId:    1001,
				Description: "前端开发小组",
			},
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(dept dao.Department) bool {
					return dept.ParentId == 1001
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "创建失败-数据库错误",
			dept: domain.Department{
				DeptId:   1003,
				Name:     "财务部",
				ParentId: 0,
			},
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("database connection error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDb := new(MockDepartmentDb)
			tt.mockSetup(mockDb)

			repo := NewDepartmentRepository(mockDb)
			err := repo.Create(context.Background(), tt.dept)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDb.AssertExpectations(t)
		})
	}
}

// TestDepartmentRepository_FindById 测试根据ID查找部门
func TestDepartmentRepository_FindById(t *testing.T) {
	tests := []struct {
		name      string
		deptId    int64
		mockSetup func(*MockDepartmentDb)
		want      domain.Department
		wantErr   bool
	}{
		{
			name:   "成功查找部门",
			deptId: 1001,
			mockSetup: func(m *MockDepartmentDb) {
				m.On("FindById", mock.Anything, int64(1001)).
					Return(dao.Department{
						ID:          1,
						DeptId:      1001,
						Name:        "技术部",
						ParentId:    0,
						Description: sql.NullString{String: "技术研发部门", Valid: true},
						Ctime:       1234567890.0,
						Utime:       1234567890.0,
					}, nil)
			},
			want: domain.Department{
				ID:          1,
				DeptId:      1001,
				Name:        "技术部",
				ParentId:    0,
				Description: "技术研发部门",
				Ctime:       1234567890.0,
				Utime:       1234567890.0,
			},
			wantErr: false,
		},
		{
			name:   "查找部门-描述为空",
			deptId: 1002,
			mockSetup: func(m *MockDepartmentDb) {
				m.On("FindById", mock.Anything, int64(1002)).
					Return(dao.Department{
						ID:          2,
						DeptId:      1002,
						Name:        "市场部",
						ParentId:    0,
						Description: sql.NullString{Valid: false},
						Ctime:       1234567890.0,
						Utime:       1234567890.0,
					}, nil)
			},
			want: domain.Department{
				ID:          2,
				DeptId:      1002,
				Name:        "市场部",
				ParentId:    0,
				Description: "",
				Ctime:       1234567890.0,
				Utime:       1234567890.0,
			},
			wantErr: false,
		},
		{
			name:   "部门不存在",
			deptId: 9999,
			mockSetup: func(m *MockDepartmentDb) {
				m.On("FindById", mock.Anything, int64(9999)).
					Return(dao.Department{}, errors.New("record not found"))
			},
			want:    domain.Department{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDb := new(MockDepartmentDb)
			tt.mockSetup(mockDb)

			repo := NewDepartmentRepository(mockDb)
			got, err := repo.FindById(context.Background(), tt.deptId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.DeptId, got.DeptId)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.ParentId, got.ParentId)
				assert.Equal(t, tt.want.Description, got.Description)
			}

			mockDb.AssertExpectations(t)
		})
	}
}

// TestDepartmentRepository_FindAll 测试查找所有部门
func TestDepartmentRepository_FindAll(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*MockDepartmentDb)
		wantCount int
		wantErr   bool
	}{
		{
			name: "成功查找所有部门",
			mockSetup: func(m *MockDepartmentDb) {
				m.On("FindAll", mock.Anything).
					Return([]dao.Department{
						{
							DeptId:      1001,
							Name:        "技术部",
							ParentId:    0,
							Description: sql.NullString{String: "技术部门", Valid: true},
						},
						{
							DeptId:      1002,
							Name:        "市场部",
							ParentId:    0,
							Description: sql.NullString{Valid: false},
						},
						{
							DeptId:   2001,
							Name:     "前端组",
							ParentId: 1001,
						},
					}, nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "空部门列表",
			mockSetup: func(m *MockDepartmentDb) {
				m.On("FindAll", mock.Anything).
					Return([]dao.Department{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "查询失败",
			mockSetup: func(m *MockDepartmentDb) {
				m.On("FindAll", mock.Anything).
					Return([]dao.Department(nil), errors.New("database error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDb := new(MockDepartmentDb)
			tt.mockSetup(mockDb)

			repo := NewDepartmentRepository(mockDb)
			got, err := repo.FindAll(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantCount)
			}

			mockDb.AssertExpectations(t)
		})
	}
}

// TestDepartmentRepository_FindByParentId 测试根据父ID查找子部门
func TestDepartmentRepository_FindByParentId(t *testing.T) {
	tests := []struct {
		name      string
		parentId  int64
		mockSetup func(*MockDepartmentDb)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "成功查找子部门",
			parentId: 1001,
			mockSetup: func(m *MockDepartmentDb) {
				m.On("FindByParentId", mock.Anything, int64(1001)).
					Return([]dao.Department{
						{DeptId: 2001, Name: "前端组", ParentId: 1001},
						{DeptId: 2002, Name: "后端组", ParentId: 1001},
						{DeptId: 2003, Name: "测试组", ParentId: 1001},
					}, nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:     "无子部门",
			parentId: 2001,
			mockSetup: func(m *MockDepartmentDb) {
				m.On("FindByParentId", mock.Anything, int64(2001)).
					Return([]dao.Department{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:     "查找顶级部门",
			parentId: 0,
			mockSetup: func(m *MockDepartmentDb) {
				m.On("FindByParentId", mock.Anything, int64(0)).
					Return([]dao.Department{
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
			mockDb := new(MockDepartmentDb)
			tt.mockSetup(mockDb)

			repo := NewDepartmentRepository(mockDb)
			got, err := repo.FindByParentId(context.Background(), tt.parentId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantCount)
				// 验证所有部门的ParentId正确
				for _, dept := range got {
					assert.Equal(t, tt.parentId, dept.ParentId)
				}
			}

			mockDb.AssertExpectations(t)
		})
	}
}

// TestDepartmentRepository_Update 测试更新部门
func TestDepartmentRepository_Update(t *testing.T) {
	tests := []struct {
		name      string
		dept      domain.Department
		mockSetup func(*MockDepartmentDb)
		wantErr   bool
	}{
		{
			name: "成功更新部门",
			dept: domain.Department{
				DeptId:      1001,
				Name:        "技术研发部",
				ParentId:    0,
				Description: "更新后的描述",
			},
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(dept dao.Department) bool {
					return dept.DeptId == 1001 &&
						dept.Name == "技术研发部" &&
						dept.Description.Valid &&
						dept.Description.String == "更新后的描述"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "更新部门-清空描述",
			dept: domain.Department{
				DeptId:      1001,
				Name:        "技术部",
				ParentId:    0,
				Description: "",
			},
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(dept dao.Department) bool {
					return dept.DeptId == 1001 && !dept.Description.Valid
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "更新部门父节点",
			dept: domain.Department{
				DeptId:   2001,
				Name:     "前端组",
				ParentId: 1002,
			},
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(dept dao.Department) bool {
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
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Update", mock.Anything, mock.Anything).
					Return(errors.New("record not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDb := new(MockDepartmentDb)
			tt.mockSetup(mockDb)

			repo := NewDepartmentRepository(mockDb)
			err := repo.Update(context.Background(), tt.dept)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDb.AssertExpectations(t)
		})
	}
}

// TestDepartmentRepository_Delete 测试删除部门
func TestDepartmentRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		deptId    int64
		mockSetup func(*MockDepartmentDb)
		wantErr   bool
	}{
		{
			name:   "成功删除部门",
			deptId: 2001,
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Delete", mock.Anything, int64(2001)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "删除失败-部门不存在",
			deptId: 9999,
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Delete", mock.Anything, int64(9999)).
					Return(errors.New("record not found"))
			},
			wantErr: true,
		},
		{
			name:   "删除失败-外键约束",
			deptId: 1001,
			mockSetup: func(m *MockDepartmentDb) {
				m.On("Delete", mock.Anything, int64(1001)).
					Return(errors.New("foreign key constraint"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDb := new(MockDepartmentDb)
			tt.mockSetup(mockDb)

			repo := NewDepartmentRepository(mockDb)
			err := repo.Delete(context.Background(), tt.deptId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDb.AssertExpectations(t)
		})
	}
}

// TestDepartmentRepository_DomainConversion 测试Domain和DAO实体转换
func TestDepartmentRepository_DomainConversion(t *testing.T) {
	t.Run("DAO到Domain转换-带描述", func(t *testing.T) {
		daoEntity := dao.Department{
			ID:          1,
			DeptId:      1001,
			Name:        "技术部",
			ParentId:    0,
			Description: sql.NullString{String: "技术研发部门", Valid: true},
			Ctime:       1234567890.0,
			Utime:       1234567890.0,
		}

		domainEntity := dao.ToDepartmentDomain(daoEntity)

		assert.Equal(t, daoEntity.ID, domainEntity.ID)
		assert.Equal(t, daoEntity.DeptId, domainEntity.DeptId)
		assert.Equal(t, daoEntity.Name, domainEntity.Name)
		assert.Equal(t, daoEntity.ParentId, domainEntity.ParentId)
		assert.Equal(t, daoEntity.Description.String, domainEntity.Description)
		assert.Equal(t, daoEntity.Ctime, domainEntity.Ctime)
		assert.Equal(t, daoEntity.Utime, domainEntity.Utime)
	})

	t.Run("DAO到Domain转换-空描述", func(t *testing.T) {
		daoEntity := dao.Department{
			ID:          2,
			DeptId:      1002,
			Name:        "市场部",
			ParentId:    0,
			Description: sql.NullString{Valid: false},
			Ctime:       1234567890.0,
			Utime:       1234567890.0,
		}

		domainEntity := dao.ToDepartmentDomain(daoEntity)

		assert.Equal(t, "", domainEntity.Description)
	})

	t.Run("Domain到DAO转换-带描述", func(t *testing.T) {
		mockDb := new(MockDepartmentDb)
		mockDb.On("Create", mock.Anything, mock.MatchedBy(func(dept dao.Department) bool {
			return dept.Description.Valid && dept.Description.String == "测试描述"
		})).Return(nil)

		repo := NewDepartmentRepository(mockDb)
		err := repo.Create(context.Background(), domain.Department{
			DeptId:      1003,
			Name:        "测试部",
			ParentId:    0,
			Description: "测试描述",
		})

		assert.NoError(t, err)
		mockDb.AssertExpectations(t)
	})

	t.Run("Domain到DAO转换-空描述", func(t *testing.T) {
		mockDb := new(MockDepartmentDb)
		mockDb.On("Create", mock.Anything, mock.MatchedBy(func(dept dao.Department) bool {
			return !dept.Description.Valid
		})).Return(nil)

		repo := NewDepartmentRepository(mockDb)
		err := repo.Create(context.Background(), domain.Department{
			DeptId:      1004,
			Name:        "测试部2",
			ParentId:    0,
			Description: "",
		})

		assert.NoError(t, err)
		mockDb.AssertExpectations(t)
	})
}
