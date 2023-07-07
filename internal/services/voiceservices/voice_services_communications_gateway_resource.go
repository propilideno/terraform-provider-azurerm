// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package voiceservices

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/go-azure-sdk/resource-manager/voiceservices/2023-01-31/communicationsgateways"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type CommunicationsGatewayModel struct {
	Name                               string                                                   `tfschema:"name"`
	ResourceGroupName                  string                                                   `tfschema:"resource_group_name"`
	ApiBridge                          string                                                   `tfschema:"api_bridge"`
	AutoGeneratedDomainNameLabelScope  communicationsgateways.AutoGeneratedDomainNameLabelScope `tfschema:"auto_generated_domain_name_label_scope"`
	Codecs                             string                                                   `tfschema:"codecs"`
	Connectivity                       string                                                   `tfschema:"connectivity"`
	E911Type                           communicationsgateways.E911Type                          `tfschema:"e911_type"`
	EmergencyDialStrings               []string                                                 `tfschema:"emergency_dial_strings"`
	Location                           string                                                   `tfschema:"location"`
	OnPremMcpEnabled                   bool                                                     `tfschema:"on_prem_mcp_enabled"`
	Platforms                          []string                                                 `tfschema:"platforms"`
	ServiceLocation                    []ServiceRegionPropertiesModel                           `tfschema:"service_location"`
	Tags                               map[string]string                                        `tfschema:"tags"`
	MicrosoftTeamsVoicemailPilotNumber string                                                   `tfschema:"microsoft_teams_voicemail_pilot_number"`
}

type ServiceRegionPropertiesModel struct {
	Location                              string   `tfschema:"location"`
	OperatorAddresses                     []string `tfschema:"operator_addresses"`
	AllowedMediaSourceAddressPrefixes     []string `tfschema:"allowed_media_source_address_prefixes"`
	AllowedSignalingSourceAddressPrefixes []string `tfschema:"allowed_signaling_source_address_prefixes"`
	EsrpAddresses                         []string `tfschema:"esrp_addresses"`
}

type PrimaryRegionPropertiesModel struct {
}

type CommunicationsGatewayResource struct{}

var _ sdk.ResourceWithUpdate = CommunicationsGatewayResource{}

var _ sdk.ResourceWithCustomizeDiff = CommunicationsGatewayResource{}

func (r CommunicationsGatewayResource) ResourceType() string {
	return "azurerm_voice_services_communications_gateway"
}

func (r CommunicationsGatewayResource) ModelObject() interface{} {
	return &CommunicationsGatewayModel{}
}

func (r CommunicationsGatewayResource) CustomizeDiff() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			var model CommunicationsGatewayModel
			if err := metadata.DecodeDiff(&model); err != nil {
				return fmt.Errorf("DecodeDiff: %+v", err)
			}

			for _, sls := range model.ServiceLocation {
				if model.E911Type == communicationsgateways.E911TypeStandard {
					if len(sls.EsrpAddresses) > 0 {
						return fmt.Errorf("the esrp_addresses of %s must not be provided for each service_location when e911_type is set to Standard", model.Name)
					}
				} else {
					if len(sls.EsrpAddresses) == 0 {
						return fmt.Errorf("the esrp_addresses of %s must be provided for each service_location when e911_type is set to DirectToEsrp", model.Name)
					}
				}
			}
			return nil
		},
		Timeout: 30 * time.Minute,
	}
}

func (r CommunicationsGatewayResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return communicationsgateways.ValidateCommunicationsGatewayID
}

