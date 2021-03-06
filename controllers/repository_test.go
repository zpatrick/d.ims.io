package controllers

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/golang/mock/gomock"
	"github.com/quintilesims/d.ims.io/mock"
	"github.com/quintilesims/d.ims.io/models"
)

func TestCreateRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewRepositoryController(mockECR, mockAccountManager)

	validateCreateRepositoryInput := func(input *ecr.CreateRepositoryInput) {
		if v, want := aws.StringValue(input.RepositoryName), "user/test"; v != want {
			t.Errorf("Name was '%v', expected '%v'", v, want)
		}
	}

	input := &ecr.CreateRepositoryInput{}
	input.SetRepositoryName("user/test")
	mockECR.EXPECT().
		CreateRepository(input).
		Do(validateCreateRepositoryInput).
		Return(&ecr.CreateRepositoryOutput{}, nil)

	validateGetRepositoryPolicyInput := func(input *ecr.GetRepositoryPolicyInput) {
		if v, want := aws.StringValue(input.RepositoryName), "user/test"; v != want {
			t.Errorf("Name was '%v', expected '%v'", v, want)
		}
	}

	getPolicyInput := &ecr.GetRepositoryPolicyInput{}
	getPolicyInput.SetRepositoryName("user/test")
	mockECR.EXPECT().
		GetRepositoryPolicy(getPolicyInput).
		Do(validateGetRepositoryPolicyInput).
		Return(&ecr.GetRepositoryPolicyOutput{}, nil)

	mockAccountManager.EXPECT().
		Accounts().
		Return([]string{"1", "2", "3"}, nil)

	validateSetRepositoryPolicyInput := func(input *ecr.SetRepositoryPolicyInput) {
		if v, want := aws.StringValue(input.RepositoryName), "user/test"; v != want {
			t.Errorf("Name was '%v', expected '%v'", v, want)
		}

		if len(aws.StringValue(input.PolicyText)) == 0 {
			t.Error("Policy text expected to be not empty")
		}
	}

	policyDoc := models.PolicyDocument{}
	policyDoc.AddAWSAccountPrincipal("1")
	policyDoc.AddAWSAccountPrincipal("2")
	policyDoc.AddAWSAccountPrincipal("3")
	policyText, err := policyDoc.RenderPolicyText()
	if err != nil {
		t.Fatal(err)
	}

	setPolicyInput := &ecr.SetRepositoryPolicyInput{}
	setPolicyInput.SetRepositoryName("user/test")
	setPolicyInput.SetPolicyText(policyText)
	mockECR.EXPECT().
		SetRepositoryPolicy(setPolicyInput).
		Do(validateSetRepositoryPolicyInput).
		Return(&ecr.SetRepositoryPolicyOutput{}, nil)

	c := generateContext(t, models.CreateRepositoryRequest{Name: "test"}, map[string]string{"owner": "user"})
	if _, err := controller.CreateRepository(c); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRepositoryFailsWithSlashes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewRepositoryController(mockECR, mockAccountManager)

	c := generateContext(t, models.CreateRepositoryRequest{Name: "slash/test"}, map[string]string{"owner": "user"})
	_, err := controller.CreateRepository(c)
	if !strings.ContainsAny(err.Error(), "cannot contain '/' characters") {
		t.Fatal(err)
	}
}

func TestDeleteRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewRepositoryController(mockECR, mockAccountManager)

	validateDeleteRepositoryInput := func(input *ecr.DeleteRepositoryInput) {
		if v, want := aws.StringValue(input.RepositoryName), "user/test"; v != want {
			t.Errorf("Name was '%v', expected '%v'", v, want)
		}

		if v, want := aws.BoolValue(input.Force), true; v != want {
			t.Errorf("Force was '%v', expected '%v'", v, want)
		}
	}

	mockECR.EXPECT().
		DeleteRepository(gomock.Any()).
		Do(validateDeleteRepositoryInput).
		Return(&ecr.DeleteRepositoryOutput{}, nil)

	c := generateContext(t, nil, map[string]string{"name": "test", "owner": "user"})
	if _, err := controller.DeleteRepository(c); err != nil {
		t.Fatal(err)
	}
}

func TestGetRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewRepositoryController(mockECR, mockAccountManager)

	validateDescribeRepositoriesInput := func(input *ecr.DescribeRepositoriesInput) {
		if v, want := aws.StringValue(input.RepositoryNames[0]), "user/test"; v != want {
			t.Errorf("Name was '%v', expected '%v'", v, want)
		}
	}

	output := &ecr.DescribeRepositoriesOutput{
		Repositories: []*ecr.Repository{&ecr.Repository{}},
	}

	mockECR.EXPECT().
		DescribeRepositories(gomock.Any()).
		Do(validateDescribeRepositoriesInput).
		Return(output, nil)

	c := generateContext(t, nil, map[string]string{"name": "test", "owner": "user"})
	if _, err := controller.GetRepository(c); err != nil {
		t.Fatal(err)
	}
}

func TestListRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewRepositoryController(mockECR, mockAccountManager)

	mockECR.EXPECT().
		DescribeRepositoriesPages(gomock.Any(), gomock.Any()).
		Return(nil)

	c := generateContext(t, nil, nil)
	if _, err := controller.ListRepositories(c); err != nil {
		t.Fatal(err)
	}
}

func TestListRepositoryImages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewRepositoryController(mockECR, mockAccountManager)

	validateListImagesInput := func(input *ecr.ListImagesInput, fn func(output *ecr.ListImagesOutput, lastPage bool) bool) {
		if v, want := aws.StringValue(input.RepositoryName), "user/test"; v != want {
			t.Errorf("Name was '%v', expected '%v'", v, want)
		}
	}

	mockECR.EXPECT().
		ListImagesPages(gomock.Any(), gomock.Any()).
		Do(validateListImagesInput).
		Return(nil)

	c := generateContext(t, nil, map[string]string{"name": "test", "owner": "user"})
	if _, err := controller.ListRepositoryImages(c); err != nil {
		t.Fatal(err)
	}
}

func TestGetRepositoryImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewRepositoryController(mockECR, mockAccountManager)

	validateDescribeImagesInput := func(input *ecr.DescribeImagesInput) {
		if v, want := aws.StringValue(input.RepositoryName), "user/test"; v != want {
			t.Errorf("Name was '%v', expected '%v'", v, want)
		}

		if v, want := aws.StringValue(input.ImageIds[0].ImageTag), "latest"; v != want {
			t.Errorf("Name was '%v', expected '%v'", v, want)
		}
	}

	output := &ecr.DescribeImagesOutput{
		ImageDetails: []*ecr.ImageDetail{&ecr.ImageDetail{}},
	}

	mockECR.EXPECT().
		DescribeImages(gomock.Any()).
		Do(validateDescribeImagesInput).
		Return(output, nil)

	c := generateContext(t, nil, map[string]string{"name": "test", "owner": "user", "tag": "latest"})
	if _, err := controller.GetRepositoryImage(c); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteRepositoryImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECR := mock.NewMockECRAPI(ctrl)
	mockAccountManager := mock.NewMockAccountManager(ctrl)
	controller := NewRepositoryController(mockECR, mockAccountManager)

	validateBatchDeleteImageInput := func(input *ecr.BatchDeleteImageInput) {
		if v, want := aws.StringValue(input.RepositoryName), "user/test"; v != want {
			t.Errorf("Name was '%v', expected '%v'", v, want)
		}

		if v, want := aws.StringValue(input.ImageIds[0].ImageTag), "latest"; v != want {
			t.Errorf("Tag was '%v', expected '%v'", v, want)
		}
	}

	mockECR.EXPECT().
		BatchDeleteImage(gomock.Any()).
		Do(validateBatchDeleteImageInput).
		Return(&ecr.BatchDeleteImageOutput{}, nil)

	c := generateContext(t, nil, map[string]string{"name": "test", "owner": "user", "tag": "latest"})
	if _, err := controller.DeleteRepositoryImage(c); err != nil {
		t.Fatal(err)
	}
}
