package internal

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func toInt32(value types.Int64) *int32 {
	if value.IsNull() {
		return nil
	}
	return aws.Int32(int32(value.Value))
}

func toString(value types.String) *string {
	if value.IsNull() {
		return aws.String("")
	}
	return aws.String(value.Value)
}

func toBool(value types.Bool) *bool {
	if value.IsNull() {
		return aws.Bool(false)
	}
	return aws.Bool(value.Value)
}