func (r CommunicationsGatewayResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
			ValidateFunc: validation.StringMatch(
				regexp.MustCompile("^[a-zA-Z0-9-]{3,24}$"),
				"The name can only contain letters, numbers and dashes, the name length must be from 3 to 24 characters.",
			),
		},

		"location": commonschema.Location(),

		"resource_group_name": commonschema.ResourceGroupName(),

		"connectivity": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
			ValidateFunc: validation.StringInSlice([]string{
				string(communicationsgateways.ConnectivityPublicAddress),
			}, false),
		},

		"codecs": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				string(communicationsgateways.TeamsCodecsPCMA),
				string(communicationsgateways.TeamsCodecsPCMU),
				string(communicationsgateways.TeamsCodecsGSevenTwoTwo),
				string(communicationsgateways.TeamsCodecsGSevenTwoTwoTwo),
				string(communicationsgateways.TeamsCodecsSILKEight),
				string(communicationsgateways.TeamsCodecsSILKOneSix),
			}, false),
		},

		"e911_type": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				string(communicationsgateways.E911TypeDirectToEsrp),
				string(communicationsgateways.E911TypeStandard),
			}, false),
		},

		"platforms": {
			Type:     pluginsdk.TypeList,
			Required: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					string(communicationsgateways.CommunicationsPlatformOperatorConnect),
					string(communicationsgateways.CommunicationsPlatformTeamsPhoneMobile),
				}, false),
			},
		},

		"service_location": {
			Type:     pluginsdk.TypeSet,
			Required: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"location": commonschema.LocationWithoutForceNew(),

					"operator_addresses": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
					"allowed_media_source_address_prefixes": {
						Type:     pluginsdk.TypeSet,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},

					"allowed_signaling_source_address_prefixes": {
						Type:     pluginsdk.TypeSet,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},

					"esrp_addresses": {
						Type:     pluginsdk.TypeSet,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
				},
			},
		},

		"auto_generated_domain_name_label_scope": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			ForceNew: true,
			Default:  string(communicationsgateways.AutoGeneratedDomainNameLabelScopeTenantReuse),
			ValidateFunc: validation.StringInSlice([]string{
				string(communicationsgateways.AutoGeneratedDomainNameLabelScopeTenantReuse),
				string(communicationsgateways.AutoGeneratedDomainNameLabelScopeSubscriptionReuse),
				string(communicationsgateways.AutoGeneratedDomainNameLabelScopeResourceGroupReuse),
				string(communicationsgateways.AutoGeneratedDomainNameLabelScopeNoReuse),
			}, false),
		},

		"api_bridge": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsJSON,
		},

		"emergency_dial_strings": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"on_prem_mcp_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"tags": commonschema.Tags(),

		"microsoft_teams_voicemail_pilot_number": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
	}
}

func (r CommunicationsGatewayResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r CommunicationsGatewayResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			var model CommunicationsGatewayModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			client := metadata.Client.VoiceServices.CommunicationsGatewaysClient
			subscriptionId := metadata.Client.Account.SubscriptionId
			id := communicationsgateways.NewCommunicationsGatewayID(subscriptionId, model.ResourceGroupName, model.Name)

			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}

			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			properties := &communicationsgateways.CommunicationsGateway{
				Location: location.Normalize(model.Location),
				Properties: &communicationsgateways.CommunicationsGatewayProperties{
					AutoGeneratedDomainNameLabelScope: &model.AutoGeneratedDomainNameLabelScope,
					Connectivity:                      communicationsgateways.Connectivity(model.Connectivity),
					Codecs: []communicationsgateways.TeamsCodecs{
						communicationsgateways.TeamsCodecs(model.Codecs),
					},
					E911Type:         model.E911Type,
					Platforms:        expandCommunicationsPlatformModel(model.Platforms),
					ServiceLocations: expandServiceRegionPropertiesModel(model.ServiceLocation),
				},
				Tags: &model.Tags,
			}

			var apiBridgeValue interface{}
			if model.ApiBridge != "" {
				log.Printf("[DEBUG] unmarshalling json for ApiBridge")
				if err = json.Unmarshal([]byte(model.ApiBridge), &apiBridgeValue); err != nil {
					return fmt.Errorf("unmarshalling value for ApiBridge: %+v", err)
				}
			}
			properties.Properties.ApiBridge = &apiBridgeValue

			if model.EmergencyDialStrings != nil {
				properties.Properties.EmergencyDialStrings = &model.EmergencyDialStrings
			}

			properties.Properties.OnPremMcpEnabled = &model.OnPremMcpEnabled

			properties.Properties.TeamsVoicemailPilotNumber = &model.MicrosoftTeamsVoicemailPilotNumber

			if err := client.CreateOrUpdateThenPoll(ctx, id, *properties); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)

			return nil
		},
	}
}

