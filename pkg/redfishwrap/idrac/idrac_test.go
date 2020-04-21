package idrac

import (
	"testing"
)

var client = &IdracRedfishClient{
	Username: "root",
	Password: "Abc.1234",
	HostIP:   "32.68.250.78",
}

func TestUpgradeFirmware(t *testing.T) {
	filelocation := "/home/ekuamaj/workspace/iDRAC-with-Lifecycle-Controller_Firmware_NKGJW_WN64_3.31.31.31_A00.EXE"
	client.UpgradeFirmware(filelocation)

}

func TestCheckJobStatus(t *testing.T) {
	jobId := "JID_851862680400"
	client.CheckJobStatus(jobId)
}

func TestGetVirtualDisks(t *testing.T) {
	systemID := "System.Embedded.1"
	controllerID := "RAID.Slot.6-1"
	client.GetVirtualDisks(systemID, controllerID)

}

func TestDeleteVirtualDisk(t *testing.T) {
	systemID := "System.Embedded.1"
	storageID := "Disk.Virtual.4:RAID.Slot.6-1"
	jobid := client.DeletVirtualDisk(systemID, storageID)
	t.Logf("Job ID %v", jobid)
	//	client.CheckJobStatus(jobid)
}

func TestCleanVirtualDisksIfAny(t * testing.T){
	systemID := "System.Embedded.1"
	controllerID := "RAID.Slot.6-1"
	client.CleanVirtualDisksIfAny(systemID,controllerID)
}
/*
name: ephemeral
          #          raid-type: 1
          #          disk:
          #            - Disk.Bay.8:Enclosure.Internal.0-1:RAID.Slot.6-1
          #            - Disk.Bay.9:Enclosure.Internal.0-1:RAID.Slot.6-1
*/

func TestCreateVirtualDisk(t *testing.T){
	systemID := "System.Embedded.1"
	controllerID := "RAID.Slot.6-1"
	volumeType:= "Mirrored"
	name := "ephemeral-1"
	drives := []string {"Disk.Bay.8:Enclosure.Internal.0-1:RAID.Slot.6-1", 
	                           "Disk.Bay.9:Enclosure.Internal.0-1:RAID.Slot.6-1" }
	jobid := client.CreateVirtualDisk(systemID,controllerID,volumeType,name, drives) 
	t.Logf("Job ID %v", jobid)
	client.CheckJobStatus(jobid)

}

func TestGetNodeUUID(t *testing.T){
	systemID := "System.Embedded.1"
	uuid,_ := client.GetNodeUUID(systemID)
    t.Logf("UUID %v", uuid )

}