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

// MockUserDb 用户DAO Mock
type MockUserDb struct {
	mock.Mock
}

func (m *MockUserDb) Create(ctx context.Context, user dao.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserDb) FindById(ctx context.Context, userId int64) (dao.User, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).(dao.User), args.Error(1)
}

func (m *MockUserDb) FindByUsername(ctx context.Context, username string) (dao.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(dao.User), args.Error(1)
}

func (m *MockUserDb) FindByDeptId(ctx context.Context, deptId int64) ([]dao.User, error) {
	args := m.Called(ctx, deptId)
	return args.Get(0).([]dao.User), args.Error(1)
}

func (m *MockUserDb) FindAll(ctx context.Context) ([]dao.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]dao.User), args.Error(1)
}

func (m *MockUserDb) Update(ctx context.Context, user dao.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserDb) Delete(ctx context.Context, userId int64) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}

// MockUserRoleDb 用户角色关联DAO Mock
type MockUserRoleDb struct {
	mock.Mock
}

func (m *MockUserRoleDb) Create(ctx context.Context, ur dao.UserRole) error {
	args := m.Called(ctx, ur)
	return args.Error(0)
}

func (m *MockUserRoleDb) Delete(ctx context.Context, userId, roleId int64) error {
	args := m.Called(ctx, userId, roleId)
	return args.Error(0)
}

func (m *MockUserRoleDb) FindRolesByUserId(ctx context.Context, userId int64) ([]dao.Role, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]dao.Role), args.Error(1)
}

func (m *MockUserRoleDb) FindUsersByRoleId(ctx context.Context, roleId int64) ([]dao.User, error) {
	args := m.Called(ctx, roleId)
	return args.Get(0).([]dao.User), args.Error(1)
}

// MockPermissionDb 权限DAO Mock
type MockPermissionDb struct {
	mock.Mock
}

func (m *MockPermissionDb) Create(ctx context.Context, perm dao.Permission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockPermissionDb) FindById(ctx context.Context, permId int64) (dao.Permission, error) {
	args := m.Called(ctx, permId)
	return args.Get(0).(dao.Permission), args.Error(1)
}

func (m *MockPermissionDb) FindByCode(ctx context.Context, code string) (dao.Permission, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(dao.Permission), args.Error(1)
}

func (m *MockPermissionDb) FindAll(ctx context.Context) ([]dao.Permission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]dao.Permission), args.Error(1)
}

func (m *MockPermissionDb) Update(ctx context.Context, perm dao.Permission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockPermissionDb) Delete(ctx context.Context, permId int64) error {
	args := m.Called(ctx, permId)
	return args.Error(0)
}

func (m *MockPermissionDb) FindByRoleId(ctx context.Context, roleId int64) ([]dao.Permission, error) {
	args := m.Called(ctx, roleId)
	return args.Get(0).([]dao.Permission), args.Error(1)
}

// MockRoleDb 角色DAO Mock
type MockRoleDb struct {
	mock.Mock
}

func (m *MockRoleDb) Create(ctx context.Context, role dao.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleDb) FindById(ctx context.Context, roleId int64) (dao.Role, error) {
	args := m.Called(ctx, roleId)
	return args.Get(0).(dao.Role), args.Error(1)
}

func (m *MockRoleDb) FindByCode(ctx context.Context, code string) (dao.Role, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(dao.Role), args.Error(1)
}

func (m *MockRoleDb) FindAll(ctx context.Context) ([]dao.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]dao.Role), args.Error(1)
}

func (m *MockRoleDb) Update(ctx context.Context, role dao.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleDb) Delete(ctx context.Context, roleId int64) error {
	args := m.Called(ctx, roleId)
	return args.Error(0)
}

// MockRolePermissionDb 角色权限关联DAO Mock
type MockRolePermissionDb struct {
	mock.Mock
}

func (m *MockRolePermissionDb) Create(ctx context.Context, rp dao.RolePermission) error {
	args := m.Called(ctx, rp)
	return args.Error(0)
}

func (m *MockRolePermissionDb) Delete(ctx context.Context, roleId, permId int64) error {
	args := m.Called(ctx, roleId, permId)
	return args.Error(0)
}

func (m *MockRolePermissionDb) FindPermissionsByRoleId(ctx context.Context, roleId int64) ([]dao.Permission, error) {
	args := m.Called(ctx, roleId)
	return args.Get(0).([]dao.Permission), args.Error(1)
}