func (r CommunicationsGatewayResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.VoiceServices.CommunicationsGatewaysClient

			id, err := communicationsgateways.ParseCommunicationsGatewayID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var model CommunicationsGatewayModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			resp, err := client.Get(ctx, *id)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}

			properties := resp.Model
			if properties == nil {
				return fmt.Errorf("retrieving %s: model was nil", id)
			}

			if metadata.ResourceData.HasChange("codecs") {
				properties.Properties.Codecs = []communicationsgateways.TeamsCodecs{
					communicationsgateways.TeamsCodecs(model.Codecs),
				}
			}

			if metadata.ResourceData.HasChange("e911_type") {
				properties.Properties.E911Type = model.E911Type
			}

			if metadata.ResourceData.HasChange("platforms") {
				properties.Properties.Platforms = expandCommunicationsPlatformModel(model.Platforms)
			}

			if metadata.ResourceData.HasChange("service_location") {
				properties.Properties.ServiceLocations = expandServiceRegionPropertiesModel(model.ServiceLocation)
			}

			if metadata.ResourceData.HasChange("api_bridge") {
				if model.ApiBridge != "" {
					var apiBridgeValue interface{}
					log.Printf("[DEBUG] unmarshalling json for ApiBridge")
					err = json.Unmarshal([]byte(model.ApiBridge), &apiBridgeValue)
					if err != nil {
						return fmt.Errorf("unmarshalling json value for ApiBridge: %+v", err)
					}
					properties.Properties.ApiBridge = &apiBridgeValue
				} else {
					properties.Properties.ApiBridge = nil
				}
			}

			if metadata.ResourceData.HasChange("emergency_dial_strings") {
				properties.Properties.EmergencyDialStrings = &model.EmergencyDialStrings
			}

			if metadata.ResourceData.HasChange("on_prem_mcp_enabled") {
				properties.Properties.OnPremMcpEnabled = &model.OnPremMcpEnabled
			}

			if metadata.ResourceData.HasChange("tags") {
				properties.Tags = &model.Tags
			}

			if metadata.ResourceData.HasChange("microsoft_teams_voicemail_pilot_number") {
				properties.Properties.TeamsVoicemailPilotNumber = &model.MicrosoftTeamsVoicemailPilotNumber
			}

			if err := client.CreateOrUpdateThenPoll(ctx, *id, *properties); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (r CommunicationsGatewayResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.VoiceServices.CommunicationsGatewaysClient

			id, err := communicationsgateways.ParseCommunicationsGatewayID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			resp, err := client.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}

				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}

			model := resp.Model
			if model == nil {
				return fmt.Errorf("retrieving %s: model was nil", id)
			}

			state := CommunicationsGatewayModel{
				Name:              id.CommunicationsGatewayName,
				ResourceGroupName: id.ResourceGroupName,
				Location:          location.Normalize(model.Location),
			}

			if properties := model.Properties; properties != nil {
				state.Connectivity = string(properties.Connectivity)

				codecsValue := ""
				if properties.Codecs != nil && len(properties.Codecs) > 0 {
					codecsValue = string(properties.Codecs[0])
				}
				state.Codecs = codecsValue

				state.E911Type = properties.E911Type

				state.Platforms = flattenCommunicationsPlatformModel(properties.Platforms)

				state.ServiceLocation = flattenServiceRegionPropertiesModel(&properties.ServiceLocations)

				if properties.AutoGeneratedDomainNameLabelScope != nil {
					state.AutoGeneratedDomainNameLabelScope = *properties.AutoGeneratedDomainNameLabelScope
				}

				if properties.ApiBridge != nil && *properties.ApiBridge != nil {
					apiBridgeValue, err := json.Marshal(*properties.ApiBridge)
					if err != nil {
						return fmt.Errorf("marshalling value for ApiBridge: %+v", err)
					}
					state.ApiBridge = string(apiBridgeValue)
				}

				if properties.EmergencyDialStrings != nil {
					state.EmergencyDialStrings = *properties.EmergencyDialStrings
				}

				onPremMcpEnabled := false
				if properties.OnPremMcpEnabled != nil {
					onPremMcpEnabled = *properties.OnPremMcpEnabled
				}
				state.OnPremMcpEnabled = onPremMcpEnabled

				v := ""
				if properties.TeamsVoicemailPilotNumber != nil {
					v = *properties.TeamsVoicemailPilotNumber
				}
				state.MicrosoftTeamsVoicemailPilotNumber = v
			}

			if model.Tags != nil {
				state.Tags = *model.Tags
			}

			return metadata.Encode(&state)
		},
	}
}

