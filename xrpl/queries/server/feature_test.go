package server

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestFeatureAllResponse(t *testing.T) {
	r := FeatureAllResponse{
		Features: map[string]types.FeatureStatus{
			"feature1": {Enabled: true, Name: "feature1", Supported: true},
			"feature2": {Enabled: false, Name: "feature2", Supported: false},
		},
	}

	s := `{
	"features": {
		"feature1": {
			"enabled": true,
			"name": "feature1",
			"supported": true
		},
		"feature2": {
			"enabled": false,
			"name": "feature2",
			"supported": false
		}
	}
}`

	if err := testutil.SerializeAndDeserialize(t, r, s); err != nil {
		t.Fatal(err)
	}
}

func TestFeatureOneRequest(t *testing.T) {
	r := FeatureOneRequest{
		Feature: "feature1",
	}

	s := `{
	"feature": "feature1"
}`

	if err := testutil.Serialize(t, r, s); err != nil {
		t.Fatal(err)
	}
}

func TestFeatureAllRequest_Validate(t *testing.T) {
	// FeatureAllRequest models xrpl.js FeatureAllRequest, which has no required
	// fields (feature is `feature?: never`). Only the valid path is exercised.
	tests := []struct {
		name    string
		request FeatureAllRequest
		wantErr error
	}{
		{
			name:    "pass - valid request",
			request: FeatureAllRequest{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeatureOneRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request FeatureOneRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: FeatureOneRequest{
				Feature: "foo",
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing feature",
			request: FeatureOneRequest{},
			wantErr: ErrNoFeature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeatureOneResponse(t *testing.T) {
	r := FeatureResponse{
		"feature1": {Enabled: true, Name: "feature1", Supported: true},
	}

	s := `{
	"feature1": {
		"enabled": true,
		"name": "feature1",
		"supported": true
	}
}`

	if err := testutil.SerializeAndDeserialize(t, r, s); err != nil {
		t.Fatal(err)
	}
}
