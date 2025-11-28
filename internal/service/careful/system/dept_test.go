/**
 * Description：
 * FileName：dept_test.go.go
 * Author：CJiaの用心
 * Create：2025/11/28 10:50:38
 * Remark：
 */

package system

import (
	"context"
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	repomocks "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/mocks"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func strPtr(s string) *string {
	return &s // 函数参数 s 是临时变量，有内存地址，可取址
}

func Test_deptService_Create(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositorySystem.DeptRepository
		domain  domainSystem.Dept
		wantErr error
	}{
		{
			name: "创建成功",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "").
					Return(false, nil)
				repo.EXPECT().CheckExistByNameAndParentId(gomock.Any(), "部门名称", "root", "").
					Return(false, nil)
				repo.EXPECT().GetByParentId(gomock.Any(), "root").
					Return(domainSystem.Dept{
						Dept: system.Dept{
							CoreModels: models.CoreModels{
								Id: "root",
							},
							Status: true,
							Name:   "父部门名称",
							Code:   "父部门编码",
							Level:  0,
							Path:   "/",
						},
					}, nil)
				repo.EXPECT().Create(gomock.Any(), domainSystem.Dept{
					Dept: system.Dept{
						Name:     "部门名称",
						Code:     "部门编码",
						ParentID: strPtr("root"),
						Level:    1,
						Path:     "/root/",
					},
				}).Return(domainSystem.Dept{}, nil)
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					Name:     "部门名称",
					Code:     "部门编码",
					ParentID: strPtr("root"),
				},
			},
			wantErr: nil,
		},
		{
			name: "部门编码已存在",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "").
					Return(true, nil)
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					Name:     "部门名称",
					Code:     "部门编码",
					ParentID: strPtr("root"),
				},
			},
			wantErr: errors.New("部门编码已存在"),
		},
		{
			name: "同级别下已存在相同的部门信息",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "").
					Return(false, nil)
				repo.EXPECT().CheckExistByNameAndParentId(gomock.Any(), "部门名称", "root", "").
					Return(true, nil)
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					Name:     "部门名称",
					Code:     "部门编码",
					ParentID: strPtr("root"),
				},
			},
			wantErr: errors.New("同级别下已存在相同的部门信息"),
		},
		{
			name: "父部门信息不存在",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "").
					Return(false, nil)
				repo.EXPECT().CheckExistByNameAndParentId(gomock.Any(), "部门名称", "root", "").
					Return(false, nil)
				repo.EXPECT().GetByParentId(gomock.Any(), "root").
					Return(domainSystem.Dept{}, repositorySystem.ErrDeptParentNotFound)
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					Name:     "部门名称",
					Code:     "部门编码",
					ParentID: strPtr("root"),
				},
			},
			wantErr: repositorySystem.ErrDeptParentNotFound,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "").
					Return(false, nil)
				repo.EXPECT().CheckExistByNameAndParentId(gomock.Any(), "部门名称", "root", "").
					Return(false, nil)
				repo.EXPECT().GetByParentId(gomock.Any(), "root").
					Return(domainSystem.Dept{
						Dept: system.Dept{
							CoreModels: models.CoreModels{
								Id: "root",
							},
							Status: true,
							Name:   "父部门名称",
							Code:   "父部门编码",
							Level:  0,
							Path:   "/",
						},
					}, nil)
				repo.EXPECT().Create(gomock.Any(), domainSystem.Dept{
					Dept: system.Dept{
						Name:     "部门名称",
						Code:     "部门编码",
						ParentID: strPtr("root"),
						Level:    1,
						Path:     "/root/",
					},
				}).Return(domainSystem.Dept{}, errors.New("数据库异常"))
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					Name:     "部门名称",
					Code:     "部门编码",
					ParentID: strPtr("root"),
				},
			},
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewDeptService(tc.mock(ctrl))
			err := dictSvc.Create(context.Background(), tc.domain)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_deptService_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositorySystem.DeptRepository
		id      string
		wantErr error
	}{
		{
			name: "删除成功",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetChildCount(gomock.Any(), "1").Return(int64(0), nil)
				repo.EXPECT().GetUserCount(gomock.Any(), "1").Return(int64(0), nil)
				repo.EXPECT().Delete(gomock.Any(), "1").Return(nil)
				return repo
			},
			id:      "1",
			wantErr: nil,
		},
		{
			name: "部门含有子部门，无法删除",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetChildCount(gomock.Any(), "1").Return(int64(1), ErrDeptHasChildren)
				return repo
			},
			id:      "1",
			wantErr: ErrDeptHasChildren,
		},
		{
			name: "部门下仍有用户，无法删除",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetChildCount(gomock.Any(), "1").Return(int64(0), nil)
				repo.EXPECT().GetUserCount(gomock.Any(), "1").Return(int64(1), ErrDeptHasUsers)
				return repo
			},
			id:      "1",
			wantErr: ErrDeptHasUsers,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetChildCount(gomock.Any(), "1").Return(int64(0), nil)
				repo.EXPECT().GetUserCount(gomock.Any(), "1").Return(int64(0), nil)
				repo.EXPECT().Delete(gomock.Any(), "1").Return(errors.New("数据库异常"))
				return repo
			},
			id:      "1",
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewDeptService(tc.mock(ctrl))
			err := dictSvc.Delete(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_deptService_Update(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositorySystem.DeptRepository
		domain  domainSystem.Dept
		wantErr error
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "1").
					Return(false, nil)
				repo.EXPECT().CheckExistByNameAndParentId(gomock.Any(), "部门名称", "root", "1").
					Return(false, nil)
				repo.EXPECT().GetByParentId(gomock.Any(), "root").
					Return(domainSystem.Dept{
						Dept: system.Dept{
							CoreModels: models.CoreModels{
								Id: "root",
							},
							Status: true,
							Name:   "父部门名称",
							Code:   "父部门编码",
							Level:  0,
							Path:   "/",
						},
					}, nil)
				repo.EXPECT().GetAncestors(gomock.Any(), domainSystem.Dept{
					Dept: system.Dept{
						CoreModels: models.CoreModels{Id: "1"},
						Name:       "部门名称",
						Code:       "部门编码",
						ParentID:   strPtr("root"),
					},
				}).Return(nil, nil)
				repo.EXPECT().Update(gomock.Any(), domainSystem.Dept{
					Dept: system.Dept{
						CoreModels: models.CoreModels{Id: "1"},
						Name:       "部门名称",
						Code:       "部门编码",
						ParentID:   strPtr("root"),
						Level:      1,
						Path:       "/root/",
					},
				}).Return(nil)
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					CoreModels: models.CoreModels{Id: "1"},
					Name:       "部门名称",
					Code:       "部门编码",
					ParentID:   strPtr("root"),
				},
			},
			wantErr: nil,
		},
		{
			name: "部门编码已存在",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "1").
					Return(true, nil)
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					CoreModels: models.CoreModels{Id: "1"},
					Name:       "部门名称",
					Code:       "部门编码",
					ParentID:   strPtr("root"),
				},
			},
			wantErr: errors.New("部门编码已存在"),
		},
		{
			name: "同级别下已存在相同的部门信息",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "1").
					Return(false, nil)
				repo.EXPECT().CheckExistByNameAndParentId(gomock.Any(), "部门名称", "root", "1").
					Return(true, nil)
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					CoreModels: models.CoreModels{Id: "1"},
					Name:       "部门名称",
					Code:       "部门编码",
					ParentID:   strPtr("root"),
				},
			},
			wantErr: errors.New("同级别下已存在相同的部门信息"),
		},
		{
			name: "父部门信息不存在",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "1").
					Return(false, nil)
				repo.EXPECT().CheckExistByNameAndParentId(gomock.Any(), "部门名称", "root", "1").
					Return(false, nil)
				repo.EXPECT().GetByParentId(gomock.Any(), "root").
					Return(domainSystem.Dept{}, repositorySystem.ErrDeptParentNotFound)
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					CoreModels: models.CoreModels{Id: "1"},
					Name:       "部门名称",
					Code:       "部门编码",
					ParentID:   strPtr("root"),
				},
			},
			wantErr: repositorySystem.ErrDeptParentNotFound,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "部门编码", "1").
					Return(false, nil)
				repo.EXPECT().CheckExistByNameAndParentId(gomock.Any(), "部门名称", "root", "1").
					Return(false, nil)
				repo.EXPECT().GetByParentId(gomock.Any(), "root").
					Return(domainSystem.Dept{
						Dept: system.Dept{
							CoreModels: models.CoreModels{
								Id: "root",
							},
							Status: true,
							Name:   "父部门名称",
							Code:   "父部门编码",
							Level:  0,
							Path:   "/",
						},
					}, nil)
				repo.EXPECT().GetAncestors(gomock.Any(), domainSystem.Dept{
					Dept: system.Dept{
						CoreModels: models.CoreModels{Id: "1"},
						Name:       "部门名称",
						Code:       "部门编码",
						ParentID:   strPtr("root"),
					},
				}).Return(nil, nil)
				repo.EXPECT().Update(gomock.Any(), domainSystem.Dept{
					Dept: system.Dept{
						CoreModels: models.CoreModels{Id: "1"},
						Name:       "部门名称",
						Code:       "部门编码",
						ParentID:   strPtr("root"),
						Level:      1,
						Path:       "/root/",
					},
				}).Return(errors.New("数据库异常"))
				return repo
			},
			domain: domainSystem.Dept{
				Dept: system.Dept{
					CoreModels: models.CoreModels{Id: "1"},
					Name:       "部门名称",
					Code:       "部门编码",
					ParentID:   strPtr("root"),
				},
			},
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewDeptService(tc.mock(ctrl))
			err := dictSvc.Update(context.Background(), tc.domain)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_deptService_GetById(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositorySystem.DeptRepository
		id      string
		wantErr error
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Dept{
						Dept: system.Dept{
							CoreModels: models.CoreModels{
								Id: "1",
							},
							Code: "部门编码",
							Name: "部门名称",
						},
					}, nil)
				return repo
			},
			id:      "1",
			wantErr: nil,
		},
		{
			name: "部门信息不存在",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Dept{}, repositorySystem.ErrDeptNotFound)
				return repo
			},
			id:      "1",
			wantErr: repositorySystem.ErrDeptNotFound,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositorySystem.DeptRepository {
				repo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Dept{}, errors.New("数据库异常"))
				return repo
			},
			id:      "1",
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewDeptService(tc.mock(ctrl))
			_, err := dictSvc.GetById(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
