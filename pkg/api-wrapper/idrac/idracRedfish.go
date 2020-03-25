package idrac

import (
	"context"
	"fmt"
	RFWrap "github.com/manojkva/go-redfish-API-Wrapper/pkg/api-wrapper/redfishapi-wrapper"
	redfish "opendev.org/airship/go-redfish/client"
)

type IdracRedfishClient struct {
	Username string
	Password string
	HostIP   string
	

}
func (a *IdracRedfishClient) UpgradeFirmware(filelocation string) {

	var auth = redfish.BasicAuth{UserName: a.Username,
		Password: a.Password,
	}
	ctx := context.WithValue(context.Background(), redfish.ContextBasicAuth, auth)

	httpPushURI := RFWrap.UpdateService(ctx)

	fmt.Printf("%v", httpPushURI)

	etag := RFWrap.GetETagHttpURI(ctx)
	fmt.Println("%v", etag)
	imageURI, _ := RFWrap.HTTPUriDownload(ctx, filelocation,etag)

	fmt.Printf("%v", imageURI)

	jobID := RFWrap.SimpleUpdateRequest(ctx, imageURI)

	fmt.Println("%v", jobID)

}
