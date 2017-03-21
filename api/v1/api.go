package apiv1

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kfrodgers/GoWBEM/src/gowbem"
)

///////////////////////////////////////////////////////////////
//            GET a list of Storage Arrays                   //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStorageArrays() ([]string, error) {
	arrays, err := smis.EnumerateInstances("Symm_StorageSystem", true, true, nil)
	if err != nil {
		return nil, err
	}

	var sidArray []string
	for _, array := range arrays {
		sid, err := GetPropertyByName(array.Instance, "ElementName")
		if err != nil {
			continue
		}
		sidArray = append(sidArray, sid.(string))
	}
	return sidArray, nil
}

///////////////////////////////////////////////////////////////
//            GET Storage Instance Name                      //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStorageInstanceName(sid string) (*gowbem.InstanceName, error) {
	arrays, err := smis.EnumerateInstances("Symm_StorageSystem", true, true, nil)
	if err != nil {
		return nil, err
	}
	for _, array := range arrays {
		name, err := GetPropertyByName(array.Instance, "ElementName")
		if err != nil {
			continue
		}
		if name.(string) == sid {
			return array.InstanceName, nil
		}
	}
	return nil, errors.New("Array not found")
}

///////////////////////////////////////////////////////////////
//            GET Storage Configuration Service              //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStorageConfigurationService(systemInstanceName *gowbem.InstanceName) (*gowbem.InstanceName, error) {
	configServices, err := smis.AssociatorNames(systemInstanceName, "", "EMC_StorageConfigurationService", nil, nil)
	if err != nil {
		return nil, err
	}
	if len(configServices) < 1 {
		return nil, errors.New("EMC_StorageConfigurationService: not found")
	}
	return configServices[0].InstancePath.InstanceName, nil
}

///////////////////////////////////////////////////////////////
//            GET Controller Configuration Service           //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetControllerConfigurationService(systemInstanceName *gowbem.InstanceName) (*gowbem.InstanceName, error) {
	controllerServices, err := smis.AssociatorNames(systemInstanceName, "", "EMC_ControllerConfigurationService", nil, nil)
	if err != nil {
		return nil, err
	}
	if len(controllerServices) < 1 {
		return nil, errors.New("EMC_ControllerConfigurationService: not found")
	}
	return controllerServices[0].InstancePath.InstanceName, nil
}

///////////////////////////////////////////////////////////////
//            GET Storage Hardware ID Management Service     //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStorageHardwareIDManagementService(systemInstanceName *gowbem.InstanceName) (*gowbem.InstanceName, error) {
	managementServices, err := smis.AssociatorNames(systemInstanceName, "", "Symm_StorageHardwareIDManagementService", nil, nil)
	if err != nil {
		return nil, err
	}
	if len(managementServices) < 1 {
		return nil, errors.New("Symm_StorageHardwareIDManagementService: not found")
	}
	return managementServices[0].InstancePath.InstanceName, nil
}

///////////////////////////////////////////////////////////////
//            GET Software Identity                          //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetSoftwareIdentity(systemInstanceName *gowbem.InstanceName) (*gowbem.Instance, error) {
	softwareIdents, err := smis.AssociatorInstances(systemInstanceName, "", "Symm_StorageSystemSoftwareIdentity", nil, nil, true, nil)
	if err != nil {
		return nil, err
	}
	if len(softwareIdents) < 1 {
		return nil, errors.New("Symm_StorageSystemSoftwareIdentity: not found")
	}
	return softwareIdents[0].Instance, nil
}

///////////////////////////////////////////////////////////////
//            Is Array V3 or Greater                         //
///////////////////////////////////////////////////////////////

func (smis *SMIS) IsArrayV3(systemInstanceName *gowbem.InstanceName) bool {
	swIdent, err := smis.GetSoftwareIdentity(systemInstanceName)
	if err != nil {
		return false
	}
	var major int
	ucode, e := GetPropertyByName(swIdent, "EMCEnginuityFamily")
	if e != nil {
		major = 0
	} else {
		major, _ = strconv.Atoi(ucode.(string))
	}
	return major >= 5900
}

