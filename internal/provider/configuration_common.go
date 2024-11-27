package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.qbee.io/client/config"
)

func typeAndIdentifier(tag types.String, node types.String) (config.EntityType, string) {
	if tag.IsNull() {
		return config.EntityTypeNode, node.ValueString()
	} else {
		return config.EntityTypeTag, tag.ValueString()
	}
}

func nullableStringValue(value string) basetypes.StringValue {
	if value == "" {
		return types.StringNull()
	} else {
		return types.StringValue(value)
	}
}

func listFromStructs[T interface{}](ctx context.Context, items []T, elemType attr.Type) (types.List, diag.Diagnostics) {
	if len(items) == 0 {
		return types.ListNull(elemType), nil
	}

	itemsValue, diags := types.ListValueFrom(
		ctx,
		elemType,
		items,
	)

	return itemsValue, diags
}
