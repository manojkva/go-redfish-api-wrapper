package redfishapi-wrapper

import (

       redfish "opendev.org/airship/go-redfish/client"
       "fmt"
       "context"
       "net/http"
       "crypto/tls"
       "time"
       "encoding/json"
//     "reflect"
        "os"
        "github.com/antihax/optional"
//      "net/url"
        "regexp"
 //     "io/ioutil"
 logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

)

var log = logf.Log.WithName("RedfishAPI")
type RedfishAPIWrapper interface {
     UpgradeFirmware( string )  (error)

}

func prettyPrint(i interface{}) string {
    s, _ := json.MarshalIndent(i, "", "\t")
    return string(s)
}
var tr *http.Transport = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    30 * time.Second,
	DisableCompression: true,
        TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
        }

func createAPIClient(HeaderInfo map[string]string) *redfish.DefaultApiService {
        client := &http.Client{Transport: tr}
        cfg := &redfish.Configuration{
                BasePath:      "https://32.68.250.78",
                DefaultHeader: make(map[string]string),
                UserAgent:     "go-redfish/client",
                HTTPClient: client,
        }

        if len(HeaderInfo) != 0 {
        
        for key,value := range HeaderInfo {
                cfg.DefaultHeader[key] = value
        }
        } 
        return redfish.NewAPIClient(cfg).DefaultApi
}

func GetTask( ctx context.Context, taskID string ) {
       redfishApi := createAPIClient(make(map[string]string)) 
       sl, response,err := redfishApi.GetTask(ctx,taskID )
       fmt.Printf( "%+v %+v %+v", prettyPrint(sl),response, err)
}

func GetVirtualMedia( ctx context.Context ) {
       redfishApi := createAPIClient(make(map[string]string)) 
       sl, response,err := redfishApi.GetManagerVirtualMedia(ctx,"iDRAC.Embedded.1","CD" )
       fmt.Printf( "%+v %+v %+v", prettyPrint(sl),response, err)
}

func UpdateService( ctx context.Context ) string  {
       redfishApi := createAPIClient(make(map[string]string)) 
       // call the UpdateService and get the HttpPushURi
       sl, response,err := redfishApi.UpdateService(ctx)
       fmt.Printf( "%+v %+v %+v", prettyPrint(sl),response, err)
       return sl.HttpPushUri
}

func HTTPUriDownload( ctx context.Context, filePath string , etag string) (string,error ) {
        filehandle, err  := os.Open(filePath)
        if err != nil {
            fmt.Println(err)
           }
        defer filehandle.Close()
        reqBody :=  redfish.FirmwareInventoryDownloadImageOpts{  SoftwareImage :  optional.NewInterface(filehandle) }
        headerInfo := make(map[string]string)
        headerInfo["if-match"] = etag
       redfishApi := createAPIClient(headerInfo) 

	sl,response,err := redfishApi.FirmwareInventoryDownloadImage(ctx,&reqBody )
        fmt.Printf( "%+v %+v %+v", prettyPrint(sl),response, err)
        location, _ := response.Location()
        return string(location.RequestURI()), err

}


func  GetETagHttpURI ( ctx context.Context ) string {
       redfishApi := createAPIClient(make(map[string]string)) 
       sl, response, err := redfishApi.FirmwareInventory(ctx)
       fmt.Printf( "%+v %+v %+v", prettyPrint(sl),response, err)
       etag :=  response.Header["Etag"]
       fmt.Printf("%v", etag[0])
       return etag[0]
}

func SimpleUpdateRequest( ctx context.Context, imageURI string) string {
        headerInfo := make(map[string]string)
       redfishApi := createAPIClient(headerInfo) 
        reqBody := new (redfish.SimpleUpdateRequestBody) 
        localUriImage := imageURI
        reqBody.ImageURI = localUriImage
	    sl,response,err := redfishApi.UpdateServiceSimpleUpdate(ctx, *reqBody,)
        fmt.Printf( "%+v %+v %+v", prettyPrint(sl),response, err)
        jobID_location := response.Header["Location"]
        re := regexp.MustCompile(`JID_(.*)`)
        jobID :=  re.FindStringSubmatch(jobID_location[0])[1]
        return jobID
}

