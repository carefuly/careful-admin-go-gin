/**
 * Description：
 * FileName：dict_test.go.go
 * Author：CJiaの用心
 * Create：2025/10/14 14:27:15
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

func Test_dictService_Create(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositoryTools.DictRepository
		domain  domainTools.Dict
		wantErr error
	}{
		{
			name: "创建成功",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "字典编码", "").
					Return(false, nil)
				repo.EXPECT().CheckExistByName(gomock.Any(), "字典名称", "").
					Return(false, nil)
				repo.EXPECT().Create(gomock.Any(), domainTools.Dict{
					Dict: tools.Dict{
						Code: "字典编码",
						Name: "字典名称",
					},
				}).Return(domainTools.Dict{
					Dict: tools.Dict{
						Code: "字典编码",
						Name: "字典名称",
					},
				}, nil)
				return repo
			},
			domain: domainTools.Dict{
				Dict: tools.Dict{
					Code: "字典编码",
					Name: "字典名称",
				},
			},
			wantErr: nil,
		},
		{
			name: "字典编码已存在",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "字典编码", "").
					Return(true, nil)
				return repo
			},
			domain: domainTools.Dict{
				Dict: tools.Dict{
					Code: "字典编码",
					Name: "字典名称",
				},
			},
			wantErr: errors.New("字典编码已存在"),
		},
		{
			name: "字典名称已存在",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "字典编码", "").
					Return(false, nil)
				repo.EXPECT().CheckExistByName(gomock.Any(), "字典名称", "").
					Return(true, nil)
				return repo
			},
			domain: domainTools.Dict{
				Dict: tools.Dict{
					Code: "字典编码",
					Name: "字典名称",
				},
			},
			wantErr: errors.New("字典名称已存在"),
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "字典编码", "").
					Return(false, nil)
				repo.EXPECT().CheckExistByName(gomock.Any(), "字典名称", "").
					Return(false, nil)
				repo.EXPECT().Create(gomock.Any(), domainTools.Dict{
					Dict: tools.Dict{
						Code: "字典编码",
						Name: "字典名称",
					},
				}).Return(domainTools.Dict{}, errors.New("数据库异常"))
				return repo
			},
			domain: domainTools.Dict{
				Dict: tools.Dict{
					Code: "字典编码",
					Name: "字典名称",
				},
			},
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewDictService(tc.mock(ctrl))
			err := dictSvc.Create(context.Background(), tc.domain)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_dictService_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositoryTools.DictRepository
		id      string
		wantErr error
	}{
		{
			name: "删除成功",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().Delete(gomock.Any(), "1").Return(nil)
				return repo
			},
			id:      "1",
			wantErr: nil,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
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

			dictSvc := NewDictService(tc.mock(ctrl))
			err := dictSvc.Delete(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_dictService_Update(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositoryTools.DictRepository
		domain  domainTools.Dict
		wantErr error
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "字典编码", "").
					Return(false, nil)
				repo.EXPECT().CheckExistByName(gomock.Any(), "字典名称", "").
					Return(false, nil)
				repo.EXPECT().Update(gomock.Any(), domainTools.Dict{
					Dict: tools.Dict{
						Code: "字典编码",
						Name: "字典名称",
					},
				}).Return(nil)
				return repo
			},
			domain: domainTools.Dict{
				Dict: tools.Dict{
					Code: "字典编码",
					Name: "字典名称",
				},
			},
			wantErr: nil,
		},
		{
			name: "字典编码已存在",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "字典编码", "1").
					Return(true, nil)
				return repo
			},
			domain: domainTools.Dict{
				Dict: tools.Dict{
					CoreModels: models.CoreModels{
						Id: "1",
					},
					Code: "字典编码",
					Name: "字典名称",
				},
			},
			wantErr: errors.New("字典编码已存在"),
		},
		{
			name: "字典名称已存在",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "字典编码", "1").
					Return(false, nil)
				repo.EXPECT().CheckExistByName(gomock.Any(), "字典名称", "1").
					Return(true, nil)
				return repo
			},
			domain: domainTools.Dict{
				Dict: tools.Dict{
					CoreModels: models.CoreModels{
						Id: "1",
					},
					Code: "字典编码",
					Name: "字典名称",
				},
			},
			wantErr: errors.New("字典名称已存在"),
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().CheckExistByCode(gomock.Any(), "字典编码", "1").
					Return(false, nil)
				repo.EXPECT().CheckExistByName(gomock.Any(), "字典名称", "1").
					Return(false, nil)
				repo.EXPECT().Update(gomock.Any(), domainTools.Dict{
					Dict: tools.Dict{
						CoreModels: models.CoreModels{
							Id: "1",
						},
						Code: "字典编码",
						Name: "字典名称",
					},
				}).Return(errors.New("数据库异常"))
				return repo
			},
			domain: domainTools.Dict{
				Dict: tools.Dict{
					CoreModels: models.CoreModels{
						Id: "1",
					},
					Code: "字典编码",
					Name: "字典名称",
				},
			},
			wantErr: errors.New("数据库异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dictSvc := NewDictService(tc.mock(ctrl))
			err := dictSvc.Update(context.Background(), tc.domain)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_dictService_GetById(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repositoryTools.DictRepository
		id      string
		wantErr error
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{
						Dict: tools.Dict{
							CoreModels: models.CoreModels{
								Id: "1",
							},
							Code: "字典编码",
							Name: "字典名称",
						},
					}, nil)
				return repo
			},
			id:      "1",
			wantErr: nil,
		},
		{
			name: "字典不存在",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{}, repositoryTools.ErrDictNotFound)
				return repo
			},
			id:      "1",
			wantErr: repositoryTools.ErrDictNotFound,
		},
		{
			name: "数据库异常",
			mock: func(ctrl *gomock.Controller) repositoryTools.DictRepository {
				repo := repomocks.NewMockDictRepository(ctrl)
				repo.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{}, errors.New("数据库异常"))
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

			dictSvc := NewDictService(tc.mock(ctrl))
			_, err := dictSvc.GetById(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