///////////////////////////////////////////////////////////////
//            GET a list of Storage Pools                    //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStoragePools(systemInstanceName *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	if smis.IsArrayV3(systemInstanceName) {
		return smis.AssociatorNames(systemInstanceName, "", "Symm_SRPStoragePool", nil, nil)
	} else {
		return smis.AssociatorNames(systemInstanceName, "", "Symm_VirtualProvisioningPool", nil, nil)
	}
}

///////////////////////////////////////////////////////////////
//            GET a list of Masking Views                    //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetMaskingViews(systemInstanceName *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	return smis.AssociatorNames(systemInstanceName, "", "Symm_LunMaskingView", nil, nil)
}

///////////////////////////////////////////////////////////////
//         GET a list of Storage (Device) Groups             //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStorageGroups(systemInstanceName *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	controllerService, err := smis.GetControllerConfigurationService(systemInstanceName)
	if err != nil {
		return nil, err
	}
	return smis.AssociatorNames(controllerService, "", "SE_DeviceMaskingGroup", nil, nil)
}

///////////////////////////////////////////////////////////////
//         GET a list of Port (Target) Groups                //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetPortGroups(systemInstanceName *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	controllerService, err := smis.GetControllerConfigurationService(systemInstanceName)
	if err != nil {
		return nil, err
	}
	return smis.AssociatorNames(controllerService, "", "SE_TargetMaskingGroup", nil, nil)
}

///////////////////////////////////////////////////////////////
//         GET a list of Host (Initiator) Groups             //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetHostGroups(systemInstanceName *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	controllerService, err := smis.GetControllerConfigurationService(systemInstanceName)
	if err != nil {
		return nil, err
	}
	return smis.AssociatorNames(controllerService, "", "SE_InitiatorMaskingGroup", nil, nil)
}

///////////////////////////////////////////////////////////////
//            GET a list of Storage Volumes                  //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetVolumes(systemInstance *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	return smis.AssociatorNames(systemInstance, "", "CIM_StorageVolume", nil, nil)
}

///////////////////////////////////////////////////////////
//            GET a Storage Volume by ID                 //
///////////////////////////////////////////////////////////

func (smis *SMIS) GetVolumeByID(systemInstance *gowbem.InstanceName, volumeID string) (*gowbem.InstanceName, error) {
	volumes, err := smis.GetVolumes(systemInstance)
	if err != nil {
		return nil, err
	}
	for _, volume := range volumes {
		name, err := GetKeyFromInstanceName(volume.InstancePath.InstanceName, "DeviceID")
		if err == nil {
			if name.(string) == volumeID {
				return volume.InstancePath.InstanceName, nil
			}
		}
	}
	return nil, errors.New("Volume not found")
}

///////////////////////////////////////////////////////////////
//            GET a list of Storage Processor Systems        //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStorageProcessorSystem(systemInstance *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	return smis.AssociatorNames(systemInstance, "", "Symm_StorageProcessorSystem", nil, nil)
}

///////////////////////////////////////////////////////////////
//            GET a list of SCSI Endpoints (directors)       //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetScsiInitiators(systemInstance *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	service, err := smis.GetStorageHardwareIDManagementService(systemInstance)
	if err != nil {
		return nil, err
	}
	return smis.AssociatorNames(service, "", "SE_StorageHardwareID", nil, nil)
}

///////////////////////////////////////////////////////////////
//            GET a list of SCSI Endpoints (directors)       //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetScsiEndpoints(storageProcessor *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	return smis.AssociatorNames(storageProcessor, "", "CIM_SCSIProtocolEndpoint", nil, nil)
}

///////////////////////////////////////////////////////////////
//            GET a list of Front End Ports                  //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetTargetEndpoints(systemInstance *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	storageProcs, err := smis.GetStorageProcessorSystem(systemInstance)
	if err != nil {
		return nil, err
	}

	var frontEndPorts []gowbem.ObjectPath
	for _, sp := range storageProcs {
		adapter, err := smis.GetInstance(sp.InstancePath.InstanceName, false, nil)
		if err != nil {
			return nil, err
		}
		elementType, err := GetPropertyByName(adapter, "EMCBSPElementType")
		if elementType.(string) != "3" {
			continue
		}

		ports, err := smis.GetScsiEndpoints(sp.InstancePath.InstanceName)
		if err != nil {
			return nil, err
		}
		for _, p := range ports {
			frontEndPorts = append(frontEndPorts, p)
		}
	}
	return frontEndPorts, nil
}

