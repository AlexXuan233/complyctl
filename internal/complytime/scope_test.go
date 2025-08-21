// SPDX-License-Identifier: Apache-2.0
package complytime

import (
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/stretchr/testify/require"
)

func TestNewAssessmentScopeFromCDs(t *testing.T) {
	testAppDir := ApplicationDirectory{}
	validator := validation.NoopValidator{}

	_, err := NewAssessmentScopeFromCDs("example", testAppDir, validator)
	require.EqualError(t, err, "no component definitions found")

	cd := oscalTypes.ComponentDefinition{
		Components: &[]oscalTypes.DefinedComponent{
			{
				Title: "Component",
				Props: &[]oscalTypes.Property{
					{
						Name:    extensions.RuleIdProp,
						Value:   "rule-1",
						Remarks: "remarks-group-1",
					},
					{
						Name:    extensions.ParameterIdProp,
						Value:   "param-1",
						Remarks: "remarks-group-1",
					},
					{
						Name:    extensions.RuleIdProp,
						Value:   "rule-2",
						Remarks: "remarks-group-2",
					},
					{
						Name:    extensions.ParameterIdProp,
						Value:   "param-2",
						Remarks: "remarks-group-2",
					},
				},
				ControlImplementations: &[]oscalTypes.ControlImplementationSet{
					{
						Props: &[]oscalTypes.Property{
							{
								Name:  extensions.FrameworkProp,
								Value: "example",
								Ns:    extensions.TrestleNameSpace,
							},
						},
						SetParameters: &[]oscalTypes.SetParameter{
							{
								ParamId: "param-1",
								Values:  []string{"value-1"},
							},
							{
								ParamId: "param-2",
								Values:  []string{"value-2"},
							},
						},
						ImplementedRequirements: []oscalTypes.ImplementedRequirementControlImplementation{
							{
								ControlId: "control-1",
								Props: &[]oscalTypes.Property{
									{
										Name:  extensions.RuleIdProp,
										Value: "rule-1",
									},
								},
							},
							{
								ControlId: "control-2",
								Props: &[]oscalTypes.Property{
									{
										Name:  extensions.RuleIdProp,
										Value: "rule-2",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	wantScope := AssessmentScope{
		FrameworkID: "example",
		IncludeControls: []ControlEntry{
			{
				ControlID:    "control-1",
				ControlTitle: "",
				IncludeRules: []string{"*"},
				SelectParameters: []ParameterEntry{
					{Name: "param-1", Value: "value-1"},
				},
			},
			{
				ControlID:    "control-2",
				ControlTitle: "",
				IncludeRules: []string{"*"},
				SelectParameters: []ParameterEntry{
					{Name: "param-2", Value: "value-2"},
				},
			},
		},
	}
	scope, err := NewAssessmentScopeFromCDs("example", testAppDir, validator, cd)
	require.NoError(t, err)

	// Check the basic structure
	require.Equal(t, wantScope.FrameworkID, scope.FrameworkID)
	require.Len(t, scope.IncludeControls, len(wantScope.IncludeControls))

	// Check each control entry, allowing for different parameter orders
	for i, wantControl := range wantScope.IncludeControls {
		actualControl := scope.IncludeControls[i]
		require.Equal(t, wantControl.ControlID, actualControl.ControlID)
		require.Equal(t, wantControl.ControlTitle, actualControl.ControlTitle)
		require.Equal(t, wantControl.IncludeRules, actualControl.IncludeRules)

		// Check parameters exist regardless of order
		require.Len(t, actualControl.SelectParameters, len(wantControl.SelectParameters))
		for _, wantParam := range wantControl.SelectParameters {
			found := false
			for _, actualParam := range actualControl.SelectParameters {
				if actualParam.Name == wantParam.Name && actualParam.Value == wantParam.Value {
					found = true
					break
				}
			}
			require.True(t, found, "Expected parameter %s=%s not found", wantParam.Name, wantParam.Value)
		}
	}

	// Reproduce duplicates
	anotherComponent := oscalTypes.DefinedComponent{
		Title: "AnotherComponent",
		ControlImplementations: &[]oscalTypes.ControlImplementationSet{
			{
				Props: &[]oscalTypes.Property{
					{
						Name:  extensions.FrameworkProp,
						Value: "example",
						Ns:    extensions.TrestleNameSpace,
					},
				},
				ImplementedRequirements: []oscalTypes.ImplementedRequirementControlImplementation{
					{
						ControlId: "control-1",
						Props: &[]oscalTypes.Property{
							{
								Name:  extensions.RuleIdProp,
								Value: "rule-1",
							},
						},
					},
					{
						ControlId: "control-2",
						Props: &[]oscalTypes.Property{
							{
								Name:  extensions.RuleIdProp,
								Value: "rule-2",
							},
						},
					},
				},
			},
		},
	}
	*cd.Components = append(*cd.Components, anotherComponent)

	scope, err = NewAssessmentScopeFromCDs("example", testAppDir, validator, cd)
	require.NoError(t, err)

	// Check the basic structure again after adding duplicates
	require.Equal(t, wantScope.FrameworkID, scope.FrameworkID)
	require.Len(t, scope.IncludeControls, len(wantScope.IncludeControls))

	// Check each control entry again, allowing for different parameter orders
	for i, wantControl := range wantScope.IncludeControls {
		actualControl := scope.IncludeControls[i]
		require.Equal(t, wantControl.ControlID, actualControl.ControlID)
		require.Equal(t, wantControl.ControlTitle, actualControl.ControlTitle)
		require.Equal(t, wantControl.IncludeRules, actualControl.IncludeRules)

		// Check parameters exist regardless of order
		require.Len(t, actualControl.SelectParameters, len(wantControl.SelectParameters))
		for _, wantParam := range wantControl.SelectParameters {
			found := false
			for _, actualParam := range actualControl.SelectParameters {
				if actualParam.Name == wantParam.Name && actualParam.Value == wantParam.Value {
					found = true
					break
				}
			}
			require.True(t, found, "Expected parameter %s=%s not found", wantParam.Name, wantParam.Value)
		}
	}
}

func TestNewAssessmentScopeFromCDs_ParameterRuleMatching(t *testing.T) {
	testAppDir := ApplicationDirectory{}
	validator := validation.NoopValidator{}

	cd := oscalTypes.ComponentDefinition{
		Components: &[]oscalTypes.DefinedComponent{
			{
				Title: "Component",
				Props: &[]oscalTypes.Property{
					// Remarks group 1: rule-1 with param-1 and param-2
					{
						Name:    extensions.RuleIdProp,
						Value:   "rule-1",
						Remarks: "remarks-group-1",
					},
					{
						Name:    extensions.ParameterIdProp + "_1",
						Value:   "param-1",
						Remarks: "remarks-group-1",
					},
					{
						Name:    extensions.ParameterIdProp + "_2",
						Value:   "param-2",
						Remarks: "remarks-group-1",
					},
					// Remarks group 2: rule-2 with param-3
					{
						Name:    extensions.RuleIdProp,
						Value:   "rule-2",
						Remarks: "remarks-group-2",
					},
					{
						Name:    extensions.ParameterIdProp,
						Value:   "param-3",
						Remarks: "remarks-group-2",
					},
					// Remarks group 3: rule-3 with param-4 (not used by any control)
					{
						Name:    extensions.RuleIdProp,
						Value:   "rule-3",
						Remarks: "remarks-group-3",
					},
					{
						Name:    extensions.ParameterIdProp,
						Value:   "param-4",
						Remarks: "remarks-group-3",
					},
				},
				ControlImplementations: &[]oscalTypes.ControlImplementationSet{
					{
						Props: &[]oscalTypes.Property{
							{
								Name:  extensions.FrameworkProp,
								Value: "example",
								Ns:    extensions.TrestleNameSpace,
							},
						},
						SetParameters: &[]oscalTypes.SetParameter{
							{
								ParamId: "param-1",
								Values:  []string{"value-1"},
							},
							{
								ParamId: "param-2",
								Values:  []string{"value-2"},
							},
							{
								ParamId: "param-3",
								Values:  []string{"value-3"},
							},
							{
								ParamId: "param-4",
								Values:  []string{"value-4"},
							},
						},
						ImplementedRequirements: []oscalTypes.ImplementedRequirementControlImplementation{
							{
								ControlId: "control-1",
								Props: &[]oscalTypes.Property{
									{
										Name:  extensions.RuleIdProp,
										Value: "rule-1",
									},
								},
							},
							{
								ControlId: "control-2",
								Props: &[]oscalTypes.Property{
									{
										Name:  extensions.RuleIdProp,
										Value: "rule-2",
									},
								},
							},
							{
								ControlId: "control-3",
								// No rules - should have default N/A parameter
							},
						},
					},
				},
			},
		},
	}

	scope, err := NewAssessmentScopeFromCDs("example", testAppDir, validator, cd)
	require.NoError(t, err)

	// Check that control-1 gets param-1 and param-2 (from rule-1's remarks group)
	require.Len(t, scope.IncludeControls, 3)

	control1 := scope.IncludeControls[0]
	require.Equal(t, "control-1", control1.ControlID)
	require.Len(t, control1.SelectParameters, 2)

	paramNames := make([]string, len(control1.SelectParameters))
	for i, param := range control1.SelectParameters {
		paramNames[i] = param.Name
	}
	require.Contains(t, paramNames, "param-1")
	require.Contains(t, paramNames, "param-2")

	// Check that control-2 gets param-3 (from rule-2's remarks group)
	control2 := scope.IncludeControls[1]
	require.Equal(t, "control-2", control2.ControlID)
	require.Len(t, control2.SelectParameters, 1)
	require.Equal(t, "param-3", control2.SelectParameters[0].Name)
	require.Equal(t, "value-3", control2.SelectParameters[0].Value)

	// Check that control-3 gets default N/A parameter (no rules)
	control3 := scope.IncludeControls[2]
	require.Equal(t, "control-3", control3.ControlID)
	require.Len(t, control3.SelectParameters, 1)
	require.Equal(t, "N/A", control3.SelectParameters[0].Name)
	require.Equal(t, "N/A", control3.SelectParameters[0].Value)
}

func TestNewAssessmentScopeFromCDs_NoParameters(t *testing.T) {
	testAppDir := ApplicationDirectory{}
	validator := validation.NoopValidator{}

	cd := oscalTypes.ComponentDefinition{
		Components: &[]oscalTypes.DefinedComponent{
			{
				Title: "Component",
				ControlImplementations: &[]oscalTypes.ControlImplementationSet{
					{
						Props: &[]oscalTypes.Property{
							{
								Name:  extensions.FrameworkProp,
								Value: "example",
								Ns:    extensions.TrestleNameSpace,
							},
						},
						ImplementedRequirements: []oscalTypes.ImplementedRequirementControlImplementation{
							{
								ControlId: "control-1",
								// No SetParameters
							},
						},
					},
				},
			},
		},
	}

	wantScope := AssessmentScope{
		FrameworkID: "example",
		IncludeControls: []ControlEntry{
			{
				ControlID:    "control-1",
				ControlTitle: "",
				IncludeRules: []string{"*"},
				SelectParameters: []ParameterEntry{
					{Name: "N/A", Value: "N/A"},
				},
			},
		},
	}

	scope, err := NewAssessmentScopeFromCDs("example", testAppDir, validator, cd)
	require.NoError(t, err)
	require.Equal(t, wantScope, scope)
}

func TestAssessmentScope_ApplyScope(t *testing.T) {
	testLogger := hclog.NewNullLogger()

	tests := []struct {
		name           string
		basePlan       *oscalTypes.AssessmentPlan
		scope          AssessmentScope
		wantSelections []oscalTypes.AssessedControls
	}{
		{
			name: "Success/Default",
			basePlan: &oscalTypes.AssessmentPlan{
				ReviewedControls: oscalTypes.ReviewedControls{
					ControlSelections: []oscalTypes.AssessedControls{
						{
							IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
								{
									ControlId: "example-1",
								},
								{
									ControlId: "example-2",
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				FrameworkID: "test",
				IncludeControls: []ControlEntry{
					{ControlID: "example-2", IncludeRules: []string{"*"}},
				},
			},
			wantSelections: []oscalTypes.AssessedControls{
				{
					IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
						{
							ControlId: "example-2",
						},
					},
				},
			},
		},
		// Testing for out-of-scope controls
		{
			name: "All Controls Out-of-Scope",
			basePlan: &oscalTypes.AssessmentPlan{
				ReviewedControls: oscalTypes.ReviewedControls{
					ControlSelections: []oscalTypes.AssessedControls{
						{
							IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
								{
									ControlId: "",
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				FrameworkID:     "test",
				IncludeControls: nil,
			},
			wantSelections: []oscalTypes.AssessedControls{
				{
					IncludeControls: nil,
				},
			},
		},
		{
			name: "Some Controls Out-of-Scope",
			basePlan: &oscalTypes.AssessmentPlan{
				ReviewedControls: oscalTypes.ReviewedControls{
					ControlSelections: []oscalTypes.AssessedControls{
						{
							IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
								{
									ControlId: "example-1",
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				FrameworkID: "test",
				IncludeControls: []ControlEntry{
					{ControlID: "example-1", IncludeRules: []string{"*"}},
					{ControlID: "example-2", IncludeRules: []string{"*"}},
				},
			},
			wantSelections: []oscalTypes.AssessedControls{
				{
					IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
						{
							ControlId: "example-1",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := tt.scope
			err := scope.ApplyScope(tt.basePlan, testLogger)
			require.NoError(t, err)
			require.Equal(t, tt.wantSelections, tt.basePlan.ReviewedControls.ControlSelections)
		})
	}
}

func TestAssessmentScope_ApplyRuleScope(t *testing.T) {
	testLogger := hclog.NewNullLogger()

	tests := []struct {
		name           string
		basePlan       *oscalTypes.AssessmentPlan
		scope          AssessmentScope
		wantActivities *[]oscalTypes.Activity
	}{
		{
			name: "Success/Default",
			basePlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "rule-1",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "example-1",
											},
											{
												ControlId: "example-2",
											},
										},
									},
								},
							},
						},
						{
							Title: "rule-2",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "example-1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				FrameworkID: "test",
				IncludeControls: []ControlEntry{
					{ControlID: "example-1", IncludeRules: []string{"*"}},
					{ControlID: "example-2", IncludeRules: []string{"*"}},
				},
			},
			wantActivities: &[]oscalTypes.Activity{
				{
					Title: "rule-1",
					RelatedControls: &oscalTypes.ReviewedControls{
						ControlSelections: []oscalTypes.AssessedControls{
							{
								IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
									{
										ControlId: "example-1",
									},
									{
										ControlId: "example-2",
									},
								},
							},
						},
					},
				},
				{
					Title: "rule-2",
					RelatedControls: &oscalTypes.ReviewedControls{
						ControlSelections: []oscalTypes.AssessedControls{
							{
								IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
									{
										ControlId: "example-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Success/ExcludeRuleForControl",
			basePlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "rule-1",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-1",
											},
											{
												ControlId: "control-2",
											},
										},
									},
								},
							},
						},
						{
							Title: "rule-2",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				FrameworkID: "test",
				IncludeControls: []ControlEntry{
					{ControlID: "control-1", IncludeRules: []string{"*"}, ExcludeRules: []string{"rule-1"}},
					{ControlID: "control-2", IncludeRules: []string{"*"}},
				},
			},
			wantActivities: &[]oscalTypes.Activity{
				{
					Title: "rule-1",
					RelatedControls: &oscalTypes.ReviewedControls{
						ControlSelections: []oscalTypes.AssessedControls{
							{
								IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
									{
										ControlId: "control-2",
									},
								},
							},
						},
					},
				},
				{
					Title: "rule-2",
					RelatedControls: &oscalTypes.ReviewedControls{
						ControlSelections: []oscalTypes.AssessedControls{
							{
								IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
									{
										ControlId: "control-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Success/ActivityMarkedSkipped",
			basePlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "rule-1",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-1",
											},
										},
									},
								},
							},
						},
						{
							Title: "rule-2",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				FrameworkID: "test",
				IncludeControls: []ControlEntry{
					{ControlID: "control-1", IncludeRules: []string{"*"}, ExcludeRules: []string{"rule-1"}},
					{ControlID: "control-2", IncludeRules: []string{"*"}},
				},
			},
			wantActivities: &[]oscalTypes.Activity{
				{
					Title: "rule-1",
					Props: &[]oscalTypes.Property{
						{
							Name:  "skipped",
							Value: "true",
							Ns:    extensions.TrestleNameSpace,
						},
					},
					RelatedControls: nil,
				},
				{
					Title: "rule-2",
					RelatedControls: &oscalTypes.ReviewedControls{
						ControlSelections: []oscalTypes.AssessedControls{
							{
								IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
									{
										ControlId: "control-2",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Success/MissingIncludeRules",
			basePlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "rule-1",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-1",
											},
										},
									},
								},
							},
						},
						{
							Title: "rule-2",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				FrameworkID: "test",
				IncludeControls: []ControlEntry{
					{ControlID: "control-1", ExcludeRules: []string{"rule-1"}}, // Missing includeRules should default to "*"
				},
			},
			wantActivities: &[]oscalTypes.Activity{
				{
					Title: "rule-1",
					Props: &[]oscalTypes.Property{
						{
							Name:  "skipped",
							Value: "true",
							Ns:    extensions.TrestleNameSpace,
						},
					},
					RelatedControls: nil,
				},
				{
					Title: "rule-2",
					RelatedControls: &oscalTypes.ReviewedControls{
						ControlSelections: []oscalTypes.AssessedControls{
							{
								IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
									{
										ControlId: "control-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Success/ExcludeAllRules",
			basePlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "rule-1",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-1",
											},
										},
									},
								},
							},
						},
						{
							Title: "rule-2",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				FrameworkID: "test",
				IncludeControls: []ControlEntry{
					{ControlID: "control-1", IncludeRules: []string{"rule-1", "rule-2"}, ExcludeRules: []string{"*"}}, // ExcludeRules="*" should override includeRules
				},
			},
			wantActivities: &[]oscalTypes.Activity{
				{
					Title: "rule-1",
					Props: &[]oscalTypes.Property{
						{
							Name:  "skipped",
							Value: "true",
							Ns:    extensions.TrestleNameSpace,
						},
					},
					RelatedControls: nil,
				},
				{
					Title: "rule-2",
					Props: &[]oscalTypes.Property{
						{
							Name:  "skipped",
							Value: "true",
							Ns:    extensions.TrestleNameSpace,
						},
					},
					RelatedControls: nil,
				},
			},
		},
		{
			name: "Success/GlobalExcludeOverridesInclude",
			basePlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "rule-1",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-1",
											},
										},
									},
								},
							},
						},
						{
							Title: "rule-2",
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{
												ControlId: "control-1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				FrameworkID:        "test",
				GlobalExcludeRules: []string{"rule-1"}, // Global exclude should override control-specific include
				IncludeControls: []ControlEntry{
					{ControlID: "control-1", IncludeRules: []string{"rule-1", "rule-2"}}, // Explicitly includes rule-1, but global exclude wins
				},
			},
			wantActivities: &[]oscalTypes.Activity{
				{
					Title: "rule-1",
					Props: &[]oscalTypes.Property{
						{
							Name:  "skipped",
							Value: "true",
							Ns:    extensions.TrestleNameSpace,
						},
					},
					RelatedControls: nil,
				},
				{
					Title: "rule-2",
					RelatedControls: &oscalTypes.ReviewedControls{
						ControlSelections: []oscalTypes.AssessedControls{
							{
								IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
									{
										ControlId: "control-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := tt.scope
			err := scope.ApplyScope(tt.basePlan, testLogger)
			require.NoError(t, err)
			require.Equal(t, tt.wantActivities, tt.basePlan.LocalDefinitions.Activities)
		})
	}
}

func TestAssessmentScope_ApplyParameterScope(t *testing.T) {
	testLogger := hclog.NewNullLogger()

	tests := []struct {
		name           string
		assessmentPlan *oscalTypes.AssessmentPlan
		componentDefs  []oscalTypes.ComponentDefinition
		scope          AssessmentScope
		expectError    bool
		expectedProps  []oscalTypes.Property
	}{
		{
			name: "Success/ParameterUpdate",
			assessmentPlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "test-activity",
							Props: &[]oscalTypes.Property{
								{
									Name:  "param-1",
									Value: "old-value",
									Class: extensions.TestParameterClass,
								},
							},
							RelatedControls: &oscalTypes.ReviewedControls{
								ControlSelections: []oscalTypes.AssessedControls{
									{
										IncludeControls: &[]oscalTypes.AssessedControlsSelectControlById{
											{ControlId: "control-1"},
										},
									},
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				IncludeControls: []ControlEntry{
					{
						ControlID: "control-1",
						SelectParameters: []ParameterEntry{
							{Name: "param-1", Value: "new-value"},
						},
					},
				},
			},
			expectError: false,
			expectedProps: []oscalTypes.Property{
				{
					Name:  "param-1",
					Value: "new-value",
					Class: extensions.TestParameterClass,
				},
			},
		},
		{
			name: "Success/NoParametersToUpdate",
			assessmentPlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "test-activity",
							Props: &[]oscalTypes.Property{
								{
									Name:  "param-1",
									Value: "old-value",
									Class: extensions.TestParameterClass,
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				IncludeControls: []ControlEntry{},
			},
			expectError: false,
			expectedProps: []oscalTypes.Property{
				{
					Name:  "param-1",
					Value: "old-value",
					Class: extensions.TestParameterClass,
				},
			},
		},
		{
			name: "Success/EmptyParameterName",
			assessmentPlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "test-activity",
							Props: &[]oscalTypes.Property{
								{
									Name:  "param-1",
									Value: "old-value",
									Class: extensions.TestParameterClass,
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				IncludeControls: []ControlEntry{
					{
						ControlID: "control-1",
						SelectParameters: []ParameterEntry{
							{Name: "", Value: "new-value"}, // Empty name should be ignored
						},
					},
				},
			},
			expectError: false,
			expectedProps: []oscalTypes.Property{
				{
					Name:  "param-1",
					Value: "old-value",
					Class: extensions.TestParameterClass,
				},
			},
		},
		{
			name: "Success/NoSelectParameters",
			assessmentPlan: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "test-activity",
							Props: &[]oscalTypes.Property{
								{
									Name:  "param-1",
									Value: "old-value",
									Class: extensions.TestParameterClass,
								},
							},
						},
					},
				},
			},
			scope: AssessmentScope{
				IncludeControls: []ControlEntry{
					{
						ControlID: "control-1",
						// No SelectParameters
					},
				},
			},
			expectError: false,
			expectedProps: []oscalTypes.Property{
				{
					Name:  "param-1",
					Value: "old-value",
					Class: extensions.TestParameterClass,
				},
			},
		},
		{
			name:           "Success/NoLocalDefinitions",
			assessmentPlan: &oscalTypes.AssessmentPlan{
				// No LocalDefinitions
			},
			scope: AssessmentScope{
				IncludeControls: []ControlEntry{
					{
						ControlID: "control-1",
						SelectParameters: []ParameterEntry{
							{Name: "param-1", Value: "new-value"},
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := tt.scope
			err := scope.applyParameterScope(tt.assessmentPlan, tt.componentDefs, testLogger)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.expectedProps != nil && tt.assessmentPlan.LocalDefinitions != nil &&
					tt.assessmentPlan.LocalDefinitions.Activities != nil {
					activity := (*tt.assessmentPlan.LocalDefinitions.Activities)[0]
					if activity.Props != nil {
						require.Equal(t, tt.expectedProps, *activity.Props)
					}
				}
			}
		})
	}
}

func TestValidateParameterValue(t *testing.T) {
	componentDefs := []oscalTypes.ComponentDefinition{
		{
			Components: &[]oscalTypes.DefinedComponent{
				{
					Props: &[]oscalTypes.Property{
						{
							Name:    extensions.ParameterIdProp,
							Value:   "test-param",
							Remarks: "group1",
						},
						{
							Name:    "Parameter_Value_Alternatives",
							Value:   `{"option1": "description1", "option2": "description2"}`,
							Remarks: "group1",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name          string
		parameterID   string
		selectedValue string
		componentDefs []oscalTypes.ComponentDefinition
		expectError   bool
	}{
		{
			name:          "Error/InvalidParameterValue",
			parameterID:   "test-param",
			selectedValue: "invalid-option",
			componentDefs: componentDefs,
			expectError:   true,
		},
		{
			name:          "Success/ValidParameterValue",
			parameterID:   "test-param",
			selectedValue: "option1",
			componentDefs: componentDefs,
			expectError:   false,
		},
		{
			name:          "Success/NoAlternativesAcceptAnyValue",
			parameterID:   "unknown-param",
			selectedValue: "any-value",
			componentDefs: componentDefs,
			expectError:   false,
		},
		{
			name:          "Success/EmptyValueAccepted",
			parameterID:   "test-param",
			selectedValue: "",
			componentDefs: componentDefs,
			expectError:   false,
		},
		{
			name:          "Success/NAValueAccepted",
			parameterID:   "test-param",
			selectedValue: "N/A",
			componentDefs: componentDefs,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParameterValue(tt.parameterID, tt.selectedValue, tt.componentDefs)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFilterParameterSelection(t *testing.T) {
	remarksProps := map[string][]oscalTypes.Property{
		"group1": {
			{
				Name:  extensions.ParameterIdProp,
				Value: "param1",
			},
			{
				Name:  "Parameter_Value_Alternatives",
				Value: `{"valid1": "desc1", "valid2": "desc2"}`,
			},
		},
		"group2": {
			{
				Name:  extensions.ParameterIdProp + "_1",
				Value: "param2",
			},
			{
				Name:  "Parameter_Value_Alternatives_1",
				Value: `{"validA": "descA", "validB": "descB"}`,
			},
		},
	}

	tests := []struct {
		name                 string
		parameterID          string
		selectedValue        string
		remarksProps         map[string][]oscalTypes.Property
		expectedValid        bool
		expectedAlternatives []string
	}{
		{
			name:                 "Success/ValidValueInAlternatives",
			parameterID:          "param1",
			selectedValue:        "valid1",
			remarksProps:         remarksProps,
			expectedValid:        true,
			expectedAlternatives: []string{"valid1", "valid2"},
		},
		{
			name:                 "Error/InvalidValueNotInAlternatives",
			parameterID:          "param1",
			selectedValue:        "invalid",
			remarksProps:         remarksProps,
			expectedValid:        false,
			expectedAlternatives: []string{"valid1", "valid2"},
		},
		{
			name:                 "Success/IndexedParameterValidAlternatives",
			parameterID:          "param2",
			selectedValue:        "validA",
			remarksProps:         remarksProps,
			expectedValid:        true,
			expectedAlternatives: []string{"validA", "validB"},
		},
		{
			name:                 "Success/EmptyValueAccepted",
			parameterID:          "param1",
			selectedValue:        "",
			remarksProps:         remarksProps,
			expectedValid:        true,
			expectedAlternatives: nil,
		},
		{
			name:                 "Success/NAValueAccepted",
			parameterID:          "param1",
			selectedValue:        "N/A",
			remarksProps:         remarksProps,
			expectedValid:        true,
			expectedAlternatives: nil,
		},
		{
			name:                 "Success/NoAlternativesAcceptAnyValue",
			parameterID:          "unknown",
			selectedValue:        "any-value",
			remarksProps:         remarksProps,
			expectedValid:        true,
			expectedAlternatives: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, alternatives := filterParameterSelection(tt.parameterID, tt.selectedValue, tt.remarksProps)
			require.Equal(t, tt.expectedValid, valid)
			if tt.expectedAlternatives != nil {
				require.ElementsMatch(t, tt.expectedAlternatives, alternatives)
			} else {
				require.Nil(t, alternatives)
			}
		})
	}
}
