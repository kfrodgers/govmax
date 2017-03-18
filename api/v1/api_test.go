package apiv1

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kfrodgers/GoWBEM/src/gowbem"
)

var smis *SMIS

var testingSID string
var testingInstance *gowbem.InstanceName

func init() {
	host := MyGetenv("GOVMAX_SMISHOST", "")
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
		pr, _ := GetPropertyByName(instance, "ElementName")
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

func TestGetStorageArrays(t *testing.T) {

	arrays, err := smis.GetStorageArrays()
	//Setup array name for rest of tests
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	testingSID = ""
	fmt.Println(testingSID)
	for _, array := range arrays {
		DumpInstanceClass(&array)
		swIdent, err2 := smis.GetSoftwareIdentity(&array)
		if err2 != nil {
			t.Log(err2.Error())
			t.Fail()
			return
		}
		DumpInstance(swIdent)
	}
}

func TestGetStoragePools(t *testing.T) {
	pools, err := smis.GetStoragePools(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	if len(pools) == 0 {
		t.Log("empty list")
		t.Fail()
		return
	}
	for _, entry := range pools {
		if entry.InstancePath == nil {
			t.Log("nil InstancePath")
			t.Fail()
			return
		}
		poolSettings, err := smis.GetStoragePoolSettings(entry.InstancePath.InstanceName)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		for _, obj := range poolSettings {
			DumpInstanceClass(obj.InstancePath.InstanceName)
		}
	}
}

func TestGetMaskingViews(t *testing.T) {
	maskingViews, err := smis.GetMaskingViews(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	if len(maskingViews) == 0 {
		t.Log("empty list")
		t.Fail()
		return
	}

	for _, entry := range maskingViews {
		devId, _ := GetKeyFromInstanceName(entry.InstancePath.InstanceName, "DeviceID")
		fmt.Println(devId.(string))
	}
}

func TestGetStorageGroups(t *testing.T) {

	groups, err := smis.GetStorageGroups(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	if len(groups) == 0 {
		t.Log("empty list")
		t.Fail()
		return
	}

	for _, entry := range groups {
		DumpInstanceClass(entry.InstancePath.InstanceName)
	}
}

func TestGetVolumes(t *testing.T) {

	inst, e := smis.GetStorageInstanceName(testingSID)
	if e != nil {
		t.Log(e.Error())
		t.Fail()
		return
	}

	vols, err := smis.GetVolumes(inst)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	for _, entry := range vols {
		DumpInstanceClass(entry.InstancePath.InstanceName)
	}
}

func TestPortGroups(t *testing.T) {

	portGroups, err := smis.GetPortGroups(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	if len(portGroups) == 0 {
		t.Log("empty list")
		t.Fail()
		return
	}

	for _, entry := range portGroups {
		DumpInstanceClass(entry.InstancePath.InstanceName)
	}
}

func TestInitiatorGroups(t *testing.T) {

	initGroups, err := smis.GetHostGroups(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	if len(initGroups) == 0 {
		t.Log("empty list")
		t.Fail()
		return
	}

	for _, entry := range initGroups {
		DumpInstanceClass(entry.InstancePath.InstanceName)
	}
}

func TestGetInitiators(t *testing.T) {

	initiators, err := smis.GetScsiInitiators(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	if len(initiators) == 0 {
		t.Log("empty list")
		t.Fail()
		return
	}

	for _, entry := range initiators {
		key, _ := GetKeyFromInstanceName(entry.InstancePath.InstanceName, "InstanceID")
		fmt.Println("initiator=", key.(string))
	}
}

func TestGetVolumeByID(t *testing.T) {

	var volumeId interface{}

	vols, err := smis.GetVolumes(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	idx := len(vols) / 2
	volumeId, err = GetKeyFromInstanceName(vols[idx].InstancePath.InstanceName, "DeviceID")
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	fmt.Println("Looking for volume = %s", volumeId.(string))

	vol, err := smis.GetVolumeByID(testingInstance, volumeId.(string))
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	DumpInstanceClass(vol)
}

func TestGetVolumeByName(t *testing.T) {

	var volumeInstance *gowbem.Instance

	vols, err := smis.GetVolumes(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	volumeInstance, err = smis.GetInstance(vols[0].InstancePath.InstanceName, false, nil)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	nameProp, _ := GetPropertyByName(volumeInstance, "ElementName")
	fmt.Println("Looking for volume = ", nameProp.(string))

	foundVols, err := smis.GetVolumeByName(testingInstance, nameProp.(string))
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	for _, v := range foundVols {
		DumpInstanceClass(v)
	}
}

func TestGetSLOs(t *testing.T) {
	if smis.IsArrayV3(testingInstance) {
		SLOs, err := smis.GetSLOs(testingInstance)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}

		for _, entry := range SLOs {
			t.Logf("%+v\n", entry)
		}
	}
}

func TestPostVolumes(t *testing.T) {

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
		t.Log(err.Error())
		t.Fail()
		return
	}

	var volPaths []gowbem.InstancePath
	for _, p := range volumes {
		DumpInstanceClass(p.InstancePath.InstanceName)
		volPaths = append(volPaths, *p.InstancePath)
	}

	err = smis.PostDeleteVol(testingInstance, volPaths)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
}

func TestGetStoragePoolSettings(t *testing.T) {
	storagePools, err := smis.GetStoragePools(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	if len(storagePools) == 0 {
		t.Log("No pools found")
		t.Fail()
		return
	}

	storagePoolSettings, err := smis.GetStoragePoolSettings(storagePools[0].InstancePath.InstanceName)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	for _, sp := range storagePoolSettings {
		DumpInstanceClass(sp.InstancePath.InstanceName)
	}
}

func TestPostCreateGroup(t *testing.T) {
	curTime := time.Now()
	groupName := "govmax_sg_" + strconv.FormatInt(curTime.Unix(), 16)
	fmt.Println("creating group = ", groupName)

	storageGroup, err := smis.PostCreateGroup(testingInstance, groupName, 4)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	DumpInstanceClass(storageGroup.InstanceName)

	err = smis.PostDeleteGroup(testingInstance, storageGroup, false)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
}

func TestAddRemoveFromGroup(t *testing.T) {
	ports, err := smis.GetTargetEndpoints(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	if len(ports) < 2 {
		t.Log("not enough ports available")
		t.Fail()
		return
	}

	var members []gowbem.InstancePath
	members = append(members, *ports[0].InstancePath)
	members = append(members, *ports[1].InstancePath)

	portGroup, err := smis.PostCreateGroup(testingInstance, "govmax_test_pg", 3)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	DumpInstanceClass(portGroup.InstanceName)

	err = smis.AddMembersToGroup(testingInstance, portGroup, members)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	fmt.Println("Added = ", members)

	err = smis.RemoveMembersFromGroup(testingInstance, portGroup, members)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	fmt.Println("Removed = ", members)

	err = smis.PostDeleteGroup(testingInstance, portGroup, false)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
}

func TestPostStorageHardwareID(t *testing.T) {
	newInit, err := smis.PostStorageHardwareID(testingInstance, "10000000C94E5D22", 2)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	DumpInstanceClass(newInit.InstanceName)

	err = smis.DeleteStorageHardwareID(testingInstance, newInit)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
}

func TestGetBaremetalHBA(t *testing.T) {

	HBAs, err := GetBaremetalHBA()
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	} else {
		fmt.Println("%+v", HBAs)
	}
}

func TestPostPortLogins(t *testing.T) {

	endpoints, err := smis.GetScsiInitiators(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	for _, ep := range endpoints {
		portValues, err := smis.PostPortLogins(testingInstance, ep.InstancePath)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}
		for i := 0; i < len(portValues); i++ {
			fmt.Println("Port Number:" + portValues[i].PortNumber + " Director:" + portValues[i].Director + " WWN:" + portValues[i].WWN)
		}
	}
}

func TestPostDeleteMV(t *testing.T) {
	var mvName string = "xxxxxx_mv"

	mvs, err := smis.GetMaskingViews(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	var found bool = false
	for _, mv := range mvs {
		name, _ := GetKeyFromInstanceName(mv.InstancePath.InstanceName, "DeviceID")
		if name.(string) == mvName {
			DumpInstanceClass(mv.InstancePath.InstanceName)
			err = smis.PostDeleteMaskingView(testingInstance, mv.InstancePath)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
			}
			found = true
			break
		}
	}
	if !found {
		t.Log(mvName + ": not found")
		t.Fail()
	}
}

func TestPostCreateMaskingView(t *testing.T) {
	var sgName string = "xxxxxx_sg"
	var igName string = "xxxxxx_ig"
	var pgName string = "xxxxxx_pg"
	var mvName string = "xxxxxx_mv"

	sgs, err := smis.GetStorageGroups(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	igs, err := smis.GetHostGroups(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	pgs, err := smis.GetPortGroups(testingInstance)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	var viewSg *gowbem.InstancePath
	for _, sg := range sgs {
		name, _ := GetKeyFromInstanceName(sg.InstancePath.InstanceName, "InstanceID")
		if strings.HasSuffix(name.(string), sgName) {
			DumpInstanceClass(sg.InstancePath.InstanceName)
			viewSg = sg.InstancePath
			break
		}
	}
	if viewSg == nil {
		t.Log(sgName + ": sg not found")
		t.Fail()
	}

	var viewIg *gowbem.InstancePath
	for _, ig := range igs {
		name, _ := GetKeyFromInstanceName(ig.InstancePath.InstanceName, "InstanceID")
		if strings.HasSuffix(name.(string), igName) {
			DumpInstanceClass(ig.InstancePath.InstanceName)
			viewIg = ig.InstancePath
			break
		}
	}
	if viewIg == nil {
		t.Log(igName + ": ig not found")
		t.Fail()
	}

	var viewPg *gowbem.InstancePath
	for _, pg := range pgs {
		name, _ := GetKeyFromInstanceName(pg.InstancePath.InstanceName, "InstanceID")
		if strings.HasSuffix(name.(string), pgName) {
			DumpInstanceClass(pg.InstancePath.InstanceName)
			viewPg = pg.InstancePath
			break
		}
	}
	if viewPg == nil {
		t.Log(pgName + ": pg not found")
		t.Fail()
	}

	if !t.Failed() {
		t.Log("Making View: " + mvName)
		mv, err := smis.PostCreateMaskingView(testingInstance, mvName, viewSg, viewIg, viewPg)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		for _, o := range mv {
			DumpInstanceClass(o.InstancePath.InstanceName)
		}
	}
}