///////////////////////////////////////////////////////////
//            GET a Storage Volume by Name               //
///////////////////////////////////////////////////////////

func (smis *SMIS) GetVolumeByName(systemInstance *gowbem.InstanceName, volumeName string) ([]*gowbem.InstanceName, error) {
	volumes, err := smis.GetVolumes(systemInstance)
	if err != nil {
		return nil, err
	}

	var foundVolumes []*gowbem.InstanceName
	for _, volume := range volumes {
		var volumeInstance *gowbem.Instance
		volumeInstance, err = smis.GetInstance(volume.InstancePath.InstanceName, false, nil)
		if err == nil {
			nameProp, _ := GetPropertyByName(volumeInstance, "ElementName")
			if nameProp == volumeName {
				foundVolumes = append(foundVolumes, volume.InstancePath.InstanceName)
				break
			}
		}
	}
	if len(foundVolumes) > 0 {
		return foundVolumes, nil
	}
	return nil, errors.New("Volume not found")
}

//////////////////////////////////////////////////////////////////////////////////////////////////
//      GET a Job Status                                                                        //
//                                                                                              //
//  2 - New: job has not been started                                                           //
//  3 - Starting: job is moving into running state                                              //
//  4 - Running: job is running                                                                 //
//  5 - Suspended: job is stopped, but can be restarted                                         //
//  6 - Shutting Down: job is moving to an completed/terminated/killed state                    //
//  7 - Completed: job has been completed normally                                              //
//  8 - Terminated: job has been stopped by a terminate state change request                    //
//  9 - Killed: job has stopped by a kill state change request                                  //
//  10 - Exception: job is in an abnormal state due to an error condition                       //
//  11 - Service: job is in a vendor-specific state that supports problem discovery/resolution  //
//  12 - Query Pending: job is waiting for a client to resolve a query                          //
//////////////////////////////////////////////////////////////////////////////////////////////////

func (smis *SMIS) GetJobStatus(jobPath *gowbem.InstancePath) (*gowbem.Instance, string, error) {
	resp, err := smis.GetInstance(jobPath.InstanceName, false, nil)
	if err != nil {
		return nil, "UNKNOWN", err
	}
	jobStatusMap := map[int]string{
		2:  "NEW",
		3:  "STARTING",
		4:  "RUNNING",
		5:  "SUSPENDED",
		6:  "SHUTTING_DOWN",
		7:  "COMPLETED",
		8:  "TERMINATED",
		9:  "KILLED",
		10: "EXCEPTION",
		11: "SERVICE",
		12: "QUERY_PENDING",
	}

	var jobState int
	var jobStatus string
	var ok bool
	value, _ := GetPropertyByName(resp, "JobState")
	jobState, _ = strconv.Atoi(value.(string))
	if jobStatus, ok = jobStatusMap[jobState]; !ok {
		jobStatus = "UNKNOWN"
	}
	return resp, jobStatus, err
}

///////////////////////////////////////////////////////////////
//         GET a list of Host (Initiator) Groups             //
///////////////////////////////////////////////////////////////

func FindJobIndex(returnParams []gowbem.ParamValue) (int, error) {
	for i, param := range returnParams {
		if param.ValueReference != nil &&
			param.ValueReference.InstancePath.InstanceName.ClassName == "SE_ConcreteJob" {
			return i, nil
		}
	}
	return -1, errors.New("SE_ConcreteJob not found")
}

func (smis *SMIS) WaitForJob(jobPath *gowbem.InstancePath, resultClass string) ([]gowbem.ObjectPath, error) {
	var status string
	var err error

	for {
		_, status, err = smis.GetJobStatus(jobPath)
		if err != nil {
			return nil, err
		}
		if status != "RUNNING" {
			break
		}
		time.Sleep(500000000)
	}
	if status != "COMPLETED" {
		return nil, errors.New("Unexpected job status: " + status)
	}

	return smis.AssociatorNames(jobPath.InstanceName, "", resultClass, nil, nil)
}

//////////////////////////////////////
//    REQUEST Structs used for      //
//   volume creation on the VMAX3.  //
//////////////////////////////////////