func (r CommunicationsGatewayResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.VoiceServices.CommunicationsGatewaysClient

			id, err := communicationsgateways.ParseCommunicationsGatewayID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			if err := client.DeleteThenPoll(ctx, *id); err != nil {
				return fmt.Errorf("deleting %s: %+v", id, err)
			}

			return nil
		},
	}
}

func expandServiceRegionPropertiesModel(inputList []ServiceRegionPropertiesModel) []communicationsgateways.ServiceRegionProperties {
	outputList := make([]communicationsgateways.ServiceRegionProperties, 0)
	for _, v := range inputList {
		output := communicationsgateways.ServiceRegionProperties{
			Name: location.Normalize(v.Location),
		}

		output.PrimaryRegionProperties = communicationsgateways.PrimaryRegionProperties{
			AllowedMediaSourceAddressPrefixes:     utils.StringSlice(v.AllowedMediaSourceAddressPrefixes),
			AllowedSignalingSourceAddressPrefixes: utils.StringSlice(v.AllowedSignalingSourceAddressPrefixes),
			EsrpAddresses:                         utils.StringSlice(v.EsrpAddresses),
			OperatorAddresses:                     v.OperatorAddresses,
		}

		outputList = append(outputList, output)
	}

	return outputList
}

func expandCommunicationsPlatformModel(input []string) []communicationsgateways.CommunicationsPlatform {
	if len(input) == 0 {
		return nil
	}

	var output []communicationsgateways.CommunicationsPlatform
	for _, v := range input {
		platform := communicationsgateways.CommunicationsPlatform(v)
		output = append(output, platform)
	}

	return output
}

func flattenServiceRegionPropertiesModel(inputList *[]communicationsgateways.ServiceRegionProperties) []ServiceRegionPropertiesModel {
	outputList := make([]ServiceRegionPropertiesModel, 0)
	if inputList == nil {
		return outputList
	}

	for _, input := range *inputList {
		output := ServiceRegionPropertiesModel{
			Location: location.Normalize(input.Name),
		}

		v := &input.PrimaryRegionProperties
		if v.OperatorAddresses != nil {
			output.OperatorAddresses = v.OperatorAddresses
		}

		if v.AllowedMediaSourceAddressPrefixes != nil {
			output.AllowedMediaSourceAddressPrefixes = *v.AllowedMediaSourceAddressPrefixes
		}

		if v.AllowedSignalingSourceAddressPrefixes != nil {
			output.AllowedSignalingSourceAddressPrefixes = *v.AllowedSignalingSourceAddressPrefixes
		}

		if v.EsrpAddresses != nil {
			output.EsrpAddresses = *v.EsrpAddresses
		}

		outputList = append(outputList, output)
	}

	return outputList
}

func flattenCommunicationsPlatformModel(input []communicationsgateways.CommunicationsPlatform) []string {
	output := make([]string, 0)
	if len(input) == 0 {
		return nil
	}

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}
