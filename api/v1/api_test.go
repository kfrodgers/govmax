package apiv1

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"GoWBEM/src/gowbem"
)

var smis *SMIS

var testingSID string
var testingInstance *gowbem.InstanceName

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
		instance, e := smis.GetInstance(&array, false, nil)
		if nil != e {
			continue
		}
		pr, _ := smis.GetPropertyByName(instance, "ElementName")
		testingSID = pr.(string)
		testingInstance = &array
	}
}

func MyGetenv(key string, defaultValue string) string {
	v, found := os.LookupEnv(key)
	if !found {
		v = defaultValue
	}
	return v
}

func DumpInstanceName(name *gowbem.InstanceName) {
	instance, err := smis.GetInstance(name, true, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	DumpInstanceClass(name)
	DumpInstance(instance)
}

func DumpInstanceClass(name *gowbem.InstanceName) {
	fmt.Print(name.ClassName, "={")
	for i, key := range name.KeyBinding {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Print(key.Name, ":", key.KeyValue.KeyValue)
	}
	fmt.Println("}")
}

func DumpInstance(instance *gowbem.Instance) {
	for _, pr := range instance.Property {
		fmt.Println("\t", pr.Name, "=", pr.Value)
	}
}

func DumpClassNames(class_name string) []gowbem.Class {
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

func DumpParmValues(values []gowbem.ParamValue) {
	for _, p := range values {
		if p.Instance != nil {
			DumpInstance(p.Instance)
		} else if p.InstanceName != nil {
			DumpInstanceName(p.InstanceName)
		} else if p.Value != nil {
			fmt.Println("Value = ", p.Value.Value)
		} else if p.ValueReference != nil {
			if p.ValueReference.InstanceName != nil {
				DumpInstanceName(p.ValueReference.InstanceName)
			} else if p.ValueReference.InstancePath != nil {
				DumpInstanceName(p.ValueReference.InstancePath.InstanceName)
			} else {
				fmt.Println("ValueReference = ", p.ValueReference)
			}
		} else {
			fmt.Println("p = ", p)
		}
	}
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
		DumpInstanceName(&array)
		swIdent, err2 := smis.GetSoftwareIdentity(&array)
		if err2 != nil {
			panic(err2)
		}
		DumpInstance(swIdent)
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
		if entry.InstancePath == nil {
			panic("nil InstancePath")
		}
		poolSettings, err := smis.GetStoragePoolSettings(entry.InstancePath.InstanceName)
		if err != nil {
			panic(err)
		}
		for _, obj := range poolSettings {
			DumpInstanceName(obj.InstancePath.InstanceName)
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
		devId, _ := smis.GetKeyFromInstanceName(entry.InstancePath.InstanceName, "DeviceID")
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
		DumpInstanceClass(entry.InstancePath.InstanceName)
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
		DumpInstanceClass(entry.InstancePath.InstanceName)
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
		DumpInstanceClass(entry.InstancePath.InstanceName)
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
		DumpInstanceClass(entry.InstancePath.InstanceName)
	}
}

func TestGetVolumeByID(*testing.T) {

	var volumeId interface{}

	vols, err := smis.GetVolumes(testingInstance)
	if err != nil {
		panic(err)
	}
	idx := len(vols) / 2
	volumeId, err = smis.GetKeyFromInstanceName(vols[idx].InstancePath.InstanceName, "DeviceID")
	if err != nil {
		panic(err)
	}
	fmt.Println("Looking for volume = ", volumeId.(string))

	vol, err := smis.GetVolumeByID(testingInstance, volumeId.(string))
	if err != nil {
		panic(err)
	}

	DumpInstanceClass(vol)
}

func TestGetVolumeByName(*testing.T) {

	var volumeInstance *gowbem.Instance

	vols, err := smis.GetVolumes(testingInstance)
	if err != nil {
		panic(err)
	}
	volumeInstance, err = smis.GetInstance(vols[0].InstancePath.InstanceName, false, nil)
	if err != nil {
		panic(err)
	}
	nameProp, _ := smis.GetPropertyByName(volumeInstance, "ElementName")
	fmt.Println("Looking for volume = ", nameProp.(string))

	foundVols, err := smis.GetVolumeByName(testingInstance, nameProp.(string))
	if err != nil {
		panic(err)
	}
	for _, v := range foundVols {
		DumpInstanceName(v)
	}
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

func TestPostVolumes(*testing.T) {

	PostVolRequest := &PostVolumesReq{
		ElementName:        "govmax_test_vol",
		ElementType:        "2",
		EMCNumberOfDevices: "1",
		Size:               "123",
	}

	pools, _ := smis.GetStoragePools(testingInstance)
	PostVolRequest.InPool = pools[0].InstancePath.InstanceName

	volumes, err := smis.PostVolumes(PostVolRequest, testingInstance)
	if err != nil {
		panic(err)
	}
	for _, p := range volumes {
		DumpInstanceName(p.InstancePath.InstanceName)
	}
}

func TestGetStoragePoolSettings(*testing.T) {
	storagePools, err := smis.GetStoragePools(testingInstance)
	if err != nil {
		panic(err)
	}
	if len(storagePools) == 0 {
		panic("No pools found")
	}

	storagePoolSettings, err := smis.GetStoragePoolSettings(storagePools[0].InstancePath.InstanceName)
	if err != nil {
		panic(err)
	}
	for _, sp := range storagePoolSettings {
		DumpInstanceName(sp.InstancePath.InstanceName)
	}
}

func TestPostCreateGroup(*testing.T) {
	curTime := time.Now()
	groupName := "govmax_sg_" + strconv.FormatInt(curTime.Unix(), 16)
	fmt.Println("group = ", groupName)

	storageGroup, err := smis.PostCreateGroup(testingInstance, groupName, 4)
	if err != nil {
		panic(err)
	}
	DumpInstanceName(storageGroup.InstanceName)

	err = smis.PostDeleteGroup(testingInstance, storageGroup, false)
	if err != nil {
		panic(err)
	}
	panic("done")
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
	err := smis.PostDeleteGroup(testingInstance, nil, false)
	if err != nil {
		panic(err)
	}
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

func TestPostPortLogins(*testing.T) {

	endpoints, err := smis.GetTargetEndpoints(testingInstance)
	if err != nil {
		panic(err)
	}

	for _, entry := range endpoints {
		fmt.Println(entry)
	}

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
