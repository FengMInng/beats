package add_cloud_metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/elastic-agent-libs/logp"
)

func initUCloudTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/meta-data/latest/cloud-name" {
			_, _ = w.Write([]byte("ucloud"))
			return
		}
		if r.RequestURI == "/meta-data/latest/instance-id" {
			_, _ = w.Write([]byte("uhost-12345"))
			return
		}
		if r.RequestURI == "/meta-data/latest/region" {
			_, _ = w.Write([]byte("cn-bj2"))
			return
		}
		if r.RequestURI == "/meta-data/latest/availability-zone" {
			_, _ = w.Write([]byte("cn-bj2-01"))
			return
		}
	}))
}

func TestRetrieveUCloudMetadata(t *testing.T) {
	logp.TestingSetup()
	server := initUCloudTestServer()
	defer server.Close()
	config, err := conf.NewConfigFrom(map[string]interface{}{
		"providers": []string{"ucloud"},
		"host":      server.Listener.Addr().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	p, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := p.Run(&beat.Event{Fields: mapstr.M{}})
	if err != nil {
		t.Fatal(err)
	}

	expected := mapstr.M{
		"cloud": mapstr.M{
			"provider": "ucloud",
			"instance": mapstr.M{
				"id": "uhost-12345",
			},
			"region":            "cn-bj2",
			"availability_zone": "cn-bj2-01",
			"service": mapstr.M{
				"name": "UHost",
			},
		},
	}
	assert.Equal(t, expected, actual.Fields)
}