type PostVolumesReq struct {
	ElementName        string               `json:"ElementName"`
	ElementType        string               `json:"ElementType"`
	EMCNumberOfDevices string               `json:"EMCNumberOfDevices"`
	InPool             *gowbem.InstanceName `json:"InPool"`
	Size               string               `json:"Size"`
}

///////////////////////////////////////////////////////////
//              CREATE a Storage Volume                  //
//     and check for Volume Creation Completion          //
///////////////////////////////////////////////////////////

func (smis *SMIS) PostVolumes(req *PostVolumesReq, systemInstance *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	storage, err := smis.GetStorageConfigurationService(systemInstance)
	if err != nil {
		return nil, err
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "ElementName", Value: &gowbem.Value{req.ElementName}})
	params = append(params, gowbem.IParamValue{Name: "ElementType", Value: &gowbem.Value{req.ElementType}})
	params = append(params, gowbem.IParamValue{Name: "EMCNumberOfDevices", Value: &gowbem.Value{req.EMCNumberOfDevices}})
	params = append(params, gowbem.IParamValue{Name: "InPool", ValueReference: &gowbem.ValueReference{InstanceName: req.InPool}})
	params = append(params, gowbem.IParamValue{Name: "Size", Value: &gowbem.Value{req.Size}})

	_, retValues, jobErr := smis.InvokeMethod(storage, "CreateOrModifyElementFromStoragePool", params)
	if jobErr != nil {
		return nil, jobErr
	}

	idx, _ := FindJobIndex(retValues)
	if idx == -1 {
		return nil, errors.New("Job instance not found")
	}

	return smis.WaitForJob(retValues[idx].ValueReference.InstancePath, "CIM_StorageVolume")
}

///////////////////////////////////////////////////////////////
//                  CREATE an Array Group                    //
//             groupType == 4 for storage Group              //
//             groupType == 3 for port Group                 //
//             groupType == 2 for initiator Group            //
///////////////////////////////////////////////////////////////

func (smis *SMIS) PostCreateGroup(systemInstance *gowbem.InstanceName, groupName string, groupType int) (*gowbem.InstancePath, error) {
	controller, err := smis.GetControllerConfigurationService(systemInstance)
	if err != nil {
		return nil, err
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "GroupName", Value: &gowbem.Value{groupName}})
	params = append(params, gowbem.IParamValue{Name: "Type", Value: &gowbem.Value{strconv.Itoa(groupType)}})

	retValue, retParms, err := smis.InvokeMethod(controller, "CreateGroup", params)
	if err != nil {
		return nil, err
	}
	if retValue != 0 || len(retParms) == 0 {
		return nil, errors.New("Job failed, rc = " + string(retValue))
	}
	return retParms[0].ValueReference.InstancePath, nil
}

///////////////////////////////////////////////////////////////////
//                GET Storage Pool Capabilities                  //
///////////////////////////////////////////////////////////////////

func (smis *SMIS) GetStoragePoolCapabilities(srp_name *gowbem.InstanceName) (*gowbem.InstanceName, error) {
	capabilities, err := smis.EnumerateInstanceNames("Symm_StoragePoolCapabilities")

	name, err := GetKeyFromInstanceName(srp_name, "InstanceID")
	if err != nil {
		return nil, err
	}
	for _, entry := range capabilities {
		key, err := GetKeyFromInstanceName(&entry, "InstanceID")
		if err != nil {
			continue
		}
		if key.(string) == name.(string) {
			return &entry, nil
		}
	}
	return nil, errors.New("Capabilities not found")

}

///////////////////////////////////////////////////////////////
//                GET Storage Pool Settings                  //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStoragePoolSettings(srp_name *gowbem.InstanceName) ([]gowbem.ObjectPath, error) {
	capabilities, err := smis.GetStoragePoolCapabilities(srp_name)
	if err != nil {
		return nil, err
	}
	return smis.AssociatorNames(capabilities, "", "CIM_StorageSetting", nil, nil)
}

///////////////////////////////////////////////////////////////
//        Struct used to store all SLO information           //
///////////////////////////////////////////////////////////////

type SLO_Struct struct {
	SLO_Name    string
	respTime    float64
	SRP         string
	Workload    string
	ElementName string
	InstanceID  string
}

////////////////////////////////////////////////////////////////
//               GET Storage Level Objectives                 //
//                                                            //
//             1 -> Get Storage Pools of VMAX3                //
//    2 -> Get Storage Pool Settings of each Storage Pool     //
//   3 -> Parse out SLO information of VMAX3 and return it    //
////////////////////////////////////////////////////////////////

