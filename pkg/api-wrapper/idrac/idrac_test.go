package idrac


import  (

	"testing"
)

func TestUpgradeFirmware(t *testing.T){
	filelocation := "/home/ekuamaj/workspace/iDRAC-with-Lifecycle-Controller_Firmware_NKGJW_WN64_3.31.31.31_A00.EXE"
	client := &IdracRedfishClient{
		Username: "root",
		Password: "Abc.1234",
		HostIP: "32.68.250.78",
	
	}
	client.UpgradeFirmware(filelocation)
}