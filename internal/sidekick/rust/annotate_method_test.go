// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rust

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	libconfig "github.com/googleapis/librarian/internal/config"
	"github.com/googleapis/librarian/internal/sidekick/api"
)

func TestAnnotateMethodNames(t *testing.T) {
	model := annotateMethodModel(t)
	err := api.CrossReference(model)
	if err != nil {
		t.Fatal(err)
	}
	codec := newTestCodec(t, libconfig.SpecProtobuf, "", map[string]string{
		"include-grpc-only-methods": "true",
	})
	_, err = annotateModel(model, codec)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		MethodID string
		Want     *methodAnnotation
	}{
		{
			MethodID: ".test.v1.ResourceService.move",
			Want: &methodAnnotation{
				Name:                "r#move",
				NameNoMangling:      "move",
				BuilderName:         "Move",
				Body:                "None::<gaxi::http::NoBody>",
				ServiceNameToPascal: "ResourceService",
				ServiceNameToCamel:  "resourceService",
				ServiceNameToSnake:  "resource_service",
				ReturnType:          "crate::model::Response",
			},
		},
		{
			MethodID: ".test.v1.ResourceService.Delete",
			Want: &methodAnnotation{
				Name:                "delete",
				NameNoMangling:      "delete",
				BuilderName:         "Delete",
				Body:                "None::<gaxi::http::NoBody>",
				ServiceNameToPascal: "ResourceService",
				ServiceNameToCamel:  "resourceService",
				ServiceNameToSnake:  "resource_service",
				ReturnType:          "()",
			},
		},
		{
			MethodID: ".test.v1.ResourceService.Self",
			Want: &methodAnnotation{
				Name:                "r#self",
				NameNoMangling:      "self",
				BuilderName:         "r#Self",
				Body:                "None::<gaxi::http::NoBody>",
				ServiceNameToPascal: "ResourceService",
				ServiceNameToCamel:  "resourceService",
				ServiceNameToSnake:  "resource_service",
				ReturnType:          "crate::model::Response",
			},
		},
	} {
		gotMethod := model.Method(test.MethodID)
		if gotMethod == nil {
			t.Errorf("missing method %s", test.MethodID)
			continue
		}
		got := gotMethod.Codec.(*methodAnnotation)
		if diff := cmp.Diff(test.Want, got, cmpopts.IgnoreFields(methodAnnotation{}, "PathInfo", "SystemParameters")); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestAnnotateDiscoveryAnnotations(t *testing.T) {
	model := annotateMethodModel(t)
	err := api.CrossReference(model)
	if err != nil {
		t.Fatal(err)
	}
	codec := newTestCodec(t, libconfig.SpecProtobuf, "", map[string]string{
		"include-grpc-only-methods": "true",
	})
	_, err = annotateModel(model, codec)
	if err != nil {
		t.Fatal(err)
	}

	methodID := ".test.v1.ResourceService.Delete"
	gotMethod := model.Method(methodID)
	if gotMethod == nil {
		t.Fatalf("missing method %s", methodID)
	}
	got := gotMethod.DiscoveryLro.Codec.(*discoveryLroAnnotations)
	want := &discoveryLroAnnotations{
		MethodName:          "delete",
		ReturnType:          "()",
		ServiceNameToPascal: "ResourceService",
		PollingPathParameters: []discoveryLroPathParameter{
			{Name: "project", SetterName: "project"},
			{Name: "zone", SetterName: "zone"},
			{Name: "r#type", SetterName: "type"},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestAnnotateMethodAPIVersion(t *testing.T) {
	model := annotateMethodModel(t)
	err := api.CrossReference(model)
	if err != nil {
		t.Fatal(err)
	}

	// Inject an APIVersion to the existing model.
	methodID := ".test.v1.ResourceService.Delete"
	gotMethod := model.Method(methodID)
	if gotMethod == nil {
		t.Fatalf("missing method %s", methodID)
	}
	gotMethod.APIVersion = "v1_20260205"

	codec := newTestCodec(t, libconfig.SpecDiscovery, "", map[string]string{})
	_, err = annotateModel(model, codec)
	if err != nil {
		t.Fatal(err)
	}

	got := gotMethod.Codec.(*methodAnnotation)
	want := []systemParameter{
		{Name: "$alt", Value: "json"},
		{Name: "$apiVersion", Value: "v1_20260205"},
	}
	if diff := cmp.Diff(want, got.SystemParameters); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestAnnotateMethodInternalBuilders(t *testing.T) {
	model := annotateMethodModel(t)
	err := api.CrossReference(model)
	if err != nil {
		t.Fatal(err)
	}

	codec := newTestCodec(t, libconfig.SpecProtobuf, "", map[string]string{
		"internal-builders": "true",
	})
	_, err = annotateModel(model, codec)
	if err != nil {
		t.Fatal(err)
	}

	methodID := ".test.v1.ResourceService.Delete"
	gotMethod := model.Method(methodID)
	if gotMethod == nil {
		t.Fatalf("missing method %s", methodID)
	}
	got := gotMethod.Codec.(*methodAnnotation)
	if !got.InternalBuilders {
		t.Errorf("expected InternalBuilders to be true for method %s", methodID)
	}
	if got.BuilderVisibility() != "pub(crate)" {
		t.Errorf("mismatch in BuilderVisibility, want=pub(crate), got=%s", got.BuilderVisibility())
	}
}

func annotateMethodModel(t *testing.T) *api.API {
	t.Helper()
	request := &api.Message{
		Name:    "Request",
		Package: "test.v1",
		ID:      ".test.v1.Request",
		Fields: []*api.Field{
			{Name: "project", ID: ".test.v1.Request.project", Typez: api.TypezString},
			{Name: "zone", ID: ".test.v1.Request.zone", Typez: api.TypezString},
			{Name: "type", ID: ".test.v1.Request.type", Typez: api.TypezString},
			{Name: "name", ID: ".test.v1.Request.name", Typez: api.TypezString},
			{Name: "location", ID: ".test.v1.Request.location", Typez: api.TypezString},
			{Name: "cluster", ID: ".test.v1.Request.cluster", Typez: api.TypezString},
		},
	}
	response := &api.Message{
		Name:    "Response",
		Package: "test.v1",
		ID:      ".test.v1.Response",
	}
	methodMove := &api.Method{
		Name:         "move",
		ID:           ".test.v1.ResourceService.move",
		InputType:    request,
		InputTypeID:  ".test.v1.Request",
		OutputTypeID: ".test.v1.Response",
		PathInfo: &api.PathInfo{
			Bindings: []*api.PathBinding{
				{
					Verb:         "POST",
					PathTemplate: &api.PathTemplate{},
				},
			},
		},
	}
	methodDelete := &api.Method{
		Name:         "Delete",
		ID:           ".test.v1.ResourceService.Delete",
		InputType:    request,
		InputTypeID:  ".test.v1.Request",
		OutputTypeID: ".google.protobuf.Empty",
		ReturnsEmpty: true,
		PathInfo: &api.PathInfo{
			Bindings: []*api.PathBinding{
				{
					Verb: "DELETE",
					PathTemplate: (&api.PathTemplate{}).
						WithLiteral("projects").
						WithVariableNamed("project").
						WithLiteral("zones").
						WithVariableNamed("zone").
						// This is unlikely, but want to test variables that
						// are reserved words.
						WithLiteral("types").
						WithVariableNamed("type"),
				},
			},
		},
		DiscoveryLro: &api.DiscoveryLro{
			PollingPathParameters: []string{"project", "zone", "type"},
		},
	}
	methodSelf := &api.Method{
		Name:         "Self",
		ID:           ".test.v1.ResourceService.Self",
		InputType:    request,
		InputTypeID:  ".test.v1.Request",
		OutputTypeID: ".test.v1.Response",
		PathInfo: &api.PathInfo{
			Bindings: []*api.PathBinding{
				{
					Verb:         "GET",
					PathTemplate: &api.PathTemplate{},
				},
			},
		},
	}
	service := &api.Service{
		Name:    "ResourceService",
		ID:      ".test.v1.ResourceService",
		Package: "test.v1",
		Methods: []*api.Method{methodMove, methodDelete, methodSelf},
	}

	model := api.NewTestAPI(
		[]*api.Message{request, response},
		[]*api.Enum{},
		[]*api.Service{service})
	api.CrossReference(model)
	return model
}

func TestAnnotateMethodResourceNameTemplate(t *testing.T) {
	model := annotateMethodModel(t)
	err := api.CrossReference(model)
	if err != nil {
		t.Fatal(err)
	}

	// Helper to inject TargetResource
	injectTargetResource := func(methodID string, template string, fields [][]string) {
		m := model.Method(methodID)
		if m == nil {
			t.Fatalf("missing method %s", methodID)
		}
		if m.PathInfo != nil && len(m.PathInfo.Bindings) > 0 {
			m.PathInfo.Bindings[0].PathTemplate = (&api.PathTemplate{}).
				WithLiteral("projects").
				WithVariableNamed("project").
				WithLiteral("zones").
				WithVariableNamed("zone").
				WithLiteral("types").
				WithVariableNamed("type")
			m.PathInfo.Bindings[0].TargetResource = &api.TargetResource{
				Template:   api.ParseTemplateForTest(template),
				FieldPaths: fields,
			}
		}
	}

	// Setup: Inject resource for the "move" method
	injectTargetResource(".test.v1.ResourceService.move", "//Test.googleapis.com/projects/{project}/zones/{zone}/types/{type}", [][]string{
		{"project"},
		{"zone"},
		{"type"},
	})

	// Setup: Inject multiple bindings for the "Self" method
	mSelf := model.Method(".test.v1.ResourceService.Self")
	mSelf.PathInfo.Bindings = []*api.PathBinding{
		{
			Verb: "GET",
			PathTemplate: (&api.PathTemplate{}).
				WithLiteral("v1").
				WithVariableNamed("name"),
			TargetResource: &api.TargetResource{
				Template:   api.ParseTemplateForTest("//Test.googleapis.com/projects/{project}/locations/{location}/clusters/{cluster}"),
				FieldPaths: [][]string{{"name"}},
			},
		},
		{
			Verb: "GET",
			PathTemplate: (&api.PathTemplate{}).
				WithLiteral("v1").
				WithLiteral("projects").
				WithVariableNamed("project").
				WithLiteral("locations").
				WithVariableNamed("location").
				WithLiteral("clusters").
				WithVariableNamed("cluster"),
			TargetResource: &api.TargetResource{
				Template:   api.ParseTemplateForTest("//Test.googleapis.com/projects/{project}/locations/{location}/clusters/{cluster}"),
				FieldPaths: [][]string{{"project"}, {"location"}, {"cluster"}},
			},
		},
		{
			Verb:         "GET",
			PathTemplate: &api.PathTemplate{},
		},
	}

	codec := newTestCodec(t, libconfig.SpecProtobuf, "", map[string]string{
		"detailed-tracing-attributes": "true",
	})
	_, err = annotateModel(model, codec)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name         string
		id           string
		want         *methodAnnotation
		wantBindings []*pathBindingAnnotation
	}{
		{
			name: "WithTargetResource",
			id:   ".test.v1.ResourceService.move",
			want: &methodAnnotation{
				HasResourceNameGeneration: true,
			},
		},
		{
			name: "WithoutTargetResource",
			id:   ".test.v1.ResourceService.Delete",
			want: &methodAnnotation{
				HasResourceNameGeneration: false,
			},
		},
		{
			name: "MultipleBindings",
			id:   ".test.v1.ResourceService.Self",
			want: &methodAnnotation{
				HasResourceNameGeneration: true,
			},
			wantBindings: []*pathBindingAnnotation{
				{
					HasResourceNameGeneration: true,
					ResourceNameTemplate:      "//Test.googleapis.com/projects/{}/locations/{}/clusters/{}",
					ResourceNameArgs:          []string{"var_name"},
				},
				{
					HasResourceNameGeneration: true,
					ResourceNameTemplate:      "//Test.googleapis.com/projects/{}/locations/{}/clusters/{}",
					ResourceNameArgs:          []string{"var_project", "var_location", "var_cluster"},
				},
				{
					HasResourceNameGeneration: true,
					ResourceNameTemplate:      "",
					ResourceNameArgs:          nil,
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := model.Method(test.id)
			if m == nil {
				t.Fatalf("missing method %s", test.id)
			}
			got := m.Codec.(*methodAnnotation)
			if diff := cmp.Diff(test.want, got, cmpopts.IgnoreFields(methodAnnotation{},
				"Name", "NameNoMangling", "BuilderName", "Body", "DocLines",
				"ServiceNameToPascal", "ServiceNameToCamel", "ServiceNameToSnake",
				"SystemParameters", "ReturnType", "PathInfo", "Attributes",
				"RoutingRequired", "DetailedTracingAttributes",
				"ResourceNameTemplateGrpc", "GrpcResourceNameArgs",
				"InternalBuilders")); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}

			if test.wantBindings != nil {
				for i, wantBinding := range test.wantBindings {
					gotBinding := m.PathInfo.Bindings[i].Codec.(*pathBindingAnnotation)
					if diff := cmp.Diff(wantBinding, gotBinding, cmpopts.IgnoreFields(pathBindingAnnotation{}, "DetailedTracingAttributes", "PathFmt", "QueryParams", "Substitutions")); diff != "" {
						t.Errorf("mismatch (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}

func TestFormatResourceNameTemplateFromPath(t *testing.T) {
	for _, test := range []struct {
		name    string
		method  *api.Method
		binding *api.PathBinding
		want    string
		wantErr bool
	}{
		{
			name: "Basic",
			method: &api.Method{
				Model: &api.API{Name: "test"},
				Service: &api.Service{
					DefaultHost: "test.googleapis.com",
				},
			},
			binding: &api.PathBinding{
				TargetResource: &api.TargetResource{
					Template: api.ParseTemplateForTest("//test.googleapis.com/projects/{project}/zones/{zone}"),
				},
			},
			want: "//test.googleapis.com/projects/{}/zones/{}",
		},
		{
			name: "With Extended Field Path",
			method: &api.Method{
				Model:   &api.API{Name: "test"},
				Service: &api.Service{},
			},
			binding: &api.PathBinding{
				TargetResource: &api.TargetResource{
					Template: api.ParseTemplateForTest("//test.googleapis.com/items/{item.id}"),
				},
			},
			want: "//test.googleapis.com/items/{}",
		},
		{
			name: "Discovery API Compute V1 Example",
			method: &api.Method{
				Model: &api.API{Name: "compute"},
				Service: &api.Service{
					DefaultHost: "compute.googleapis.com",
				},
			},
			binding: &api.PathBinding{
				PathTemplate: (&api.PathTemplate{}).
					WithLiteral("compute").
					WithLiteral("v1").
					WithLiteral("projects").
					WithVariableNamed("project").
					WithLiteral("zones").
					WithVariableNamed("zone"),
				TargetResource: &api.TargetResource{
					// Notice that constructTemplate already stripped out "compute/v1"
					Template: api.ParseTemplateForTest("//compute.googleapis.com/projects/{project}/zones/{zone}"),
				},
			},
			want: "//compute.googleapis.com/projects/{}/zones/{}",
		},
		{
			name: "Missing TargetResource",
			method: &api.Method{
				ID:    "test.method",
				Model: &api.API{Name: "test"},
			},
			binding: &api.PathBinding{},
			wantErr: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := formatResourceNameTemplateFromPath(test.method, test.binding)
			if (err != nil) != test.wantErr {
				t.Fatalf("formatResourceNameTemplateFromPath() error = %v, wantErr %v", err, test.wantErr)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAnnotateSampleInfo(t *testing.T) {
	field := &api.Field{
		Name:  "name",
		ID:    ".test.v1.Message.name",
		Typez: api.TypezString,
		ResourceNamePattern: &api.ResourceNamePattern{
			Segments: []api.ResourceNameSegment{
				{Literal: "projects"},
				{Variable: "project"},
			},
		},
	}
	message := &api.Message{
		Name:    "TestMessage",
		Package: "test.v1",
		ID:      ".test.v1.TestMessage",
		Fields:  []*api.Field{field},
	}
	method := &api.Method{
		Name: "TestMethod",
		ID:   ".test.v1.Service.TestMethod",
	}
	si := &api.SampleInfo{
		ResourceNameField: field,
	}

	model := api.NewTestAPI([]*api.Message{message}, []*api.Enum{}, []*api.Service{})
	api.CrossReference(model)
	codec := newTestCodec(t, libconfig.SpecProtobuf, "test", map[string]string{})
	annotateModel(model, codec)

	codec.annotateSampleInfo(si, method)

	got := si.Codec.(*sampleInfoAnnotation)
	want := &sampleInfoAnnotation{
		StringParameters:   []string{"project_id"},
		FormatResourceName: true,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestMethodAnnotationsDetailedTracing(t *testing.T) {
	model := serviceAnnotationsModel()
	method := model.Method(".test.v1.ResourceService.GetResource")
	if method == nil {
		t.Fatal("cannot find .test.v1.ResourceService.GetResource")
	}
	codec := newTestCodec(t, libconfig.SpecProtobuf, "", map[string]string{
		"detailed-tracing-attributes": "true",
	})
	annotateModel(model, codec)
	got := method.Codec.(*methodAnnotation)
	if !got.DetailedTracingAttributes {
		t.Errorf("methodAnnotation.DetailedTracingAttributes = %v, want %v", got.DetailedTracingAttributes, true)
	}
}

func TestIsBidiStreaming(t *testing.T) {
	for _, test := range []struct {
		name   string
		client bool
		server bool
		want   bool
	}{
		{"bidi streaming", true, true, true},
		{"client-only streaming", true, false, false},
		{"server-only streaming", false, true, false},
		{"unary", false, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			ann := &methodAnnotation{ClientSideStreaming: test.client, ServerSideStreaming: test.server}
			if got := ann.IsBidiStreaming(); got != test.want {
				t.Errorf("methodAnnotation.IsBidiStreaming() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestIsServerStreaming(t *testing.T) {
	for _, test := range []struct {
		name   string
		client bool
		server bool
		want   bool
	}{
		{"bidi streaming", true, true, false},
		{"client-only streaming", true, false, false},
		{"server-only streaming", false, true, true},
		{"unary", false, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			ann := &methodAnnotation{ClientSideStreaming: test.client, ServerSideStreaming: test.server}
			if got := ann.IsServerStreaming(); got != test.want {
				t.Errorf("methodAnnotation.IsServerStreaming() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestAnnotateMethodPreconditions(t *testing.T) {
	reqWithOptionalMatch := &api.Message{
		Name:    "DeleteObjectRequest",
		Package: "google.storage.v2",
		ID:      ".google.storage.v2.DeleteObjectRequest",
		Fields: []*api.Field{
			{Name: "bucket", ID: ".google.storage.v2.DeleteObjectRequest.bucket", Typez: api.TypezString},
			{Name: "object", ID: ".google.storage.v2.DeleteObjectRequest.object", Typez: api.TypezString},
			{Name: "if_generation_match", ID: ".google.storage.v2.DeleteObjectRequest.if_generation_match", Typez: api.TypezInt64, Optional: true},
			{Name: "if_metageneration_match", ID: ".google.storage.v2.DeleteObjectRequest.if_metageneration_match", Typez: api.TypezInt64, Optional: true},
		},
	}
	reqWithScalarMatch := &api.Message{
		Name:    "LockBucketRetentionPolicyRequest",
		Package: "google.storage.v2",
		ID:      ".google.storage.v2.LockBucketRetentionPolicyRequest",
		Fields: []*api.Field{
			{Name: "bucket", ID: ".google.storage.v2.LockBucketRetentionPolicyRequest.bucket", Typez: api.TypezString},
			{Name: "if_metageneration_match", ID: ".google.storage.v2.LockBucketRetentionPolicyRequest.if_metageneration_match", Typez: api.TypezInt64, Optional: false},
		},
	}
	reqWithoutMatch := &api.Message{
		Name:    "GetObjectRequest",
		Package: "google.storage.v2",
		ID:      ".google.storage.v2.GetObjectRequest",
		Fields: []*api.Field{
			{Name: "bucket", ID: ".google.storage.v2.GetObjectRequest.bucket", Typez: api.TypezString},
			{Name: "object", ID: ".google.storage.v2.GetObjectRequest.object", Typez: api.TypezString},
		},
	}

	methodWithOptional := &api.Method{
		Name:         "DeleteObject",
		ID:           ".google.storage.v2.Storage.DeleteObject",
		InputType:    reqWithOptionalMatch,
		InputTypeID:  ".google.storage.v2.DeleteObjectRequest",
		OutputTypeID: ".google.protobuf.Empty",
		ReturnsEmpty: true,
		PathInfo:     &api.PathInfo{},
	}
	methodWithScalar := &api.Method{
		Name:         "LockBucketRetentionPolicy",
		ID:           ".google.storage.v2.Storage.LockBucketRetentionPolicy",
		InputType:    reqWithScalarMatch,
		InputTypeID:  ".google.storage.v2.LockBucketRetentionPolicyRequest",
		OutputTypeID: ".google.protobuf.Empty",
		ReturnsEmpty: true,
		PathInfo:     &api.PathInfo{},
	}
	methodWithout := &api.Method{
		Name:         "GetObject",
		ID:           ".google.storage.v2.Storage.GetObject",
		InputType:    reqWithoutMatch,
		InputTypeID:  ".google.storage.v2.GetObjectRequest",
		OutputTypeID: ".google.protobuf.Empty",
		ReturnsEmpty: true,
		PathInfo:     &api.PathInfo{},
	}

	service := &api.Service{
		Name:    "Storage",
		Package: "google.storage.v2",
		ID:      ".google.storage.v2.Storage",
		Methods: []*api.Method{methodWithOptional, methodWithScalar, methodWithout},
	}

	model := api.NewTestAPI([]*api.Message{reqWithOptionalMatch, reqWithScalarMatch, reqWithoutMatch}, []*api.Enum{}, []*api.Service{service})
	if err := api.CrossReference(model); err != nil {
		t.Fatal(err)
	}
	codec := newTestCodec(t, libconfig.SpecProtobuf, "storage", map[string]string{
		"include-grpc-only-methods": "true",
	})
	if _, err := annotateModel(model, codec); err != nil {
		t.Fatal(err)
	}

	// 1. Verify Optional match fields
	gotOpt := methodWithOptional.Codec.(*methodAnnotation)
	if !gotOpt.HasPreconditions {
		t.Errorf("methodWithOptional.HasPreconditions = false, want true")
	}
	wantOptExpr := "req.if_generation_match.is_some() ||\n            req.if_metageneration_match.is_some()"
	if gotOpt.PreconditionExpr != wantOptExpr {
		t.Errorf("methodWithOptional.PreconditionExpr mismatch:\ngot:  %q\nwant: %q", gotOpt.PreconditionExpr, wantOptExpr)
	}

	// 2. Verify Scalar match field
	gotScalar := methodWithScalar.Codec.(*methodAnnotation)
	if !gotScalar.HasPreconditions {
		t.Errorf("methodWithScalar.HasPreconditions = false, want true")
	}
	wantScalarExpr := "req.if_metageneration_match != 0"
	if gotScalar.PreconditionExpr != wantScalarExpr {
		t.Errorf("methodWithScalar.PreconditionExpr mismatch:\ngot:  %q\nwant: %q", gotScalar.PreconditionExpr, wantScalarExpr)
	}

	// 3. Verify request without match fields
	gotWithout := methodWithout.Codec.(*methodAnnotation)
	if gotWithout.HasPreconditions {
		t.Errorf("methodWithout.HasPreconditions = true, want false")
	}
	if gotWithout.PreconditionExpr != "" {
		t.Errorf("methodWithout.PreconditionExpr = %q, want empty", gotWithout.PreconditionExpr)
	}
}