func (smis *SMIS) GetSLOs(systemInstanceName *gowbem.InstanceName) (SLOs []SLO_Struct, err error) {
	if !smis.IsArrayV3(systemInstanceName) {
		return nil, errors.New("SLOs not supportted")
	}

	storagePools, err := smis.GetStoragePools(systemInstanceName)
	if err != nil {
		return nil, err
	}

	for _, SRP := range storagePools {
		storagePoolSettings, err := smis.GetStoragePoolSettings(SRP.InstancePath.InstanceName)
		if err != nil {
			return nil, err
		}
		for _, storagePoolSetting := range storagePoolSettings {
			base_name, _ := GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "EMCSLOBaseName")
			resp_time, _ := GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "EMCApproxAverageResponseTime")
			srp, _ := GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "EMCSRP")
			workload, _ := GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "EMCWorkload")
			elem_name, _ := GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "ElementName")
			inst_id, _ := GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "InstanceID")
			newSLO := SLO_Struct{
				SLO_Name:    base_name.(string),
				respTime:    resp_time.(float64),
				SRP:         srp.(string),
				Workload:    workload.(string),
				ElementName: elem_name.(string),
				InstanceID:  inst_id.(string),
			}
			SLOs = append(SLOs, newSLO)
		}
	}
	return SLOs, nil
}

///////////////////////////////////////////////////////////////
//             ADD Members to a Group                //
///////////////////////////////////////////////////////////////

func (smis *SMIS) AddMembersToGroup(systemInstance *gowbem.InstanceName, group *gowbem.InstancePath, members []gowbem.InstancePath) error {
	controller, err := smis.GetControllerConfigurationService(systemInstance)
	if err != nil {
		return err
	}

	var memberArray gowbem.ValueRefArray
	memberArray.ValueReference = make([]gowbem.ValueReference, len(members))
	for idx := 0; idx < len(members); idx++ {
		memberArray.ValueReference[idx].InstancePath = &members[idx]
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "MaskingGroup", ValueReference: &gowbem.ValueReference{InstancePath: group}})
	params = append(params, gowbem.IParamValue{Name: "Members", ValueRefArray: &memberArray})

	retValue, retParms, err := smis.InvokeMethod(controller, "AddMembers", params)
	if err != nil {
		return err
	}
	if retValue != 0 {
		_, err = smis.WaitForJob(retParms[0].ValueReference.InstancePath, group.InstanceName.ClassName)
	}
	return err
}

///////////////////////////////////////////////////////////////
//          REMOVE Members from a  Group              //
///////////////////////////////////////////////////////////////

func (smis *SMIS) RemoveMembersFromGroup(systemInstance *gowbem.InstanceName, group *gowbem.InstancePath, members []gowbem.InstancePath) error {
	controller, err := smis.GetControllerConfigurationService(systemInstance)
	if err != nil {
		return err
	}

	var memberArray gowbem.ValueRefArray
	memberArray.ValueReference = make([]gowbem.ValueReference, len(members))
	for idx := 0; idx < len(members); idx++ {
		memberArray.ValueReference[idx].InstancePath = &members[idx]
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "MaskingGroup", ValueReference: &gowbem.ValueReference{InstancePath: group}})
	params = append(params, gowbem.IParamValue{Name: "Members", ValueRefArray: &memberArray})

	retValue, retParms, err := smis.InvokeMethod(controller, "RemoveMembers", params)
	if err != nil {
		return err
	}
	if retValue != 0 {
		_, err = smis.WaitForJob(retParms[0].ValueReference.InstancePath, group.InstanceName.ClassName)
	}
	return err
}

///////////////////////////////////////////////////////////////
//          Create Storage Host Initiator                    //
//     idType == 2 for WWN                                   //
//     idType == 5 for IQN                                   //
///////////////////////////////////////////////////////////////