// MockCronPermissionDb 任务权限DAO Mock
type MockCronPermissionDb struct {
	mock.Mock
}

func (m *MockCronPermissionDb) Create(ctx context.Context, cp dao.CronPermission) error {
	args := m.Called(ctx, cp)
	return args.Error(0)
}

func (m *MockCronPermissionDb) Delete(ctx context.Context, cronId, deptId int64) error {
	args := m.Called(ctx, cronId, deptId)
	return args.Error(0)
}

func (m *MockCronPermissionDb) FindDeptsByCronId(ctx context.Context, cronId int64) ([]dao.Department, error) {
	args := m.Called(ctx, cronId)
	return args.Get(0).([]dao.Department), args.Error(1)
}

func (m *MockCronPermissionDb) CheckPermission(ctx context.Context, cronId, deptId int64) (bool, error) {
	args := m.Called(ctx, cronId, deptId)
	return args.Bool(0), args.Error(1)
}

// TestUserRepository_Create 测试创建用户
func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		user      domain.User
		mockSetup func(*MockUserDb)
		wantErr   bool
	}{
		{
			name: "成功创建用户-完整信息",
			user: domain.User{
				UserId:   1001,
				Username: "zhangsan",
				Password: "hashed_password",
				Email:    "zhangsan@example.com",
				Phone:    "13800138000",
				DeptId:   2001,
				Status:   "active",
			},
			mockSetup: func(m *MockUserDb) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(user dao.User) bool {
					return user.UserId == 1001 &&
						user.Username == "zhangsan" &&
						user.Email.Valid &&
						user.Phone.Valid &&
						user.Email.String == "zhangsan@example.com"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "创建用户-可选字段为空",
			user: domain.User{
				UserId:   1002,
				Username: "lisi",
				Password: "hashed_password",
				Email:    "",
				Phone:    "",
				DeptId:   2001,
				Status:   "active",
			},
			mockSetup: func(m *MockUserDb) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(user dao.User) bool {
					return user.UserId == 1002 && !user.Email.Valid && !user.Phone.Valid
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "创建失败-用户名重复",
			user: domain.User{
				UserId:   1003,
				Username: "existing_user",
				Password: "password",
				DeptId:   2001,
			},
			mockSetup: func(m *MockUserDb) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("duplicate username"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserDb := new(MockUserDb)
			mockUserRoleDb := new(MockUserRoleDb)
			mockPermDb := new(MockPermissionDb)
			tt.mockSetup(mockUserDb)

			repo := NewUserRepository(mockUserDb, mockUserRoleDb, mockPermDb)
			err := repo.Create(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockUserDb.AssertExpectations(t)
		})
	}
}

// TestUserRepository_FindUserPermissions 测试获取用户权限
func TestUserRepository_FindUserPermissions(t *testing.T) {
	tests := []struct {
		name      string
		userId    int64
		mockSetup func(*MockUserRoleDb, *MockPermissionDb)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "成功获取用户权限-单个角色",
			userId: 1001,
			mockSetup: func(mur *MockUserRoleDb, mp *MockPermissionDb) {
				// 用户有一个角色
				mur.On("FindRolesByUserId", mock.Anything, int64(1001)).
					Return([]dao.Role{
						{RoleId: 2001, Code: "admin"},
					}, nil)
				// 该角色有3个权限
				mp.On("FindByRoleId", mock.Anything, int64(2001)).
					Return([]dao.Permission{
						{PermId: 3001, Code: "cron:read"},
						{PermId: 3002, Code: "cron:create"},
						{PermId: 3003, Code: "dept:read"},
					}, nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:   "成功获取用户权限-多个角色有重复权限",
			userId: 1002,
			mockSetup: func(mur *MockUserRoleDb, mp *MockPermissionDb) {
				// 用户有两个角色
				mur.On("FindRolesByUserId", mock.Anything, int64(1002)).
					Return([]dao.Role{
						{RoleId: 2001, Code: "operator"},
						{RoleId: 2002, Code: "viewer"},
					}, nil)
				// 第一个角色的权限
				mp.On("FindByRoleId", mock.Anything, int64(2001)).
					Return([]dao.Permission{
						{PermId: 3001, Code: "cron:read"},
						{PermId: 3002, Code: "cron:create"},
					}, nil)
				// 第二个角色的权限（有重复）
				mp.On("FindByRoleId", mock.Anything, int64(2002)).
					Return([]dao.Permission{
						{PermId: 3001, Code: "cron:read"}, // 重复权限
						{PermId: 3004, Code: "dept:read"},
					}, nil)
			},
			wantCount: 3, // 去重后应该是3个
			wantErr:   false,
		},
		{
			name:   "用户无角色",
			userId: 1003,
			mockSetup: func(mur *MockUserRoleDb, mp *MockPermissionDb) {
				mur.On("FindRolesByUserId", mock.Anything, int64(1003)).
					Return([]dao.Role{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "用户有角色但角色无权限",
			userId: 1004,
			mockSetup: func(mur *MockUserRoleDb, mp *MockPermissionDb) {
				mur.On("FindRolesByUserId", mock.Anything, int64(1004)).
					Return([]dao.Role{
						{RoleId: 2003, Code: "empty_role"},
					}, nil)
				mp.On("FindByRoleId", mock.Anything, int64(2003)).
					Return([]dao.Permission{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserDb := new(MockUserDb)
			mockUserRoleDb := new(MockUserRoleDb)
			mockPermDb := new(MockPermissionDb)
			tt.mockSetup(mockUserRoleDb, mockPermDb)

			repo := NewUserRepository(mockUserDb, mockUserRoleDb, mockPermDb)
			perms, err := repo.FindUserPermissions(context.Background(), tt.userId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, perms, tt.wantCount)
			}

			mockUserRoleDb.AssertExpectations(t)
			mockPermDb.AssertExpectations(t)
		})
	}
}

// TestRoleRepository_AssignPermission 测试分配权限给角色
func TestRoleRepository_AssignPermission(t *testing.T) {
	tests := []struct {
		name      string
		roleId    int64
		permId    int64
		mockSetup func(*MockRolePermissionDb)
		wantErr   bool
	}{
		{
			name:   "成功分配权限",
			roleId: 2001,
			permId: 3001,
			mockSetup: func(m *MockRolePermissionDb) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(rp dao.RolePermission) bool {
					return rp.RoleId == 2001 && rp.PermId == 3001
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "分配失败-权限已存在",
			roleId: 2001,
			permId: 3001,
			mockSetup: func(m *MockRolePermissionDb) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("duplicate permission"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleDb := new(MockRoleDb)
			mockRolePermDb := new(MockRolePermissionDb)
			tt.mockSetup(mockRolePermDb)

			repo := NewRoleRepository(mockRoleDb, mockRolePermDb)
			err := repo.AssignPermission(context.Background(), tt.roleId, tt.permId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRolePermDb.AssertExpectations(t)
		})
	}
}

// TestRoleRepository_RemovePermission 测试移除角色权限
func TestRoleRepository_RemovePermission(t *testing.T) {
	tests := []struct {
		name      string
		roleId    int64
		permId    int64
		mockSetup func(*MockRolePermissionDb)
		wantErr   bool
	}{
		{
			name:   "成功移除权限",
			roleId: 2001,
			permId: 3001,
			mockSetup: func(m *MockRolePermissionDb) {
				m.On("Delete", mock.Anything, int64(2001), int64(3001)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "移除失败-权限不存在",
			roleId: 2001,
			permId: 9999,
			mockSetup: func(m *MockRolePermissionDb) {
				m.On("Delete", mock.Anything, int64(2001), int64(9999)).
					Return(errors.New("permission not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleDb := new(MockRoleDb)
			mockRolePermDb := new(MockRolePermissionDb)
			tt.mockSetup(mockRolePermDb)

			repo := NewRoleRepository(mockRoleDb, mockRolePermDb)
			err := repo.RemovePermission(context.Background(), tt.roleId, tt.permId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRolePermDb.AssertExpectations(t)
		})
	}
}

// TestAuthRepository_AssignRoleToUser 测试分配角色给用户
func TestAuthRepository_AssignRoleToUser(t *testing.T) {
	tests := []struct {
		name      string
		userId    int64
		roleId    int64
		mockSetup func(*MockUserRoleDb)
		wantErr   bool
	}{
		{
			name:   "成功分配角色",
			userId: 1001,
			roleId: 2001,
			mockSetup: func(m *MockUserRoleDb) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(ur dao.UserRole) bool {
					return ur.UserId == 1001 && ur.RoleId == 2001
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "分配失败-用户已有该角色",
			userId: 1001,
			roleId: 2001,
			mockSetup: func(m *MockUserRoleDb) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("role already assigned"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRoleDb := new(MockUserRoleDb)
			mockCronPermDb := new(MockCronPermissionDb)
			tt.mockSetup(mockUserRoleDb)

			repo := NewAuthRepository(mockUserRoleDb, mockCronPermDb)
			err := repo.AssignRoleToUser(context.Background(), tt.userId, tt.roleId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockUserRoleDb.AssertExpectations(t)
		})
	}
}

// TestAuthRepository_RemoveRoleFromUser 测试移除用户角色
func TestAuthRepository_RemoveRoleFromUser(t *testing.T) {
	tests := []struct {
		name      string
		userId    int64
		roleId    int64
		mockSetup func(*MockUserRoleDb)
		wantErr   bool
	}{
		{
			name:   "成功移除角色",
			userId: 1001,
			roleId: 2001,
			mockSetup: func(m *MockUserRoleDb) {
				m.On("Delete", mock.Anything, int64(1001), int64(2001)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "移除失败-用户无该角色",
			userId: 1001,
			roleId: 9999,
			mockSetup: func(m *MockUserRoleDb) {
				m.On("Delete", mock.Anything, int64(1001), int64(9999)).
					Return(errors.New("role not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRoleDb := new(MockUserRoleDb)
			mockCronPermDb := new(MockCronPermissionDb)
			tt.mockSetup(mockUserRoleDb)

			repo := NewAuthRepository(mockUserRoleDb, mockCronPermDb)
			err := repo.RemoveRoleFromUser(context.Background(), tt.userId, tt.roleId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockUserRoleDb.AssertExpectations(t)
		})
	}
}

// TestAuthRepository_CheckCronPermission 测试检查任务权限
func TestAuthRepository_CheckCronPermission(t *testing.T) {
	tests := []struct {
		name      string
		cronId    int64
		deptId    int64
		mockSetup func(*MockCronPermissionDb)
		want      bool
		wantErr   bool
	}{
		{
			name:   "部门有任务权限",
			cronId: 5001,
			deptId: 2001,
			mockSetup: func(m *MockCronPermissionDb) {
				m.On("CheckPermission", mock.Anything, int64(5001), int64(2001)).
					Return(true, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "部门无任务权限",
			cronId: 5001,
			deptId: 2002,
			mockSetup: func(m *MockCronPermissionDb) {
				m.On("CheckPermission", mock.Anything, int64(5001), int64(2002)).
					Return(false, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "查询失败",
			cronId: 5001,
			deptId: 2001,
			mockSetup: func(m *MockCronPermissionDb) {
				m.On("CheckPermission", mock.Anything, int64(5001), int64(2001)).
					Return(false, errors.New("database error"))
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRoleDb := new(MockUserRoleDb)
			mockCronPermDb := new(MockCronPermissionDb)
			tt.mockSetup(mockCronPermDb)

			repo := NewAuthRepository(mockUserRoleDb, mockCronPermDb)
			got, err := repo.CheckCronPermission(context.Background(), tt.cronId, tt.deptId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockCronPermDb.AssertExpectations(t)
		})
	}
}

// TestAuthRepository_GrantCronPermission 测试授予任务权限
func TestAuthRepository_GrantCronPermission(t *testing.T) {
	tests := []struct {
		name      string
		cronId    int64
		deptId    int64
		mockSetup func(*MockCronPermissionDb)
		wantErr   bool
	}{
		{
			name:   "成功授予权限",
			cronId: 5001,
			deptId: 2001,
			mockSetup: func(m *MockCronPermissionDb) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(cp dao.CronPermission) bool {
					return cp.CronId == 5001 && cp.DeptId == 2001
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "授予失败-权限已存在",
			cronId: 5001,
			deptId: 2001,
			mockSetup: func(m *MockCronPermissionDb) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("permission already granted"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRoleDb := new(MockUserRoleDb)
			mockCronPermDb := new(MockCronPermissionDb)
			tt.mockSetup(mockCronPermDb)

			repo := NewAuthRepository(mockUserRoleDb, mockCronPermDb)
			err := repo.GrantCronPermission(context.Background(), tt.cronId, tt.deptId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCronPermDb.AssertExpectations(t)
		})
	}
}

// TestPermissionRepository_CreateAndFind 测试权限的创建和查询
func TestPermissionRepository_CreateAndFind(t *testing.T) {
	t.Run("创建权限-完整信息", func(t *testing.T) {
		mockPermDb := new(MockPermissionDb)
		mockPermDb.On("Create", mock.Anything, mock.MatchedBy(func(perm dao.Permission) bool {
			return perm.PermId == 3001 &&
				perm.Code == "cron:create" &&
				perm.Resource == "cron" &&
				perm.Action == "create" &&
				perm.Description.Valid
		})).Return(nil)

		repo := NewPermissionRepository(mockPermDb)
		err := repo.Create(context.Background(), domain.Permission{
			PermId:      3001,
			Name:        "创建任务",
			Code:        "cron:create",
			Resource:    "cron",
			Action:      "create",
			Description: "创建定时任务权限",
		})

		assert.NoError(t, err)
		mockPermDb.AssertExpectations(t)
	})

	t.Run("根据Code查找权限", func(t *testing.T) {
		mockPermDb := new(MockPermissionDb)
		mockPermDb.On("FindByCode", mock.Anything, "cron:read").
			Return(dao.Permission{
				PermId:      3001,
				Code:        "cron:read",
				Resource:    "cron",
				Action:      "read",
				Description: sql.NullString{String: "查看任务", Valid: true},
			}, nil)

		repo := NewPermissionRepository(mockPermDb)
		perm, err := repo.FindByCode(context.Background(), "cron:read")

		assert.NoError(t, err)
		assert.Equal(t, int64(3001), perm.PermId)
		assert.Equal(t, "cron:read", perm.Code)
		assert.Equal(t, "cron", perm.Resource)
		assert.Equal(t, "read", perm.Action)
		mockPermDb.AssertExpectations(t)
	})
}

// TestEntityConversion 测试实体转换
func TestEntityConversion(t *testing.T) {
	t.Run("User DAO到Domain转换", func(t *testing.T) {
		daoUser := dao.User{
			ID:       1,
			UserId:   1001,
			Username: "zhangsan",
			Password: "hashed_password",
			Email:    sql.NullString{String: "zhangsan@example.com", Valid: true},
			Phone:    sql.NullString{String: "13800138000", Valid: true},
			DeptId:   2001,
			Status:   "active",
			Ctime:    1234567890.0,
			Utime:    1234567890.0,
		}

		domainUser := dao.ToUserDomain(daoUser)

		assert.Equal(t, daoUser.UserId, domainUser.UserId)
		assert.Equal(t, daoUser.Username, domainUser.Username)
		assert.Equal(t, daoUser.Email.String, domainUser.Email)
		assert.Equal(t, daoUser.Phone.String, domainUser.Phone)
		assert.Equal(t, daoUser.DeptId, domainUser.DeptId)
		assert.Equal(t, daoUser.Status, domainUser.Status)
	})

	t.Run("Role DAO到Domain转换", func(t *testing.T) {
		daoRole := dao.Role{
			ID:          1,
			RoleId:      2001,
			Name:        "管理员",
			Code:        "admin",
			Description: sql.NullString{String: "系统管理员", Valid: true},
		}

		domainRole := dao.ToRoleDomain(daoRole)

		assert.Equal(t, daoRole.RoleId, domainRole.RoleId)
		assert.Equal(t, daoRole.Name, domainRole.Name)
		assert.Equal(t, daoRole.Code, domainRole.Code)
		assert.Equal(t, daoRole.Description.String, domainRole.Description)
	})

	t.Run("Permission DAO到Domain转换", func(t *testing.T) {
		daoPerm := dao.Permission{
			ID:          1,
			PermId:      3001,
			Name:        "创建任务",
			Code:        "cron:create",
			Resource:    "cron",
			Action:      "create",
			Description: sql.NullString{String: "创建定时任务", Valid: true},
		}

		domainPerm := dao.ToPermissionDomain(daoPerm)

		assert.Equal(t, daoPerm.PermId, domainPerm.PermId)
		assert.Equal(t, daoPerm.Name, domainPerm.Name)
		assert.Equal(t, daoPerm.Code, domainPerm.Code)
		assert.Equal(t, daoPerm.Resource, domainPerm.Resource)
		assert.Equal(t, daoPerm.Action, domainPerm.Action)
		assert.Equal(t, daoPerm.Description.String, domainPerm.Description)
	})
}
