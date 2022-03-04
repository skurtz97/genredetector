package auth

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *Auth
	}{
		{name: "startup", want: &Auth{
			Id:           os.Getenv("SPOTIFY_CLIENT_ID"),
			Secret:       os.Getenv("SPOTIFY_CLIENT_SECRET"),
			AccessToken:  "",
			AuthorizedAt: time.Time{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
