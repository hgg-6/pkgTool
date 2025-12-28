package service

import (
	"context"
	"errors"
	"testing"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository 用户仓储Mock
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindById(ctx context.Context, userId int64) (domain.User, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByDeptId(ctx context.Context, deptId int64) ([]domain.User, error) {
	args := m.Called(ctx, deptId)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) FindAll(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, userId int64) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}

func (m *MockUserRepository) FindUserRoles(ctx context.Context, userId int64) ([]domain.Role, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]domain.Role), args.Error(1)
}

func (m *MockUserRepository) FindUserPermissions(ctx context.Context, userId int64) ([]domain.Permission, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]domain.Permission), args.Error(1)
}

// MockRoleRepository 角色仓储Mock
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) FindById(ctx context.Context, roleId int64) (domain.Role, error) {
	args := m.Called(ctx, roleId)
	return args.Get(0).(domain.Role), args.Error(1)
}

func (m *MockRoleRepository) FindByCode(ctx context.Context, code string) (domain.Role, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(domain.Role), args.Error(1)
}

func (m *MockRoleRepository) FindAll(ctx context.Context) ([]domain.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, roleId int64) error {
	args := m.Called(ctx, roleId)
	return args.Error(0)
}

func (m *MockRoleRepository) AssignPermission(ctx context.Context, roleId, permId int64) error {
	args := m.Called(ctx, roleId, permId)
	return args.Error(0)
}

func (m *MockRoleRepository) RemovePermission(ctx context.Context, roleId, permId int64) error {
	args := m.Called(ctx, roleId, permId)
	return args.Error(0)
}

func (m *MockRoleRepository) FindRolePermissions(ctx context.Context, roleId int64) ([]domain.Permission, error) {
	args := m.Called(ctx, roleId)
	return args.Get(0).([]domain.Permission), args.Error(1)
}

// MockPermissionRepository 权限仓储Mock
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(ctx context.Context, perm domain.Permission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockPermissionRepository) FindById(ctx context.Context, permId int64) (domain.Permission, error) {
	args := m.Called(ctx, permId)
	return args.Get(0).(domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) FindByCode(ctx context.Context, code string) (domain.Permission, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) FindAll(ctx context.Context) ([]domain.Permission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) Update(ctx context.Context, perm domain.Permission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockPermissionRepository) Delete(ctx context.Context, permId int64) error {
	args := m.Called(ctx, permId)
	return args.Error(0)
}

// MockAuthRepository 认证仓储Mock
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) AssignRoleToUser(ctx context.Context, userId, roleId int64) error {
	args := m.Called(ctx, userId, roleId)
	return args.Error(0)
}

func (m *MockAuthRepository) RemoveRoleFromUser(ctx context.Context, userId, roleId int64) error {
	args := m.Called(ctx, userId, roleId)
	return args.Error(0)
}

func (m *MockAuthRepository) GrantCronPermission(ctx context.Context, cronId, deptId int64) error {
	args := m.Called(ctx, cronId, deptId)
	return args.Error(0)
}

func (m *MockAuthRepository) RevokeCronPermission(ctx context.Context, cronId, deptId int64) error {
	args := m.Called(ctx, cronId, deptId)
	return args.Error(0)
}

func (m *MockAuthRepository) CheckCronPermission(ctx context.Context, cronId, deptId int64) (bool, error) {
	args := m.Called(ctx, cronId, deptId)
	return args.Bool(0), args.Error(1)
}

