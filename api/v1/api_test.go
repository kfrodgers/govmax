package apiv1

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/runner-mei/gowbem"
)

var smis *SMIS

var testingSID string
var testingInstance gowbem.CIMInstanceName

func init() {
	host := MyGetenv("GOVMAX_SMISHOST", "10.108.247.22")
	port := MyGetenv("GOVMAX_SMISPORT", "5988")
	insecure, _ := strconv.ParseBool(MyGetenv("GOVMAX_INSECURE", "true"))
	username := MyGetenv("GOVMAX_USERNAME", "admin")
	password := MyGetenv("GOVMAX_PASSWORD", "#1Password")

	var err error
	smis, err = New(host, port, insecure, username, password)
	if err != nil {
		panic(err)
	}
	arrays, err := smis.GetStorageArrays()
	//Setup array name for rest of tests
	if err != nil {
		panic(err)
	}
	for _, array := range arrays {
		fmt.Println(array)
		instance, e := smis.GetInstanceByInstanceName(array, nil)
		if nil != e {
			continue
		}
		pr := instance.GetPropertyByName("ElementName")
		testingSID = pr.GetValue().(string)
		testingInstance = array
		serviceName, _ := smis.GetStorageConfigurationService(testingInstance)
		fmt.Println(serviceName)
		serviceName, _ = smis.GetControllerConfigurationService(testingInstance)
		fmt.Println(serviceName)
	}

	names := DumpClassNames("")
	for _, n := range names {
		names2 := DumpClassNames(n)
		for _, n2 := range names2 {
			DumpClassNames(n2)
		}
	}
	panic("x")
}

func MyGetenv(key string, defaultValue string) string {
	v, found := os.LookupEnv(key)
	if !found {
		v = defaultValue
	}
	return v
}

