/**
 * Description：
 * FileName：dict_type_test.go.go
 * Author：CJiaの用心
 * Create：2025/10/20 11:13:19
 * Remark：
 */

package tools

import (
	"context"
	"errors"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	repomocks "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/mocks"
	repositoryTools "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_dictTypeService_Create(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository)
		domain  domainTools.DictType
		wantErr error
	}{
		{
			name: "创建成功",
			mock: func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository) {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				dictRepo := repomocks.NewMockDictRepository(ctrl)
				dictRepo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{
						Dict: tools.Dict{
							CoreModels: models.CoreModels{
								Id: "1",
							},
						},
					}, nil)
				repo.EXPECT().Create(gomock.Any(), domainTools.DictType{
					DictType: tools.DictType{
						Name:   "男",
						DictId: "1",
					},
				}).Return(domainTools.DictType{
					DictType: tools.DictType{
						Name:   "男",
						DictId: "1",
					},
				}, nil)
				return repo, dictRepo
			},
			domain: domainTools.DictType{
				DictType: tools.DictType{
					Name:   "男",
					DictId: "1",
				},
			},
			wantErr: nil,
		},
		{
			name: "数据字典不存在",
			mock: func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository) {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				dictRepo := repomocks.NewMockDictRepository(ctrl)
				dictRepo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{}, repositoryTools.ErrDictNotFound)
				return repo, dictRepo
			},
			domain: domainTools.DictType{
				DictType: tools.DictType{
					Name:   "男",
					DictId: "1",
				},
			},
			wantErr: repositoryTools.ErrDictNotFound,
		},
		{
			name: "违反唯一约束",
			mock: func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository) {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				dictRepo := repomocks.NewMockDictRepository(ctrl)
				dictRepo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{
						Dict: tools.Dict{
							CoreModels: models.CoreModels{
								Id: "1",
							},
						},
					}, nil)
				repo.EXPECT().Create(gomock.Any(), domainTools.DictType{
					DictType: tools.DictType{
						Name:   "男",
						DictId: "1",
					},
				}).Return(domainTools.DictType{}, repositoryTools.ErrDictTypeDuplicate)
				return repo, dictRepo
			},
			domain: domainTools.DictType{
				DictType: tools.DictType{
					Name:   "男",
					DictId: "1",
				},
			},
			wantErr: repositoryTools.ErrDictTypeDuplicate,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository) {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				dictRepo := repomocks.NewMockDictRepository(ctrl)
				dictRepo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{
						Dict: tools.Dict{
							CoreModels: models.CoreModels{
								Id: "1",
							},
						},
					}, nil)
				repo.EXPECT().Create(gomock.Any(), domainTools.DictType{
					DictType: tools.DictType{
						Name:   "男",
						DictId: "1",
					},
				}).Return(domainTools.DictType{}, errors.New("数据库异常"))
				return repo, dictRepo
			},
			domain: domainTools.DictType{
				DictType: tools.DictType{
					Name:   "男",
					DictId: "1",
				},
			},
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewDictTypeService(tc.mock(ctrl))
			err := dictSvc.Create(context.Background(), tc.domain)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_dictTypeService_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositoryTools.DictTypeRepository
		id      string
		wantErr error
	}{
		{
			name: "删除成功",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictTypeRepository {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				repo.EXPECT().Delete(gomock.Any(), "1").Return(nil)
				return repo
			},
			id:      "1",
			wantErr: nil,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictTypeRepository {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
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

			dictSvc := NewDictTypeService(tc.mock(ctrl), nil)
			err := dictSvc.Delete(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_dictTypeService_Update(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository)
		domain  domainTools.DictType
		wantErr error
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository) {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				dictRepo := repomocks.NewMockDictRepository(ctrl)
				dictRepo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{
						Dict: tools.Dict{
							CoreModels: models.CoreModels{
								Id: "1",
							},
						},
					}, nil)
				repo.EXPECT().Update(gomock.Any(), domainTools.DictType{
					DictType: tools.DictType{
						Name:   "男",
						DictId: "1",
					},
				}).Return(nil)
				return repo, dictRepo
			},
			domain: domainTools.DictType{
				DictType: tools.DictType{
					Name:   "男",
					DictId: "1",
				},
			},
			wantErr: nil,
		},
		{
			name: "数据字典不存在",
			mock: func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository) {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				dictRepo := repomocks.NewMockDictRepository(ctrl)
				dictRepo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{}, repositoryTools.ErrDictNotFound)
				return repo, dictRepo
			},
			domain: domainTools.DictType{
				DictType: tools.DictType{
					Name:   "男",
					DictId: "1",
				},
			},
			wantErr: repositoryTools.ErrDictNotFound,
		},
		{
			name: "字典信息不存在",
			mock: func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository) {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				dictRepo := repomocks.NewMockDictRepository(ctrl)
				dictRepo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{
						Dict: tools.Dict{
							CoreModels: models.CoreModels{
								Id: "1",
							},
						},
					}, nil)
				repo.EXPECT().Update(gomock.Any(), domainTools.DictType{
					DictType: tools.DictType{
						Name:   "男",
						DictId: "1",
					},
				}).Return(repositoryTools.ErrDictTypeNotFound)
				return repo, dictRepo
			},
			domain: domainTools.DictType{
				DictType: tools.DictType{
					Name:   "男",
					DictId: "1",
				},
			},
			wantErr: repositoryTools.ErrDictTypeNotFound,
		},
		{
			name: "数据已被修改，请刷新后重试",
			mock: func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository) {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				dictRepo := repomocks.NewMockDictRepository(ctrl)
				dictRepo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{
						Dict: tools.Dict{
							CoreModels: models.CoreModels{
								Id: "1",
							},
						},
					}, nil)
				repo.EXPECT().Update(gomock.Any(), domainTools.DictType{
					DictType: tools.DictType{
						Name:   "男",
						DictId: "1",
					},
				}).Return(repositoryTools.ErrDictTypeVersionInconsistency)
				return repo, dictRepo
			},
			domain: domainTools.DictType{
				DictType: tools.DictType{
					Name:   "男",
					DictId: "1",
				},
			},
			wantErr: repositoryTools.ErrDictTypeVersionInconsistency,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) (repositoryTools.DictTypeRepository, repositoryTools.DictRepository) {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				dictRepo := repomocks.NewMockDictRepository(ctrl)
				dictRepo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{
						Dict: tools.Dict{
							CoreModels: models.CoreModels{
								Id: "1",
							},
						},
					}, nil)
				repo.EXPECT().Update(gomock.Any(), domainTools.DictType{
					DictType: tools.DictType{
						Name:   "男",
						DictId: "1",
					},
				}).Return(errors.New("数据库异常"))
				return repo, dictRepo
			},
			domain: domainTools.DictType{
				DictType: tools.DictType{
					Name:   "男",
					DictId: "1",
				},
			},
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewDictTypeService(tc.mock(ctrl))
			err := dictSvc.Update(context.Background(), tc.domain)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_dictTypeService_GetById(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositoryTools.DictTypeRepository
		id      string
		wantErr error
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictTypeRepository {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.DictType{
						DictType: tools.DictType{
							CoreModels: models.CoreModels{
								Id: "1",
							},
							Name: "男",
						},
					}, nil)
				return repo
			},
			id:      "1",
			wantErr: nil,
		},
		{
			name: "字典不存在",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictTypeRepository {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.DictType{}, repositoryTools.ErrDictNotFound)
				return repo
			},
			id:      "1",
			wantErr: repositoryTools.ErrDictNotFound,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictTypeRepository {
				repo := repomocks.NewMockDictTypeRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.DictType{}, errors.New("数据库异常"))
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

			dictSvc := NewDictTypeService(tc.mock(ctrl), nil)
			_, err := dictSvc.GetById(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