// TestUserService_CreateUser 测试创建用户
func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		user      domain.User
		mockSetup func(*MockUserRepository)
		wantErr   bool
	}{
		{
			name: "成功创建用户",
			user: domain.User{
				UserId:   1001,
				Username: "zhangsan",
				Password: "password123",
				Email:    "zhangsan@example.com",
				DeptId:   1001,
				Status:   "active",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					// 验证密码已被加密
					return user.Username == "zhangsan" && user.Password != "password123"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "创建用户失败-用户名重复",
			user: domain.User{
				UserId:   1002,
				Username: "existing_user",
				Password: "password123",
				DeptId:   1001,
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("duplicate username"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			svc := NewUserService(mockRepo)
			err := svc.CreateUser(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUserService_Login 测试用户登录
func TestUserService_Login(t *testing.T) {
	// 预先生成一个加密的密码
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		username  string
		password  string
		mockSetup func(*MockUserRepository)
		wantErr   bool
		checkErr  func(error) bool
	}{
		{
			name:     "登录成功",
			username: "zhangsan",
			password: "correct_password",
			mockSetup: func(m *MockUserRepository) {
				m.On("FindByUsername", mock.Anything, "zhangsan").
					Return(domain.User{
						UserId:   1001,
						Username: "zhangsan",
						Password: string(hashedPassword),
						DeptId:   1001,
						Status:   "active",
					}, nil)
			},
			wantErr: false,
		},
		{
			name:     "登录失败-用户不存在",
			username: "nonexistent",
			password: "password123",
			mockSetup: func(m *MockUserRepository) {
				m.On("FindByUsername", mock.Anything, "nonexistent").
					Return(domain.User{}, errors.New("user not found"))
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, ErrInvalidCredentials)
			},
		},
		{
			name:     "登录失败-密码错误",
			username: "zhangsan",
			password: "wrong_password",
			mockSetup: func(m *MockUserRepository) {
				m.On("FindByUsername", mock.Anything, "zhangsan").
					Return(domain.User{
						UserId:   1001,
						Username: "zhangsan",
						Password: string(hashedPassword),
						DeptId:   1001,
					}, nil)
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, ErrInvalidCredentials)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			svc := NewUserService(mockRepo)
			user, err := svc.Login(context.Background(), tt.username, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.username, user.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUserService_GetUserPermissions 测试获取用户权限
func TestUserService_GetUserPermissions(t *testing.T) {
	tests := []struct {
		name      string
		userId    int64
		mockSetup func(*MockUserRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "成功获取用户权限",
			userId: 1001,
			mockSetup: func(m *MockUserRepository) {
				m.On("FindUserPermissions", mock.Anything, int64(1001)).
					Return([]domain.Permission{
						{PermId: 1, Code: "cron:read", Name: "查看任务"},
						{PermId: 2, Code: "cron:create", Name: "创建任务"},
						{PermId: 3, Code: "dept:read", Name: "查看部门"},
					}, nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:   "用户无权限",
			userId: 1002,
			mockSetup: func(m *MockUserRepository) {
				m.On("FindUserPermissions", mock.Anything, int64(1002)).
					Return([]domain.Permission{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			svc := NewUserService(mockRepo)
			perms, err := svc.GetUserPermissions(context.Background(), tt.userId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, perms, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUserService_ChangePassword 测试修改密码
func TestUserService_ChangePassword(t *testing.T) {
	oldPassword := "old_password"
	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		userId      int64
		oldPassword string
		newPassword string
		mockSetup   func(*MockUserRepository)
		wantErr     bool
	}{
		{
			name:        "成功修改密码",
			userId:      1001,
			oldPassword: oldPassword,
			newPassword: "new_password",
			mockSetup: func(m *MockUserRepository) {
				m.On("FindById", mock.Anything, int64(1001)).
					Return(domain.User{
						UserId:   1001,
						Username: "zhangsan",
						Password: string(hashedOldPassword),
					}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					// 验证新密码已被加密且不同于旧密码
					return user.UserId == 1001 && user.Password != string(hashedOldPassword)
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "修改失败-旧密码错误",
			userId:      1001,
			oldPassword: "wrong_old_password",
			newPassword: "new_password",
			mockSetup: func(m *MockUserRepository) {
				m.On("FindById", mock.Anything, int64(1001)).
					Return(domain.User{
						UserId:   1001,
						Password: string(hashedOldPassword),
					}, nil)
			},
			wantErr: true,
		},
		{
			name:        "修改失败-用户不存在",
			userId:      9999,
			oldPassword: oldPassword,
			newPassword: "new_password",
			mockSetup: func(m *MockUserRepository) {
				m.On("FindById", mock.Anything, int64(9999)).
					Return(domain.User{}, errors.New("user not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			svc := NewUserService(mockRepo)
			err := svc.ChangePassword(context.Background(), tt.userId, tt.oldPassword, tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRoleService_CreateRole 测试创建角色
func TestRoleService_CreateRole(t *testing.T) {
	tests := []struct {
		name      string
		role      domain.Role
		mockSetup func(*MockRoleRepository)
		wantErr   bool
	}{
		{
			name: "成功创建角色",
			role: domain.Role{
				RoleId:      1001,
				Name:        "管理员",
				Code:        "admin",
				Description: "系统管理员角色",
			},
			mockSetup: func(m *MockRoleRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(role domain.Role) bool {
					return role.Code == "admin"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "创建失败-角色代码重复",
			role: domain.Role{
				RoleId: 1002,
				Name:   "操作员",
				Code:   "admin",
			},
			mockSetup: func(m *MockRoleRepository) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("duplicate role code"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRoleRepository)
			tt.mockSetup(mockRepo)

			svc := NewRoleService(mockRepo)
			err := svc.CreateRole(context.Background(), tt.role)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRoleService_AssignPermission 测试分配权限给角色
func TestRoleService_AssignPermission(t *testing.T) {
	tests := []struct {
		name      string
		roleId    int64
		permId    int64
		mockSetup func(*MockRoleRepository)
		wantErr   bool
	}{
		{
			name:   "成功分配权限",
			roleId: 1001,
			permId: 2001,
			mockSetup: func(m *MockRoleRepository) {
				m.On("AssignPermission", mock.Anything, int64(1001), int64(2001)).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "分配失败-角色不存在",
			roleId: 9999,
			permId: 2001,
			mockSetup: func(m *MockRoleRepository) {
				m.On("AssignPermission", mock.Anything, int64(9999), int64(2001)).
					Return(errors.New("role not found"))
			},
			wantErr: true,
		},
		{
			name:   "分配失败-权限已存在",
			roleId: 1001,
			permId: 2001,
			mockSetup: func(m *MockRoleRepository) {
				m.On("AssignPermission", mock.Anything, int64(1001), int64(2001)).
					Return(errors.New("permission already assigned"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRoleRepository)
			tt.mockSetup(mockRepo)

			svc := NewRoleService(mockRepo)
			err := svc.AssignPermission(context.Background(), tt.roleId, tt.permId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRoleService_GetRolePermissions 测试获取角色的所有权限
func TestRoleService_GetRolePermissions(t *testing.T) {
	tests := []struct {
		name      string
		roleId    int64
		mockSetup func(*MockRoleRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "成功获取角色权限",
			roleId: 1001,
			mockSetup: func(m *MockRoleRepository) {
				m.On("FindRolePermissions", mock.Anything, int64(1001)).
					Return([]domain.Permission{
						{PermId: 1, Code: "cron:read"},
						{PermId: 2, Code: "cron:create"},
						{PermId: 3, Code: "cron:update"},
						{PermId: 4, Code: "cron:delete"},
					}, nil)
			},
			wantCount: 4,
			wantErr:   false,
		},
		{
			name:   "角色无权限",
			roleId: 1002,
			mockSetup: func(m *MockRoleRepository) {
				m.On("FindRolePermissions", mock.Anything, int64(1002)).
					Return([]domain.Permission{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRoleRepository)
			tt.mockSetup(mockRepo)

			svc := NewRoleService(mockRepo)
			perms, err := svc.GetRolePermissions(context.Background(), tt.roleId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, perms, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestAuthService_CheckUserPermission 测试检查用户权限
func TestAuthService_CheckUserPermission(t *testing.T) {
	tests := []struct {
		name           string
		userId         int64
		permissionCode string
		mockSetup      func(*MockUserRepository)
		want           bool
		wantErr        bool
	}{
		{
			name:           "用户拥有该权限",
			userId:         1001,
			permissionCode: "cron:create",
			mockSetup: func(m *MockUserRepository) {
				m.On("FindUserPermissions", mock.Anything, int64(1001)).
					Return([]domain.Permission{
						{Code: "cron:read"},
						{Code: "cron:create"},
						{Code: "cron:update"},
					}, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:           "用户无该权限",
			userId:         1001,
			permissionCode: "cron:delete",
			mockSetup: func(m *MockUserRepository) {
				m.On("FindUserPermissions", mock.Anything, int64(1001)).
					Return([]domain.Permission{
						{Code: "cron:read"},
						{Code: "cron:create"},
					}, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:           "用户无任何权限",
			userId:         1002,
			permissionCode: "cron:read",
			mockSetup: func(m *MockUserRepository) {
				m.On("FindUserPermissions", mock.Anything, int64(1002)).
					Return([]domain.Permission{}, nil)
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := new(MockAuthRepository)
			mockUserRepo := new(MockUserRepository)
			tt.mockSetup(mockUserRepo)

			svc := NewAuthService(mockAuthRepo, mockUserRepo)
			got, err := svc.CheckUserPermission(context.Background(), tt.userId, tt.permissionCode)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

// TestAuthService_CheckCronPermission 测试检查任务权限
func TestAuthService_CheckCronPermission(t *testing.T) {
	tests := []struct {
		name      string
		userId    int64
		cronId    int64
		mockSetup func(*MockAuthRepository, *MockUserRepository)
		want      bool
		wantErr   bool
	}{
		{
			name:   "用户部门有任务权限",
			userId: 1001,
			cronId: 5001,
			mockSetup: func(ma *MockAuthRepository, mu *MockUserRepository) {
				mu.On("FindById", mock.Anything, int64(1001)).
					Return(domain.User{
						UserId: 1001,
						DeptId: 2001,
					}, nil)
				ma.On("CheckCronPermission", mock.Anything, int64(5001), int64(2001)).
					Return(true, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "用户部门无任务权限",
			userId: 1001,
			cronId: 5002,
			mockSetup: func(ma *MockAuthRepository, mu *MockUserRepository) {
				mu.On("FindById", mock.Anything, int64(1001)).
					Return(domain.User{
						UserId: 1001,
						DeptId: 2001,
					}, nil)
				ma.On("CheckCronPermission", mock.Anything, int64(5002), int64(2001)).
					Return(false, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "用户不存在",
			userId: 9999,
			cronId: 5001,
			mockSetup: func(ma *MockAuthRepository, mu *MockUserRepository) {
				mu.On("FindById", mock.Anything, int64(9999)).
					Return(domain.User{}, errors.New("user not found"))
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := new(MockAuthRepository)
			mockUserRepo := new(MockUserRepository)
			tt.mockSetup(mockAuthRepo, mockUserRepo)

			svc := NewAuthService(mockAuthRepo, mockUserRepo)
			got, err := svc.CheckCronPermission(context.Background(), tt.userId, tt.cronId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockUserRepo.AssertExpectations(t)
			mockAuthRepo.AssertExpectations(t)
		})
	}
}

// TestAuthService_AssignRoleToUser 测试分配角色给用户
func TestAuthService_AssignRoleToUser(t *testing.T) {
	tests := []struct {
		name      string
		userId    int64
		roleId    int64
		mockSetup func(*MockAuthRepository)
		wantErr   bool
	}{
		{
			name:   "成功分配角色",
			userId: 1001,
			roleId: 2001,
			mockSetup: func(m *MockAuthRepository) {
				m.On("AssignRoleToUser", mock.Anything, int64(1001), int64(2001)).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "分配失败-用户已有该角色",
			userId: 1001,
			roleId: 2001,
			mockSetup: func(m *MockAuthRepository) {
				m.On("AssignRoleToUser", mock.Anything, int64(1001), int64(2001)).
					Return(errors.New("role already assigned"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := new(MockAuthRepository)
			mockUserRepo := new(MockUserRepository)
			tt.mockSetup(mockAuthRepo)

			svc := NewAuthService(mockAuthRepo, mockUserRepo)
			err := svc.AssignRoleToUser(context.Background(), tt.userId, tt.roleId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockAuthRepo.AssertExpectations(t)
		})
	}
}

// TestPermissionService_CreatePermission 测试创建权限
func TestPermissionService_CreatePermission(t *testing.T) {
	tests := []struct {
		name      string
		perm      domain.Permission
		mockSetup func(*MockPermissionRepository)
		wantErr   bool
	}{
		{
			name: "成功创建权限",
			perm: domain.Permission{
				PermId:      3001,
				Name:        "创建任务",
				Code:        "cron:create",
				Resource:    "cron",
				Action:      "create",
				Description: "创建定时任务的权限",
			},
			mockSetup: func(m *MockPermissionRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(perm domain.Permission) bool {
					return perm.Code == "cron:create" && perm.Resource == "cron"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "创建失败-权限代码重复",
			perm: domain.Permission{
				PermId:   3002,
				Code:     "cron:create",
				Resource: "cron",
				Action:   "create",
			},
			mockSetup: func(m *MockPermissionRepository) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("duplicate permission code"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPermissionRepository)
			tt.mockSetup(mockRepo)

			svc := NewPermissionService(mockRepo)
			err := svc.CreatePermission(context.Background(), tt.perm)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPermissionService_GetAllPermissions 测试获取所有权限
func TestPermissionService_GetAllPermissions(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*MockPermissionRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name: "成功获取所有权限",
			mockSetup: func(m *MockPermissionRepository) {
				m.On("FindAll", mock.Anything).
					Return([]domain.Permission{
						{Code: "cron:create", Resource: "cron", Action: "create"},
						{Code: "cron:read", Resource: "cron", Action: "read"},
						{Code: "cron:update", Resource: "cron", Action: "update"},
						{Code: "cron:delete", Resource: "cron", Action: "delete"},
						{Code: "dept:read", Resource: "dept", Action: "read"},
					}, nil)
			},
			wantCount: 5,
			wantErr:   false,
		},
		{
			name: "空权限列表",
			mockSetup: func(m *MockPermissionRepository) {
				m.On("FindAll", mock.Anything).
					Return([]domain.Permission{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPermissionRepository)
			tt.mockSetup(mockRepo)

			svc := NewPermissionService(mockRepo)
			perms, err := svc.GetAllPermissions(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, perms, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