func DumpInstanceName(name gowbem.CIMInstanceName) {
	instance, err := smis.GetInstanceByInstanceName(name, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(name)
	for i := 0; i < instance.GetPropertyCount(); i++ {
		pr := instance.GetPropertyByIndex(i)
		fmt.Println("\t", pr.GetName(), "=", pr.GetValue())
	}
}

func DumpClassNames(class_name string) []string {
	names, err := smis.EnumerateClassNames(class_name, false)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println(class_name)
	for _, name := range names {
		fmt.Println("\t", name)
	}
	return names
}

func TestGetStorageArrays(*testing.T) {

	arrays, err := smis.GetStorageArrays()
	//Setup array name for rest of tests
	if err != nil {
		panic(err)
	}

	testingSID = ""
	fmt.Println(testingSID)
	for _, array := range arrays {
		fmt.Println(array)

		instance, e := smis.GetInstanceByInstanceName(array, nil)
		if nil != e {
			continue
		}
		pr := instance.GetPropertyByName("ElementName")
		testingSID = pr.GetValue().(string)
		testingInstance = array
	}
}

func TestGetStoragePools(*testing.T) {

	pools, err := smis.GetStoragePools(testingInstance)
	if err != nil {
		panic(err)
	}
	if len(pools) == 0 {
		panic("empty list")
	}
	for _, entry := range pools {
		fmt.Println(fmt.Sprintf("%+v", entry))

		poolSettings, err := smis.GetStoragePoolSettings(entry)
		if err != nil {
			panic(err)
		}
		for _, obj := range poolSettings {
			fmt.Println("\t", fmt.Sprintf("%+v", obj))
		}
	}
}

func TestGetMaskingViews(*testing.T) {
	maskingViews, err := smis.GetMaskingViews(testingInstance)
	if err != nil {
		panic(err)
	}
	if len(maskingViews) == 0 {
		panic("empty list")
	}

	for _, entry := range maskingViews {
		devId, _ := smis.GetKeyFromInstanceName(entry, "DeviceID")
		fmt.Println(fmt.Sprintf("%+v", devId))
	}
}

func TestGetStorageGroups(*testing.T) {

	groups, err := smis.GetStorageGroups(testingInstance)
	if err != nil {
		panic(err)
	}
	if len(groups) == 0 {
		panic("empty list")
	}

	for _, entry := range groups {
		fmt.Println(fmt.Sprintf("%+v", entry))
	}
}

func TestGetVolumes(*testing.T) {

	inst, e := smis.GetStorageInstanceName(testingSID)
	if e != nil {
		panic(e)
	}

	vols, err := smis.GetVolumes(inst)
	if err != nil {
		panic(err)
	}

	for _, entry := range vols {
		fmt.Println(fmt.Sprintf("%+v", entry))
	}
}

func TestPortGroups(*testing.T) {

	portGroups, err := smis.GetPortGroups(testingInstance)
	if err != nil {
		panic(err)
	}
	if len(portGroups) == 0 {
		panic("empty list")
	}

	for _, entry := range portGroups {
		fmt.Println(fmt.Sprintf("%+v", entry))
	}
}

func TestInitiatorGroups(*testing.T) {

	initGroups, err := smis.GetHostGroups(testingInstance)
	if err != nil {
		panic(err)
	}
	if len(initGroups) == 0 {
		panic("empty list")
	}

	for _, entry := range initGroups {
		fmt.Println(fmt.Sprintf("%+v", entry))
	}
}

func TestGetVolumeByID(*testing.T) {

	var volumeId interface{}

	vols, err := smis.GetVolumes(testingInstance)
	if err != nil {
		panic(err)
	}
	idx := len(vols) / 2
	volumeId, err = smis.GetKeyFromInstanceName(vols[idx], "DeviceID")
	if err != nil {
		panic(err)
	}
	fmt.Println("Looking for volume = ", volumeId.(string))

	vol, err := smis.GetVolumeByID(testingInstance, volumeId.(string))
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", *vol))
}

func TestGetVolumeByName(*testing.T) {

	var volumeInstance gowbem.CIMInstance

	vols, err := smis.GetVolumes(testingInstance)
	if err != nil {
		panic(err)
	}
	volumeInstance, err = smis.GetInstanceByInstanceName(vols[0], nil)
	if err != nil {
		panic(err)
	}
	nameProp := volumeInstance.GetPropertyByName("ElementName")
	fmt.Println("Looking for volume = ", nameProp.GetValue().(string))

	vol, err := smis.GetVolumeByName(testingInstance, nameProp.GetValue().(string))
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", *vol))
}

func TestGetSLOs(*testing.T) {
	if smis.IsArrayV3(testingInstance) {
		SLOs, err := smis.GetSLOs(testingInstance)
		if err != nil {
			panic(err)
		}

		for _, entry := range SLOs {
			fmt.Printf("%+v\n", entry)
		}
	}
}

func TestPostPortLogins(*testing.T) {

	endpoints, err := smis.GetTargetEndpoints(testingInstance)
	if err != nil {
		panic(err)
	}

	for _, entry := range endpoints {
		DumpInstanceName(entry)
	}
	panic("done")

	PostPortLoginsReq := &PostPortLoggedInReq{
		PostPortLoggedInRequestContent: &PostPortLoggedInReqContent{
			PostPortLoggedInRequestHardwareID: &PostPortLoggedInReqHardwareID{
				AtType:     "http://schemas.emc.com/ecom/edaa/root/emc/SE_StorageHardwareID",
				InstanceID: "10000000C94E5D22",
			},
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageHardwareIDManagementService",
		},
	}

	portValues, err := smis.PostPortLogins(PostPortLoginsReq, testingSID)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(portValues); i++ {
		fmt.Println("Port Number:" + portValues[i].PortNumber + " Director:" + portValues[i].Director + " WWN:" + portValues[i].WWN)
	}
}

func TestPostVolumes(*testing.T) {

	PostVolRequest := &PostVolumesReq{
		PostVolumesRequestContent: &PostVolumesReqContent{
			AtType:             "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageConfigurationService",
			ElementName:        "test_vol",
			ElementType:        "2",
			EMCNumberOfDevices: "1",
			Size:               "123",
		},
	}
	queuedJob, _, err := smis.PostVolumes(PostVolRequest, testingSID)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", queuedJob))
}

func TestPostCreateGroup(*testing.T) {
	curTime := time.Now()
	PostGroupRequest := &PostGroupReq{
		PostGroupRequestContent: &PostGroupReqContent{
			AtType:    "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerConfigurationService",
			GroupName: "TestingSG_" + curTime.Format("Jan-2-2006--15-04-05"),
			Type:      "4",
		},
	}
	storageGroup, err := smis.PostCreateGroup(PostGroupRequest, testingSID)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", storageGroup))
}

func TestGetStoragePoolSettings(*testing.T) {
	storagePools, err := smis.GetStoragePools(testingInstance)
	if err != nil {
		panic(err)
	}

	storagePoolSettings, err := smis.GetStoragePoolSettings(storagePools[0])
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", storagePoolSettings))
}

func TestPostVolumeToSG(*testing.T) {

	PostVol2SGRequest := &PostVolumesToSGReq{
		PostVolumesToSGRequestContent: &PostVolumesToSGReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerconfigurationService",
			PostVolumesToSGRequestContentMG: &PostVolumesToSGReqContentMG{
				AtType: "http://schemas.emc.com/ecom/edaa/root/emc/SE_DeviceMaskingGroup",
				//Change SMI_sg2 to any existing Storage Group ID
				InstanceID: "SYMMETRIX-+-" + testingSID + "-+-SMI_sg2",
			},
			PostVolumesToSGRequestContentMember: []*PostVolumesToSGReqContentMember{
				&PostVolumesToSGReqContentMember{
					AtType:            "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageVolume",
					CreationClassName: "Symm_StorageVolume",
					//Change DeviceID to existing Volume ID
					DeviceID:                "00051",
					SystemCreationClassName: "Symm_StorageSystem",
					SystemName:              "SYMMETRIX-+-" + testingSID,
				},
			},
		},
	}

	vol2SG, err := smis.PostVolumesToSG(PostVol2SGRequest, testingSID)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", vol2SG))
}

