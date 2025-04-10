package add_cloud_metadata

import (
	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

var ucloudMetadataFetcher = provider{
	Name: "ucloud",

	DefaultEnabled: false,
	Create: func(_ string, c *conf.C) (metadataFetcher, error) {
		ucloudMetadataHost := "100.80.80.80"
		ucloudMetadataInstanceIDURI := "/meta-data/latest/instance-id"
		ucloudMetadataRegionURI := "/meta-data/latest/region"
		ucloudMetadataZoneURI := "/meta-data/latest/availability-zone"
		ucloudSchema := func(m map[string]interface{}) mapstr.M {
			m["service"] = mapstr.M{
				"name": "UHost",
			}

			return mapstr.M{"cloud": m}
		}
		urls, err := getMetadataURLs(c, ucloudMetadataHost, []string{
			ucloudMetadataInstanceIDURI,
			ucloudMetadataRegionURI,
			ucloudMetadataZoneURI,
		})
		if err != nil {
			return nil, err
		}

		responseHandlers := map[string]responseHandler{
			urls[0]: func(all []byte, result *result) error {
				result.metadata.Put("instance.id", string(all))
				return nil
			},
			urls[1]: func(all []byte, result *result) error {
				result.metadata["region"] = string(all)
				return nil
			},
			urls[2]: func(all []byte, result *result) error {
				result.metadata["availability_zone"] = string(all)
				return nil
			},
		}

		fetcher := &httpMetadataFetcher{"ucloud", nil, responseHandlers, ucloudSchema}
		return fetcher, err
	},
}
