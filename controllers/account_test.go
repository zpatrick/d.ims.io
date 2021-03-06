package controllers

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/golang/mock/gomock"
	"github.com/quintilesims/d.ims.io/mock"
	"github.com/quintilesims/d.ims.io/models"
	"github.com/stretchr/testify/assert"
)

func TestGrantAccessInputValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewAccountController(mockECR, mockAccountManager)

	c := generateContext(t, models.GrantAccessRequest{Account: ""}, nil)
	resp, err := controller.GrantAccess(c)
	if err != nil {
		t.Fatal(err)
	}

	recorder := unmarshalBody(t, resp, nil)
	assert.Equal(t, 400, recorder.Code)
}

func TestGrantAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewAccountController(mockECR, mockAccountManager)

	fnListRepos := func(input *ecr.DescribeRepositoriesInput, fn func(output *ecr.DescribeRepositoriesOutput, lastPage bool) bool) error {
		output := &ecr.DescribeRepositoriesOutput{
			Repositories: []*ecr.Repository{
				{RepositoryName: aws.String("user/name-*")},
				{RepositoryName: aws.String("user/name-*")},
			},
		}

		fn(output, false)
		return nil
	}

	mockECR.EXPECT().
		DescribeRepositoriesPages(gomock.Any(), gomock.Any()).
		Do(fnListRepos).
		Return(nil)

	mockAccountManager.EXPECT().
		Accounts().
		Return([]string{"account-id"}, nil)

	getPolicyInput := &ecr.GetRepositoryPolicyInput{}
	getPolicyInput.SetRepositoryName("user/name-*")
	mockECR.EXPECT().
		GetRepositoryPolicy(getPolicyInput).
		Return(&ecr.GetRepositoryPolicyOutput{}, nil).
		Times(2)

	policyDoc := models.PolicyDocument{}
	policyDoc.AddAWSAccountPrincipal("account-id")
	policyText, err := policyDoc.RenderPolicyText()
	if err != nil {
		t.Fatal(err)
	}

	setPolicyInput := &ecr.SetRepositoryPolicyInput{}
	setPolicyInput.SetRepositoryName("user/name-*")
	setPolicyInput.SetPolicyText(policyText)
	mockECR.EXPECT().
		SetRepositoryPolicy(setPolicyInput).
		Return(&ecr.SetRepositoryPolicyOutput{}, nil).
		Times(2)

	mockAccountManager.EXPECT().
		GrantAccess("account-id").
		Return(nil)

	c := generateContext(t, models.GrantAccessRequest{Account: "account-id"}, nil)
	if _, err := controller.GrantAccess(c); err != nil {
		t.Fatal(err)
	}
}

func TestRevokeAccessInputValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewAccountController(mockECR, mockAccountManager)

	c := generateContext(t, nil, nil)
	resp, err := controller.RevokeAccess(c)
	if err != nil {
		t.Fatal(err)
	}

	recorder := unmarshalBody(t, resp, nil)
	assert.Equal(t, 400, recorder.Code)
}

func TestRevokeAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewAccountController(mockECR, mockAccountManager)

	fnListRepos := func(input *ecr.DescribeRepositoriesInput, fn func(output *ecr.DescribeRepositoriesOutput, lastPage bool) bool) error {
		output := &ecr.DescribeRepositoriesOutput{
			Repositories: []*ecr.Repository{
				{RepositoryName: aws.String("user/name-1")},
			},
		}

		fn(output, false)
		return nil
	}

	mockECR.EXPECT().
		DescribeRepositoriesPages(gomock.Any(), gomock.Any()).
		Do(fnListRepos).
		Return(nil)

	policyDoc := &models.PolicyDocument{}
	policyDoc.AddAWSAccountPrincipal("account-id")
	policyText, err := policyDoc.RenderPolicyText()
	if err != nil {
		t.Fatal(err)
	}

	getPolicyOutput := &ecr.GetRepositoryPolicyOutput{}
	getPolicyOutput.SetPolicyText(policyText)

	getPolicyInput := &ecr.GetRepositoryPolicyInput{}
	getPolicyInput.SetRepositoryName("user/name-1")
	mockECR.EXPECT().
		GetRepositoryPolicy(getPolicyInput).
		Return(getPolicyOutput, nil)

	setPolicyInput := &ecr.SetRepositoryPolicyInput{}
	setPolicyInput.SetRepositoryName("user/name-1")
	setPolicyInput.SetPolicyText("")
	mockECR.EXPECT().
		SetRepositoryPolicy(setPolicyInput).
		Return(&ecr.SetRepositoryPolicyOutput{}, nil)

	mockAccountManager.EXPECT().
		RevokeAccess(gomock.Any()).
		Return(nil)

	c := generateContext(t, nil, map[string]string{"id": "account-id"})
	if _, err := controller.RevokeAccess(c); err != nil {
		t.Fatal(err)
	}
}

func TestAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewAccountController(mockECR, mockAccountManager)

	mockAccountManager.EXPECT().
		Accounts().
		Return([]string{}, nil)

	c := generateContext(t, nil, nil)
	if _, err := controller.ListAccounts(c); err != nil {
		t.Fatal(err)
	}
}