func TestPostStorageHardwareID(*testing.T) {
	PostSHIDRequest := &PostStorageHardwareIDReq{
		PostStorageHardwareIDRequestContent: &PostStorageHardwareIDReqContent{
			AtType:    "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageHardwareIDManagementService",
			IDType:    "2",
			StorageID: "10000000c94e5d22",
		},
	}
	storageGroup, err := smis.PostStorageHardwareID(PostSHIDRequest, testingSID)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", storageGroup))
}

func TestPostInitiatorToHG(*testing.T) {

	PostInit2SGRequest := &PostInitiatorToHGReq{
		PostInitiatorToHGRequestContent: &PostInitiatorToHGReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerconfigurationService",
			PostInitiatorToHGRequestContentMG: &PostInitiatorToHGReqContentMG{
				AtType: "http://schemas.emc.com/ecom/edaa/root/emc/SE_InitiatorMaskingGroup",
				//Change test_hg to any existing Host Group ID
				InstanceID: "SYMMETRIX-+-" + testingSID + "-+-test_hg",
			},
			PostInitiatorToHGRequestContentMember: []*PostInitiatorToHGReqContentMember{
				&PostInitiatorToHGReqContentMember{
					AtType: "http://schemas.emc.com/ecom/edaa/root/emc/SE_StorageHardwareID",
					//Change InstanceID to existing Initiator
					InstanceID: "10000000C94E5D22",
				},
			},
		},
	}

	init2HG, err := smis.PostInitiatorToHG(PostInit2SGRequest, testingSID)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", init2HG))
}

