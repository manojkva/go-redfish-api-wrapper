package idrac

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	RFWrap "github.com/manojkva/go-redfish-API-Wrapper/pkg/redfishwrap"
	redfish "opendev.org/airship/go-redfish/client"
)

type IdracRedfishClient struct {
	Username  string
	Password  string
	HostIP    string
	IDRAC_ver string
}

func (a *IdracRedfishClient) createContext() context.Context {

	var auth = redfish.BasicAuth{UserName: a.Username,
		Password: a.Password,
	}
	ctx := context.WithValue(context.Background(), redfish.ContextBasicAuth, auth)
	return ctx
}

func (a *IdracRedfishClient) UpgradeFirmware(filelocation string) {

	ctx := a.createContext()

	httpPushURI := RFWrap.UpdateService(ctx, a.HostIP)

	fmt.Printf("%v", httpPushURI)

	etag := RFWrap.GetETagHttpURI(ctx, a.HostIP)
	fmt.Printf("%v", etag)
	imageURI, _ := RFWrap.HTTPUriDownload(ctx, a.HostIP, filelocation, etag)

	fmt.Printf("%v", imageURI)

	jobID := RFWrap.SimpleUpdateRequest(ctx, a.HostIP, imageURI)

	fmt.Printf("%v", jobID)

	a.CheckJobStatus(jobID)
}

func (a *IdracRedfishClient) CheckJobStatus(jobId string) {
	ctx := a.createContext()
	start := time.Now()

	for {

		statusCode, jobInfo := RFWrap.GetTask(ctx, a.HostIP, jobId)

		timeelapsedInMinutes := time.Since(start).Minutes()

		if (statusCode == 202) || (statusCode == 200) {
			fmt.Printf("HTTP  status OK")

		} else {
			fmt.Printf("Failed to check the status")
			os.Exit(3)
		}

		if timeelapsedInMinutes >= 60 {
			fmt.Println("\n- FAIL: Timeout of 1 hour has been hit, update job should of already been marked completed. Check the iDRAC job queue and LC logs to debug the issue")
			os.Exit(3)
		} else if strings.Contains(jobInfo.Messages[0].Message, "failed") {
			fmt.Println("FAIL")
			os.Exit(3)

		} else if strings.Contains(jobInfo.Messages[0].Message, "scheduled") {
			//	fmt.Prinln("\n- PASS, job ID %s successfully marked as scheduled, powering on or rebooting the server to apply the update" % data[u"Id"] ")
			break
		} else if strings.Contains(jobInfo.Messages[0].Message, "completed successfully") {
			//		fmt.Prinln("\n- PASS, job ID %s successfully marked as scheduled, powering on or rebooting the server to apply the update" % data[u"Id"] ")
			fmt.Println("Success")
			break
		} else {
			time.Sleep(1)
			continue
		}
	}
}

func (a *IdracRedfishClient) RebootServer(systemID string) bool {

	ctx := a.createContext()

	//Systems/System.Embedded.1/Actions/ComputerSystem.Reset

	return RFWrap.ResetServer(ctx, a.HostIP, systemID)

}

func (a *IdracRedfishClient) PowerOn(systemID string) bool {
	ctx := a.createContext()
	computeSystem := redfish.ComputerSystem{PowerState: redfish.POWERSTATE_ON}

	return RFWrap.SetSystem(ctx, a.HostIP, systemID, computeSystem)
}

func (a *IdracRedfishClient) PowerOff(systemID string) bool {
	ctx := a.createContext()

	computeSystem := redfish.ComputerSystem{PowerState: redfish.POWERSTATE_OFF}

	return RFWrap.SetSystem(ctx, a.HostIP, systemID, computeSystem)
}

func (a *IdracRedfishClient) GetVirtualMediaStatus(managerID string, media string) bool {
	ctx := a.createContext()
	return RFWrap.GetVirtualMediaConnectedStatus(ctx, a.HostIP, managerID, media)
}

func (a *IdracRedfishClient) EjectISO(managerID string, media string) bool {
	ctx := a.createContext()
	return RFWrap.EjectVirtualMedia(ctx, a.HostIP, managerID, media)
}

func (a *IdracRedfishClient) SetOneTimeBoot(systemID string) bool {
	ctx := a.createContext()
	computeSystem := redfish.ComputerSystem{Boot: redfish.Boot{BootSourceOverrideEnabled: "Once"}}

	return RFWrap.SetSystem(ctx, a.HostIP, systemID, computeSystem)

}

func (a *IdracRedfishClient) InsertISO(managerID string, mediaID string, imageURL string) bool {

	ctx := a.createContext()

	if a.GetVirtualMediaStatus(managerID, mediaID) {
		fmt.Println("Exiting .. Already connected")
		return false
	}
	insertMediaReqBody := redfish.InsertMediaRequestBody{
		Image: imageURL,
	}
	return RFWrap.InsertVirtualMedia(ctx, a.HostIP, managerID, mediaID, insertMediaReqBody)

}

func (a *IdracRedfishClient) GetVirtualDisks(systemID string, controllerID string)[]string {

	ctx := a.createContext()
	idrefs := RFWrap.GetVolumes(ctx, a.HostIP, systemID, controllerID)
	if idrefs == nil{
		return nil
	}
	virtualDisks := []string{}
	for _,id := range idrefs{

		fmt.Printf("VirtualDisk Info %v\n", id.OdataId)
		vd := strings.Split(id.OdataId,"/")
		if vd != nil {
		  virtualDisks = append(virtualDisks,vd[len(vd)-1])
		}
	}
	return virtualDisks

}

