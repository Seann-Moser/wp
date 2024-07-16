package source_code

import (
	"context"
	"net/http"
	"testing"
)

func TestDirect(t *testing.T) {
	ctx := context.Background()
	d := NewDirect(http.DefaultClient)
	d.Ping(ctx)
	_, _, err := d.Get(ctx, "https://idope.se/browse.html")
	if err == nil || err.Error() != HasChallengeErr.Error() {
		t.Fatalf("expected HasChallengeErr, got %v", err)
	}

	_, _, err = d.Get(ctx, "https://github.com/Seann-Moser/wp")
	if err != nil {
		t.Fatalf("expected HasChallengeErr, got %v", err)
	}

}