func (smis *SMIS) PostStorageHardwareID(systemInstance *gowbem.InstanceName, storageID string, idType int) (*gowbem.InstancePath, error) {
	management, err := smis.GetStorageHardwareIDManagementService(systemInstance)
	if err != nil {
		return nil, err
	}

	if idType != 2 && idType != 5 {
		return nil, errors.New("Invalid type, must be 2 or 5")
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "StorageID", Value: &gowbem.Value{storageID}})
	params = append(params, gowbem.IParamValue{Name: "IDType", Value: &gowbem.Value{strconv.Itoa(idType)}})

	retValue, retParms, err := smis.InvokeMethod(management, "CreateStorageHardwareID", params)
	if err != nil {
		return nil, err
	}
	if retValue != 0 || len(retParms) == 0 {
		return nil, errors.New("Job failed, rc = " + string(retValue))
	}
	return retParms[0].ValueReference.InstancePath, nil
}

///////////////////////////////////////////////////////////////
//         Delete Initiator/Storage Hardware ID              //
///////////////////////////////////////////////////////////////

func (smis *SMIS) DeleteStorageHardwareID(systemInstance *gowbem.InstanceName, hardwardId *gowbem.InstancePath) error {
	management, err := smis.GetStorageHardwareIDManagementService(systemInstance)
	if err != nil {
		return err
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "HardwareID", ValueReference: &gowbem.ValueReference{InstancePath: hardwardId}})

	retValue, retParms, err := smis.InvokeMethod(management, "DeleteStorageHardwareID", params)
	if err != nil {
		return err
	}
	if retValue != 0 {
		_, err = smis.WaitForJob(retParms[0].ValueReference.InstancePath, hardwardId.InstanceName.ClassName)
	}
	return err
}

///////////////////////////////////////////////////////////////
//                  CREATE a Masking View                    //
///////////////////////////////////////////////////////////////

func (smis *SMIS) PostCreateMaskingView(systemInstance *gowbem.InstanceName, mvName string, sg, ig, pg *gowbem.InstancePath) ([]gowbem.ObjectPath, error) {
	controller, err := smis.GetControllerConfigurationService(systemInstance)
	if err != nil {
		return nil, err
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "ElementName", Value: &gowbem.Value{mvName}})
	params = append(params, gowbem.IParamValue{Name: "DeviceMaskingGroup", ValueReference: &gowbem.ValueReference{InstancePath: sg}})
	params = append(params, gowbem.IParamValue{Name: "InitiatorMaskingGroup", ValueReference: &gowbem.ValueReference{InstancePath: ig}})
	params = append(params, gowbem.IParamValue{Name: "TargetMaskingGroup", ValueReference: &gowbem.ValueReference{InstancePath: pg}})

	_, retValues, err := smis.InvokeMethod(controller, "CreateMaskingView", params)
	if err != nil {
		return nil, err
	}

	idx, _ := FindJobIndex(retValues)
	if idx == -1 {
		return nil, errors.New("Job instance not found")
	}

	return smis.WaitForJob(retValues[idx].ValueReference.InstancePath, "Symm_LunMaskingView")
}

////////////////////////////////////////////////////////////////
//     Struct used to store all Baremetal HBA Information     //
////////////////////////////////////////////////////////////////

type HostAdapter struct {
	HostID string
	WWN    string
}

////////////////////////////////////////////////////////////////
//             GET Baremetal HBA Information                  //
////////////////////////////////////////////////////////////////

func GetBaremetalHBA() (myHosts []HostAdapter, err error) {
	//Works for RedHat 5 and above (including CentOS and SUSE Linux)
	hostDir, _ := ioutil.ReadDir("/sys/class/scsi_host/")

	for _, host := range hostDir {
		if _, err := os.Stat("/sys/class/scsi_host/" + host.Name() + "/device/fc_host/" + host.Name() + "/port_name"); err == nil {
			byteArray, _ := ioutil.ReadFile("/sys/class/scsi_host/" + host.Name() + "/device/fc_host/" + host.Name() + "/port_name")
			newHost := HostAdapter{
				HostID: host.Name(),
				WWN:    string(byteArray),
			}
			myHosts = append(myHosts, newHost)
		}
	}
	return myHosts, nil
}

/////////////////////////////////////////////////////////////////
//                  DELETE an Array Group                      //
// Type Depends on AtType field specified in requesting struct //
/////////////////////////////////////////////////////////////////