func (a *IdracRedfishClient) DeletVirtualDisk(systemID string, storageID string) string {
	ctx := a.createContext()

	return RFWrap.DeleteVirtualDisk(ctx, a.HostIP, systemID, storageID)
}

func (a *IdracRedfishClient) CreateVirtualDisk(systemID string, controllerID string, volumeType string, name string, urilist []string) string {
	ctx := a.createContext()

	drives := []redfish.IdRef{}

	for _, uri := range urilist {
		driveinfo := fmt.Sprintf("/redfish/v1/Systems/%s/Storage/Drives/%s",systemID, uri)
		drives = append(drives, redfish.IdRef{OdataId: driveinfo})
	}

	createvirtualBodyReq := redfish.CreateVirtualDiskRequestBody{
		VolumeType: redfish.VolumeType(volumeType),
		Name:       name,
		Drives:     drives,
	}

	return RFWrap.CreateVirtualDisk(ctx, a.HostIP, systemID, controllerID, createvirtualBodyReq)
}

func (a *IdracRedfishClient)CleanVirtualDisksIfAny(systemID string, controllerID string) bool{

	var result bool = false

	// Get the list of VirtualDisks
	virtualDisks := a.GetVirtualDisks(systemID, controllerID)
	// for testing skip the OS Disk
	//virtualDisks = virtualDisks[1:] 
	if len(virtualDisks) == 0 {
		fmt.Printf("No existing RAIS found")
		
	} else {
		for _,vd  := range virtualDisks {
			jobid  := a.DeletVirtualDisk(systemID,vd)
			fmt.Printf("Delete Job ID %v\n",jobid)
		//	a.CheckJobStatus(jobid)
			result = true

			if result == false {
				fmt.Printf("Failed to delete virtual disk %v\n",vd)
				return result
         
			}
		}
	}

    return result
}

func (a * IdracRedfishClient)GetNodeUUID(systemID string )(string, bool){

	ctx := a.createContext()
	computerSystem, _  := RFWrap.GetSystem(ctx,a.HostIP,systemID)

	if computerSystem != nil{
		return computerSystem.UUID, true
	}
	return "", false
}

//TODO
// Add function to clean up RAID
// Add function to create new RAID

/*
func (hp hardwareProfile) cleanVirtualDIskIfEExists() bool {
		// url := "https://32.67.151.80/redfish/v1/Systems/System.Embedded.1/Storage/Volumes"
		var result bool = false
		endpoint := hp.UrlMappings.SystemURL("get_virtual_disks")
    header := hp.RedfishClient.Header
    r, err := hp.RedfishClient.HttpClient.Get(endpoint, header)
    resp := rp.CheckErrorAndReturn(r,err)
    var data map[string]interface{}
    resp.ToJSON(&data)
    var disks []map[string]string
    tmp, _ := json.Marshal(data["Members"])
    json.Unmarshal(tmp, &disks)
		if len(disks) == 0{
				log.Println("No existing RAID found. Creating Virtual disks")
				return true
		}
		log.Println("Found existing RAID config. Deleting existing RAID")
    tmp1 := strings.Split(hp.RedfishClient.BaseURL, "/")
    disk_url := tmp1[0] + "//" + tmp1[2]
    for _,disk := range disks{
				msg := fmt.Sprintf("Deleting the virtual disk %s", disk["@odata.id"])
				log.Println(msg)
				endpoint = disk_url + disk["@odata.id"]
        r, _ = hp.RedfishClient.HttpClient.Delete(endpoint, header)
        job_id := strings.Split(r.Response().Header["Location"][0], "/")
        job := job_id[len(job_id) - 1]
        job_url := hp.UrlMappings.ManagerURL("") + "/Jobs/" + job
				log.Println("Waiting for the delete job to finish")
        result = hp.checkJobStatus(job_url)
    }
    return result
}

func (h hardwareProfile) CreateVirtualDisks() bool{
		if !h.cleanVirtualDIskIfEExists() {
				return false
		}
		var result bool = false
    for _, tmp := range h.HP  {
  		for _, d := range tmp.Disk {
  				var VolumeType string
  				switch {
  				case d.RaidType == "50":
  						VolumeType = "SpannedStripesWithParity"
  				case d.RaidType == "1":
  						VolumeType = "Mirrored"
  				case d.RaidType == "5":
  						VolumeType = "StripedWithParity"
  				case d.RaidType == "10":
  						VolumeType = "SpannedMirrors"
  				default:
  						VolumeType = "NonRedundant"
  				}
  				var drives []string
          for _, disk := range d.Disk{
              drives = append(drives, fmt.Sprintf(`{"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Storage/Drives/%s"}`, disk))
  						log.Println(fmt.Sprintf("Creating RAID for disk %s", disk))
          }
          tmp := strings.Join(drives, ",")
          drive := `[` + tmp + `]`
  				payload := fmt.Sprintf(`{
  					"VolumeType": "%s",
  					"Name": "%s",
  					"Drives": %s
  					}`, VolumeType, d.Name, drive)
  				endpoint := h.UrlMappings.SystemURL("get_virtual_disks")
  				header := h.RedfishClient.Header
  				r, err := h.RedfishClient.HttpClient.Post(endpoint, header, payload)
  				resp := rp.CheckErrorAndReturn(r,err)
  				job_id := strings.Split(resp.Response().Header["Location"][0], "/")
  				job := job_id[len(job_id) - 1]
  				job_url := h.UrlMappings.ManagerURL("") + "/Jobs/" + job
  				// body := `{}`
  				log.Println(fmt.Sprintf("Waiting for job %s to complete", job))
  				result = h.checkJobStatus(job_url)
  		}
    }
		return resul


*/
