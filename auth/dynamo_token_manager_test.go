package auth

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/mock/gomock"
	"github.com/quintilesims/d.ims.io/mock"
)

func TestDynamoCreateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamoDB := mock.NewMockDynamoDBAPI(ctrl)
	target := NewDynamoTokenManager("table", mockDynamoDB)

	validatePutItemInput := func(input *dynamodb.PutItemInput) {
		if v, want := aws.StringValue(input.TableName), "table"; v != want {
			t.Errorf("Table was '%v', expected '%v'", v, want)
		}

		if v, want := aws.StringValue(input.Item["User"].S), "user"; v != want {
			t.Errorf("Column 'User' was '%v', expected '%v'", v, want)
		}

		if input.Item["Token"].S == nil {
			t.Error("Column 'Token' was nil")
		}
	}

	mockDynamoDB.EXPECT().
		PutItem(gomock.Any()).
		Do(validatePutItemInput).
		Return(&dynamodb.PutItemOutput{}, nil)

	if _, err := target.CreateToken("user"); err != nil {
		t.Fatal(err)
	}
}

func TestDynamoDeleteToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamoDB := mock.NewMockDynamoDBAPI(ctrl)
	target := NewDynamoTokenManager("table", mockDynamoDB)

	validateDeleteItemInput := func(input *dynamodb.DeleteItemInput) {
		if v, want := aws.StringValue(input.TableName), "table"; v != want {
			t.Errorf("Table was '%v', expected '%v'", v, want)
		}

		if v, want := aws.StringValue(input.Key["Token"].S), "token"; v != want {
			t.Errorf("Key 'Token' was '%v', expected '%v'", v, want)
		}
	}

	mockDynamoDB.EXPECT().
		DeleteItem(gomock.Any()).
		Do(validateDeleteItemInput).
		Return(&dynamodb.DeleteItemOutput{}, nil)

	if err := target.DeleteToken("token"); err != nil {
		t.Fatal(err)
	}
}

func TestDynamoAuthenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamoDB := mock.NewMockDynamoDBAPI(ctrl)
	target := NewDynamoTokenManager("table", mockDynamoDB)

	validateGetItemInput := func(input *dynamodb.GetItemInput) {
		if v, want := aws.StringValue(input.TableName), "table"; v != want {
			t.Errorf("Table was '%v', expected '%v'", v, want)
		}

		if v, want := aws.BoolValue(input.ConsistentRead), true; v != want {
			t.Errorf("ConsistentRead was '%v', expected '%v'", v, want)
		}

		if input.Key["Token"].S == nil {
			t.Error("Key 'Token' was nil")
		}
	}

	cases := []struct {
		ExpectedResult bool
		Items          map[string]*dynamodb.AttributeValue
	}{
		{
			ExpectedResult: false,
			Items:          nil,
		},
		{
			ExpectedResult: true,
			Items: map[string]*dynamodb.AttributeValue{
				"token": &dynamodb.AttributeValue{},
			},
		},
	}

	for _, c := range cases {
		mockDynamoDB.EXPECT().
			GetItem(gomock.Any()).
			Do(validateGetItemInput).
			Return(&dynamodb.GetItemOutput{Item: c.Items}, nil)

		ok, err := target.Authenticate("user", "pass")
		if err != nil {
			t.Fatal(err)
		}

		if v, want := ok, c.ExpectedResult; v != want {
			t.Errorf("Result was '%v', expected '%v' (items: %v)", v, want, c.Items)
		}
	}
}