func (smis *SMIS) PostDeleteGroup(systemInstance *gowbem.InstanceName, group *gowbem.InstancePath, force bool) error {
	controller, err := smis.GetControllerConfigurationService(systemInstance)
	if err != nil {
		return err
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "MaskingGroup", ValueReference: &gowbem.ValueReference{InstancePath: group}})
	params = append(params, gowbem.IParamValue{Name: "Force", Value: &gowbem.Value{strconv.FormatBool(force)}})

	retValue, retParms, err := smis.InvokeMethod(controller, "DeleteGroup", params)
	if err != nil {
		return err
	}
	if retValue != 0 {
		_, err = smis.WaitForJob(retParms[0].ValueReference.InstancePath, group.InstanceName.ClassName)
	}
	return err
}

/////////////////////////////////////////////////////////////////
//                  DELETE a Volume                            //
/////////////////////////////////////////////////////////////////

func (smis *SMIS) PostDeleteVol(systemInstance *gowbem.InstanceName, volumes []gowbem.InstancePath) error {
	controller, err := smis.GetStorageConfigurationService(systemInstance)
	if err != nil {
		return err
	}

	var volumeArray gowbem.ValueRefArray
	volumeArray.ValueReference = make([]gowbem.ValueReference, len(volumes))
	for idx := 0; idx < len(volumes); idx++ {
		volumeArray.ValueReference[idx].InstancePath = &volumes[idx]
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "TheElements", ValueRefArray: &volumeArray})

	retValue, retParms, err := smis.InvokeMethod(controller, "ReturnElementsToStoragePool", params)
	if err != nil {
		return err
	}
	if retValue != 0 {
		_, err = smis.WaitForJob(retParms[0].ValueReference.InstancePath, "Symm_StorageVolume")
	}
	return err
}

/////////////////////////////////////////////////////////////////
//               DELETE a Masking View                         //
/////////////////////////////////////////////////////////////////

func (smis *SMIS) PostDeleteMaskingView(systemInstance *gowbem.InstanceName, maskingView *gowbem.InstancePath) error {
	controller, err := smis.GetControllerConfigurationService(systemInstance)
	if err != nil {
		return err
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "ProtocolController", ValueReference: &gowbem.ValueReference{InstancePath: maskingView}})

	retValue, retParms, err := smis.InvokeMethod(controller, "DeleteMaskingView", params)
	if err != nil {
		return err
	}
	if retValue != 0 {
		_, err = smis.WaitForJob(retParms[0].ValueReference.InstancePath, "Symm_LunMaskingView")
	}
	return err

}

/////////////////////////////////////////////////////////
//               Structs used for                      //
//        getting Ports logged in, on the VMAX3.       //
/////////////////////////////////////////////////////////

type PortValues struct {
	WWN        string
	PortNumber string
	Director   string
}

///////////////////////////////////////////////////////////////
//           Getting Ports Logged In                         //
///////////////////////////////////////////////////////////////

func (smis *SMIS) PostPortLogins(systemInstance *gowbem.InstanceName, initiator *gowbem.InstancePath) ([]PortValues, error) {

	service, err := smis.GetStorageHardwareIDManagementService(systemInstance)
	if err != nil {
		return nil, err
	}

	var params []gowbem.IParamValue
	params = append(params, gowbem.IParamValue{Name: "HardwareID", ValueReference: &gowbem.ValueReference{InstancePath: initiator}})

	retValue, retParms, err := smis.InvokeMethod(service, "EMCGetTargetEndpoints", params)
	if err != nil {
		return nil, err
	}
	if retValue != 0 {
		return nil, errors.New("Job failed, rc = " + strconv.Itoa(retValue))
	}

	var portValues []PortValues
	for _, ref := range retParms[0].ValueRefArray.ValueReference {
		wwn, _ := GetKeyFromInstanceName(initiator.InstanceName, "InstanceID")
		wwnSplit := strings.Split(wwn.(string), "-+-")

		eSystemName, _ := GetKeyFromInstanceName(ref.InstancePath.InstanceName, "SystemName")
		eSystemNameSplit := strings.Split(eSystemName.(string), "-+-")
		PortAndDirector := strings.Split(eSystemNameSplit[2], "-")
		PV := PortValues{
			WWN:        wwnSplit[1],
			PortNumber: PortAndDirector[0],
			Director:   PortAndDirector[1],
		}
		portValues = append(portValues, PV)
	}
	return portValues, err
}
