package provider

import (
	"bitbucket.org/booqsoftware/terraform-provider-qbee/internal/qbee"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FiledistributionFile struct {
	Command      types.String `tfsdk:"command"`
	PreCondition types.String `tfsdk:"pre_condition"`
	Templates    types.List   `tfsdk:"templates"`
	Parameters   types.List   `tfsdk:"parameters"`
}

func (f FiledistributionFile) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"command":       types.StringType,
		"pre_condition": types.StringType,
		"templates": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: FiledistributionTemplate{}.attrTypes(),
			},
		},
		"parameters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: FiledistributionParameter{}.attrTypes(),
			},
		},
	}
}

type FiledistributionParameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (f FiledistributionParameter) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}
}

type FiledistributionTemplate struct {
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	IsTemplate  types.Bool   `tfsdk:"is_template"`
}

func (f FiledistributionTemplate) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source":      types.StringType,
		"destination": types.StringType,
		"is_template": types.BoolType,
	}
}

func PlanToQbeeFilesets(ctx context.Context, files []FiledistributionFile) []qbee.FilesetConfig {
	var filesets []qbee.FilesetConfig

	for _, file := range files {
		var fsc qbee.FilesetConfig

		if !file.PreCondition.IsNull() {
			fsc.PreCondition = file.PreCondition.ValueString()
		}

		if !file.Command.IsNull() {
			fsc.Command = file.Command.ValueString()
		}

		var paramValues []FiledistributionParameter
		file.Parameters.ElementsAs(ctx, &paramValues, false)
		var params []qbee.FilesetParameter
		for _, value := range paramValues {
			params = append(params, qbee.FilesetParameter{
				Key:   value.Key.ValueString(),
				Value: value.Value.ValueString(),
			})
		}

		var templateValues []FiledistributionTemplate
		file.Templates.ElementsAs(ctx, &templateValues, false)
		var templates []qbee.FilesetTemplate
		for _, value := range templateValues {
			templates = append(templates, qbee.FilesetTemplate{
				Source:      value.Source.ValueString(),
				Destination: value.Destination.ValueString(),
				IsTemplate:  value.IsTemplate.ValueBool(),
			})
		}

		filesets = append(filesets, qbee.FilesetConfig{
			PreCondition: file.PreCondition.ValueString(),
			Command:      file.Command.ValueString(),
			Templates:    templates,
			Parameters:   params,
		})
	}

	return filesets
}
