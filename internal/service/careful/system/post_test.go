/**
 * Description：
 * FileName：post_test.go.go
 * Author：CJiaの用心
 * Create：2025/12/2 01:09:54
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

func Test_postService_Create(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository)
		domain  domainSystem.Post
		wantErr error
	}{
		{
			name: "创建成功",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "岗位编码", "").
					Return(false, nil)
				repo.EXPECT().Create(gomock.Any(), domainSystem.Post{
					Post: system.Post{
						Name:     "岗位名称",
						Code:     "岗位编码",
						PostType: 5,
						Level:    4,
					},
				}).Return(domainSystem.Post{}, nil)
				return repo, deptRepo
			},
			domain: domainSystem.Post{
				Post: system.Post{
					Name:     "岗位名称",
					Code:     "岗位编码",
					PostType: 5,
					Level:    4,
				},
			},
			wantErr: nil,
		},
		{
			name: "岗位编码已存在",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "岗位编码", "").
					Return(true, nil)
				return repo, deptRepo
			},
			domain: domainSystem.Post{
				Post: system.Post{
					Name:     "岗位名称",
					Code:     "岗位编码",
					PostType: 5,
					Level:    4,
				},
			},
			wantErr: errors.New("岗位编码已存在"),
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "岗位编码", "").
					Return(false, nil)
				repo.EXPECT().Create(gomock.Any(), domainSystem.Post{
					Post: system.Post{
						Name:     "岗位名称",
						Code:     "岗位编码",
						PostType: 5,
						Level:    4,
					},
				}).Return(domainSystem.Post{}, errors.New("数据库异常"))
				return repo, deptRepo
			},
			domain: domainSystem.Post{
				Post: system.Post{
					Name:     "岗位名称",
					Code:     "岗位编码",
					PostType: 5,
					Level:    4,
				},
			},
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewPostService(tc.mock(ctrl))
			err := dictSvc.Create(context.Background(), tc.domain)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_postService_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository)
		id      string
		wantErr error
	}{
		{
			name: "删除成功",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetUserCount(gomock.Any(), "1").Return(int64(0))
				repo.EXPECT().Delete(gomock.Any(), "1").Return(nil)
				return repo, deptRepo
			},
			id:      "1",
			wantErr: nil,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetUserCount(gomock.Any(), "1").Return(int64(0))
				repo.EXPECT().Delete(gomock.Any(), "1").Return(errors.New("数据库异常"))
				return repo, deptRepo
			},
			id:      "1",
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewPostService(tc.mock(ctrl))
			err := dictSvc.Delete(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_postService_Update(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository)
		domain  domainSystem.Post
		wantErr error
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "岗位编码", "1").
					Return(false, nil)
				repo.EXPECT().Update(gomock.Any(), domainSystem.Post{
					Post: system.Post{
						CoreModels: models.CoreModels{Id: "1"},
						Name:       "岗位名称",
						Code:       "岗位编码",
						PostType:   5,
						Level:      4,
					},
				}).Return(nil)
				return repo, deptRepo
			},
			domain: domainSystem.Post{
				Post: system.Post{
					CoreModels: models.CoreModels{Id: "1"},
					Name:       "岗位名称",
					Code:       "岗位编码",
					PostType:   5,
					Level:      4,
				},
			},
			wantErr: nil,
		},
		{
			name: "岗位编码已存在",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "岗位编码", "1").
					Return(true, nil)
				return repo, deptRepo
			},
			domain: domainSystem.Post{
				Post: system.Post{
					CoreModels: models.CoreModels{Id: "1"},
					Name:       "岗位名称",
					Code:       "岗位编码",
					PostType:   5,
					Level:      4,
				},
			},
			wantErr: errors.New("岗位编码已存在"),
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "岗位编码", "1").
					Return(false, nil)
				repo.EXPECT().Update(gomock.Any(), domainSystem.Post{
					Post: system.Post{
						CoreModels: models.CoreModels{Id: "1"},
						Name:       "岗位名称",
						Code:       "岗位编码",
						PostType:   5,
						Level:      4,
					},
				}).Return(errors.New("数据库异常"))
				return repo, deptRepo
			},
			domain: domainSystem.Post{
				Post: system.Post{
					CoreModels: models.CoreModels{Id: "1"},
					Name:       "岗位名称",
					Code:       "岗位编码",
					PostType:   5,
					Level:      4,
				},
			},
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewPostService(tc.mock(ctrl))
			err := dictSvc.Update(context.Background(), tc.domain)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_postService_GetById(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository)
		id      string
		wantErr error
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Post{
						Post: system.Post{
							CoreModels: models.CoreModels{
								Id: "1",
							},
							Code: "岗位编码",
							Name: "岗位名称",
						},
					}, nil)
				return repo, deptRepo
			},
			id:      "1",
			wantErr: nil,
		},
		{
			name: "岗位信息不存在",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Post{
						Post: system.Post{
							CoreModels: models.CoreModels{
								Id: "1",
							},
							Code: "岗位编码",
							Name: "岗位名称",
						},
					}, repositorySystem.ErrPostNotFound)
				return repo, deptRepo
			},
			id:      "1",
			wantErr: repositorySystem.ErrPostNotFound,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) (repositorySystem.PostRepository, repositorySystem.DeptRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				deptRepo := repomocks.NewMockDeptRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Post{
						Post: system.Post{
							CoreModels: models.CoreModels{
								Id: "1",
							},
							Code: "岗位编码",
							Name: "岗位名称",
						},
					}, errors.New("数据库异常"))
				return repo, deptRepo
			},
			id:      "1",
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewPostService(tc.mock(ctrl))
			_, err := dictSvc.GetById(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