func TestPostPortToPG(*testing.T) {

	PostPort2PGRequest := &PostPortToPGReq{
		PostPortToPGRequestContent: &PostPortToPGReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerconfigurationService",
			PostPortToPGRequestContentMG: &PostPortToPGReqContentMG{
				AtType: "http://schemas.emc.com/ecom/edaa/root/emc/SE_TargetMaskingGroup",
				//Change test_pg to any existing Port Group ID
				InstanceID: "SYMMETRIX-+-" + testingSID + "-+-test_pg",
			},
			PostPortToPGRequestContentMember: []*PostPortToPGReqContentMember{
				&PostPortToPGReqContentMember{
					AtType:            "http://schemas.emc.com/ecom/edaa/root/emc/Symm_FCSCSIProtocolEndpoint",
					CreationClassName: "Symm_FCSCSIProtocolEndpoint",
					//Change Name to existing FE port
					Name: "5000097350159009",
					SystemCreationClassName: "Symm_StorageProcessorSystem",
					//Change to existing director and port
					SystemName: "SYMMETRIX-+-" + testingSID + "-+-FA-1D-+-9",
				},
			},
		},
	}

	port2PG, err := smis.PostPortToPG(PostPort2PGRequest, testingSID)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", port2PG))
}

func TestPostCreateMaskingView(*testing.T) {

	PostCreateMaskingViewReq := &PostCreateMaskingViewReq{
		PostCreateMaskingViewRequestContent: &PostCreateMaskingViewReqContent{
			AtType:      "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerconfigurationService",
			ElementName: "Sak_MV_TEST",
			PostInitiatorMaskingGroupRequest: &PostInitiatorMaskingGroupReq{
				AtType:     "http://schemas.emc.com/ecom/edaa/root/emc/SE_InitiatorMaskingGroup",
				InstanceID: "SYMMETRIX-+-000196701380-+-tt",
			},
			PostTargetMaskingGroupRequest: &PostTargetMaskingGroupReq{
				AtType:     "http://schemas.emc.com/ecom/edaa/root/emc/SE_TargetMaskingGroup",
				InstanceID: "SYMMETRIX-+-000196701380-+-ttt",
			},
			PostDeviceMaskingGroupRequest: &PostDeviceMaskingGroupReq{
				AtType:     "http://schemas.emc.com/ecom/edaa/root/emc/SE_DeviceMaskingGroup",
				InstanceID: "SYMMETRIX-+-000196701380-+-SMI_sg2",
			},
		},
	}

	fmt.Println(fmt.Sprintf("%+v", PostCreateMaskingViewReq))

	storageGroup, err := smis.PostCreateMaskingView(PostCreateMaskingViewReq, testingSID)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", storageGroup))
}

func TestGetBaremetalHBA(*testing.T) {

	HBAs, err := GetBaremetalHBA()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", HBAs)
}

func TestRemoveVolumeFromSG(*testing.T) {

	RemVol2SGRequest := &PostVolumesToSGReq{
		PostVolumesToSGRequestContent: &PostVolumesToSGReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerconfigurationService",
			PostVolumesToSGRequestContentMG: &PostVolumesToSGReqContentMG{
				AtType: "http://schemas.emc.com/ecom/edaa/root/emc/SE_DeviceMaskingGroup",
				//Change SMI_sg2 to any existing Storage Group ID
				InstanceID: "SYMMETRIX-+-" + testingSID + "-+-Kim_SG",
			},
			PostVolumesToSGRequestContentMember: []*PostVolumesToSGReqContentMember{
				&PostVolumesToSGReqContentMember{
					AtType:            "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageVolume",
					CreationClassName: "Symm_StorageVolume",
					//Change DeviceID to existing Volume ID
					DeviceID:                "0001",
					SystemCreationClassName: "Symm_StorageSystem",
					SystemName:              "SYMMETRIX-+-" + testingSID,
				},
			},
		},
	}

	rmVol, err := smis.RemoveVolumeFromSG(RemVol2SGRequest, testingSID)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", rmVol))
}

