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

func TestGetInitiators(*testing.T) {

	initiators, err := smis.GetScsiInitiators(testingInstance)
	if err != nil {
		panic(err)
	}
	if len(initiators) == 0 {
		panic("empty list")
	}

	for _, entry := range initiators {
		key, _ := smis.GetKeyFromInstanceName(entry.InstancePath.InstanceName, "InstanceID")
		fmt.Println("initiator=", key.(string))
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

	var volPaths []gowbem.InstancePath
	for _, p := range volumes {
		DumpInstanceName(p.InstancePath.InstanceName)
		volPaths = append(volPaths, *p.InstancePath)
	}

	err = smis.PostDeleteVol(testingInstance, volPaths)
	if err != nil {
		panic(err)
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
}

func TestAddRemoveFromGroup(*testing.T) {
	ports, err := smis.GetTargetEndpoints(testingInstance)
	if err != nil {
		panic(err)
	}
	if len(ports) < 2 {
		panic("not enough ports available")
	}

	var members []gowbem.InstancePath
	members = append(members, *ports[0].InstancePath)
	members = append(members, *ports[1].InstancePath)

	portGroup, err := smis.PostCreateGroup(testingInstance, "govmax_test_pg", 3)
	if err != nil {
		panic(err)
	}
	DumpInstanceName(portGroup.InstanceName)

	err = smis.AddMembersToGroup(testingInstance, portGroup, members)
	if err != nil {
		panic(err)
	}
	fmt.Println("Added = ", members)

	err = smis.RemoveMembersFromGroup(testingInstance, portGroup, members)
	if err != nil {
		panic(err)
	}
	fmt.Println("Removed = ", members)

	err = smis.PostDeleteGroup(testingInstance, portGroup, false)
	if err != nil {
		panic(err)
	}
}

func TestPostStorageHardwareID(*testing.T) {
	newInit, err := smis.PostStorageHardwareID(testingInstance, "10000000C94E5D22", 2)
	if err != nil {
		panic(err)
	}
	DumpInstanceName(newInit.InstanceName)

	err = smis.DeleteStorageHardwareID(testingInstance, newInit)
	if err != nil {
		panic(err)
	}
}

func TestPostPortLogins(*testing.T) {

	endpoints, err := smis.GetScsiInitiators(testingInstance)
	if err != nil {
		panic(err)
	}

	for _, ep := range endpoints {
		portValues, err := smis.PostPortLogins(testingInstance, ep.InstancePath)
		if err != nil {
			panic(err)
		}
		for i := 0; i < len(portValues); i++ {
			fmt.Println("Port Number:" + portValues[i].PortNumber + " Director:" + portValues[i].Director + " WWN:" + portValues[i].WWN)
		}
	}
	panic("done")
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
