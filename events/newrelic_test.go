package events

import (
	"fmt"
	"testing"
)

func TestFilterAppMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    `appMessageFilterSuccess`,
			message: `*errors.errorString: Forbidden: {"timestamp":"2020-11-04T22:47:56.580+00:00","status":403,"error":"Forbidden","message":"Access denied to application this-is-my-app-name - required authorization: READ"}`,
			want:    `*errors.errorString: Forbidden: {"timestamp":"2020-11-04T22:47:56.580+00:00","status":403,"error":"Forbidden","message":"Access denied to application obfuscated_app_name- required authorization: READ"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterAppMessage([]byte(tt.message)); string(got) != tt.want {
				t.Errorf("FilterAppMessage() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestFilterErrorMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    `TestFilterErrorMessageSuccess`,
			message: `*errors.errorString: Forbidden: {"timestamp":"2020-11-04T22:47:56.580+00:00","status":403,"error":"Forbidden","message":"Access denied to application this-is-my-app-name - required authorization: READ"}`,
			want:    `*errors.errorString: Forbidden: {"timestamp":"2020-11-04T22:47:56.580+00:00","status":403,"error":"Forbidden","message":"Access denied to application obfuscated_app_name- required authorization: READ"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterErrorMessage(fmt.Errorf("%v", tt.message)); string(got) != tt.want {
				t.Errorf("FilterErrorMessage() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestFilterServiceUrlMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    `localhostExampleSuccess`,
			message: `could not create application - Post http://localhost:8083/ops: dial tcp [::1]:8083: connect: connection refused`,
			want:    `could not create application - Post http://obfuscated_url:8083/ops: dial tcp [::1]:8083: connect: connection refused`,
		},
		{
			name:    `spinEchoExampleSuccess`,
			message: `could not create application - Post http://spin-echo:8083/ops: dial tcp [::1]:8083: connect: connection refused`,
			want:    `could not create application - Post http://obfuscated_url:8083/ops: dial tcp [::1]:8083: connect: connection refused`,
		},
		{
			name:    `localhostHttpsExampleSuccess`,
			message: `could not create application - Post https://localhost:8083/ops: dial tcp [::1]:8083: connect: connection refused`,
			want:    `could not create application - Post http://obfuscated_url:8083/ops: dial tcp [::1]:8083: connect: connection refused`,
		},
		{
			name:    `spinEchoHttpsExampleSuccess`,
			message: `could not create application - Post https://spin-echo:8083/ops: dial tcp [::1]:8083: connect: connection refused`,
			want:    `could not create application - Post http://obfuscated_url:8083/ops: dial tcp [::1]:8083: connect: connection refused`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterServiceUrlMessage([]byte(tt.message)); string(got) != tt.want {
				t.Errorf("FilterServiceUrlMessage() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
