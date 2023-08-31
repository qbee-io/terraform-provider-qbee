package provider

import (
	"bitbucket.org/booqsoftware/terraform-provider-qbee/internal/qbee"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type filedistributionFile struct {
	Command      types.String `tfsdk:"command"`
	PreCondition types.String `tfsdk:"pre_condition"`
	Templates    types.List   `tfsdk:"templates"`
	Parameters   types.List   `tfsdk:"parameters"`
}

func (f filedistributionFile) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"command":       types.StringType,
		"pre_condition": types.StringType,
		"templates": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: filedistributionTemplate{}.attrTypes(),
			},
		},
		"parameters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: filedistributionParameter{}.attrTypes(),
			},
		},
	}
}

type filedistributionParameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (f filedistributionParameter) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}
}

type filedistributionTemplate struct {
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	IsTemplate  types.Bool   `tfsdk:"is_template"`
}

func (f filedistributionTemplate) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source":      types.StringType,
		"destination": types.StringType,
		"is_template": types.BoolType,
	}
}

func planToQbeeFilesets(ctx context.Context, files []filedistributionFile) []qbee.FilesetConfig {
	var filesets []qbee.FilesetConfig

	for _, file := range files {
		var fsc qbee.FilesetConfig

		if !file.PreCondition.IsNull() {
			fsc.PreCondition = file.PreCondition.ValueString()
		}

		if !file.Command.IsNull() {
			fsc.Command = file.Command.ValueString()
		}

		var paramValues []filedistributionParameter
		file.Parameters.ElementsAs(ctx, &paramValues, false)
		var params []qbee.FilesetParameter
		for _, value := range paramValues {
			params = append(params, qbee.FilesetParameter{
				Key:   value.Key.ValueString(),
				Value: value.Value.ValueString(),
			})
		}

		var templateValues []filedistributionTemplate
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

func filedistributionToListValue(ctx context.Context, filedistribution *qbee.GetFileDistributionResponse, resp *resource.ReadResponse) basetypes.ListValue {
	var files []filedistributionFile

	for _, file := range filedistribution.Files {
		files = append(files, filedistributionFile{
			Command:      nullableStringValue(file.Command),
			PreCondition: nullableStringValue(file.PreCondition),
			Templates:    templatesToListValue(ctx, file.Templates, resp),
			Parameters:   parametersToListValue(ctx, file.Parameters, resp),
		})
	}

	fileValues, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: filedistributionFile{}.attrTypes(),
	}, files)
	resp.Diagnostics.Append(diags...)
	return fileValues
}

func nullableStringValue(value string) basetypes.StringValue {
	if value == "" {
		return types.StringNull()
	} else {
		return types.StringValue(value)
	}
}

func templatesToListValue(ctx context.Context, templates []qbee.FiledistributionFileTemplateResponse, resp *resource.ReadResponse) types.List {
	if templates == nil {
		return types.ListNull(basetypes.ObjectType{
			AttrTypes: filedistributionTemplate{}.attrTypes(),
		})
	}

	var result []filedistributionTemplate
	for _, template := range templates {
		result = append(result, filedistributionTemplate{
			Source:      types.StringValue(template.Source),
			Destination: types.StringValue(template.Destination),
			IsTemplate:  types.BoolValue(template.IsTemplate),
		})
	}

	templatesValue, diags := types.ListValueFrom(ctx, basetypes.ObjectType{AttrTypes: filedistributionTemplate{}.attrTypes()}, result)
	resp.Diagnostics.Append(diags...)
	return templatesValue
}

func parametersToListValue(ctx context.Context, parameters []qbee.FiledistributionFileParameterResponse, resp *resource.ReadResponse) basetypes.ListValue {
	if parameters == nil {
		return types.ListNull(basetypes.ObjectType{
			AttrTypes: filedistributionParameter{}.attrTypes(),
		})
	}

	var result []filedistributionParameter
	for _, parameter := range parameters {
		result = append(result, filedistributionParameter{
			Key:   types.StringValue(parameter.Key),
			Value: types.StringValue(parameter.Value),
		})
	}

	parametersValue, diags := types.ListValueFrom(ctx, basetypes.ObjectType{AttrTypes: filedistributionParameter{}.attrTypes()}, result)
	resp.Diagnostics.Append(diags...)
	return parametersValue
}