func TestRemovePortFromPG(*testing.T) {

	RemPort2PGRequest := &PostPortToPGReq{
		PostPortToPGRequestContent: &PostPortToPGReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerconfigurationService",
			PostPortToPGRequestContentMG: &PostPortToPGReqContentMG{
				AtType: "http://schemas.emc.com/ecom/edaa/root/emc/SE_TargetMaskingGroup",
				//Change test_pg to any existing Port Group ID
				InstanceID: "SYMMETRIX-+-" + testingSID + "-+-ttt",
			},
			PostPortToPGRequestContentMember: []*PostPortToPGReqContentMember{
				&PostPortToPGReqContentMember{
					AtType:            "http://schemas.emc.com/ecom/edaa/root/emc/Symm_FCSCSIProtocolEndpoint",
					CreationClassName: "Symm_FCSCSIProtocolEndpoint",
					//Change Name to existing FE port
					Name: "500009735015908a",
					SystemCreationClassName: "Symm_StorageProcessorSystem",
					//Change to existing director and port
					SystemName: "SYMMETRIX-+-" + testingSID + "-+-FA-3D-+-10",
				},
			},
		},
	}

	rmPort, err := smis.RemovePortFromPG(RemPort2PGRequest, testingSID)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", rmPort))
}

func TestRemoveInitiatorFromHG(*testing.T) {

	RemInit2SGRequest := &PostInitiatorToHGReq{
		PostInitiatorToHGRequestContent: &PostInitiatorToHGReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerconfigurationService",
			PostInitiatorToHGRequestContentMG: &PostInitiatorToHGReqContentMG{
				AtType: "http://schemas.emc.com/ecom/edaa/root/emc/SE_InitiatorMaskingGroup",
				//Change test_hg to any existing Host Group ID
				InstanceID: "SYMMETRIX-+-" + testingSID + "-+-test_hg",
			},
			PostInitiatorToHGRequestContentMember: []*PostInitiatorToHGReqContentMember{
				&PostInitiatorToHGReqContentMember{
					AtType: "http://schemas.emc.com/ecom/edaa/root/emc/SE_StorageHardwareID",
					//Change InstanceID to existing Initiator
					InstanceID: "10000000C94E5D22",
				},
			},
		},
	}

	rmInit, err := smis.RemoveInitiatorFromHG(RemInit2SGRequest, testingSID)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", rmInit))
}

func TestPostDeleteGroup(*testing.T) {
	DeleteGroupRequest := &DeleteGroupReq{
		DeleteGroupRequestContent: &DeleteGroupReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerConfigurationService",
			DeleteGroupRequestContentMaskingGroup: &DeleteGroupReqContentMaskingGroup{
				//Change AtType to type of Group and InstanceID to existing name of Group
				AtType:     "http://schemas.emc.com/ecom/edaa/root/emc/SE_DeviceMaskingGroup",
				InstanceID: "SYMMETRIX-+-" + testingSID + "-+-Test_SG",
			},
		},
	}
	deleteGroup, err := smis.PostDeleteGroup(DeleteGroupRequest, testingSID)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", deleteGroup))
}

func TestPostDeleteVol(*testing.T) {
	DeleteVolumeRequest := &DeleteVolReq{
		DeleteVolRequestContent: &DeleteVolReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageConfigurationService",
			DeleteVolRequestContentElement: &DeleteVolReqContentElement{
				AtType:                  "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageVolume",
				DeviceID:                "0001F",
				CreationClassName:       "Symm_StorageVolume",
				SystemName:              "SYMMETRIX-+-" + testingSID,
				SystemCreationClassName: "Symm_StorageSystem",
			},
		},
	}
	deleteVol, err := smis.PostDeleteVol(DeleteVolumeRequest, testingSID)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", deleteVol))
}

func TestPostDeleteMV(*testing.T) {
	DeleteMVRequest := &DeleteMaskingViewReq{
		DeleteMaskingViewRequestContent: &DeleteMaskingViewReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerConfigurationService",
			DeleteMaskingViewRequestContentPC: &DeleteMaskingViewReqContentPC{
				AtType:                  "http://schemas.emc.com/ecom/edaa/root/emc/Symm_LunMaskingView",
				DeviceID:                "Kim_MV",
				CreationClassName:       "Symm_LunMaskingView",
				SystemName:              "SYMMETRIX-+-" + testingSID,
				SystemCreationClassName: "Symm_StorageSystem",
			},
		},
	}

	deleteMV, err := smis.PostDeleteMaskingView(DeleteMVRequest, testingSID)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", deleteMV))
}
