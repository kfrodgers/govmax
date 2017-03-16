package apiv1

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"GoWBEM/src/gowbem"
)

///////////////////////////////////////////////////////////////
//            GET a list of Storage Arrays                   //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStorageArrays() ([]gowbem.InstanceName, error) {
	return smis.EnumerateInstanceNames("Symm_StorageSystem")
}

func (smis *SMIS) GetStorageInstanceName(sid string) (*gowbem.InstanceName, error) {
	arrays, err := smis.GetStorageArrays()
	if err != nil {
		return nil, err
	}
	for _, array := range arrays {
		name, err := smis.GetKeyFromInstanceName(&array, "Name")
		if err != nil {
			continue
		}
		if strings.HasSuffix(name.(string), sid) {
			return &array, nil
		}
	}
	return nil, errors.New("Array not found")
}

func (smis *SMIS) GetStorageConfigurationService(systemInstanceName *gowbem.InstanceName) (*gowbem.InstanceName, error) {
	configServices, err := smis.EnumerateInstanceNames("EMC_StorageConfigurationService")
	if err != nil {
		return nil, err
	}

	sysName, _ := smis.GetKeyFromInstanceName(systemInstanceName, "Name")
	for _, service := range configServices {
		name, _ := smis.GetKeyFromInstanceName(&service, "SystemName")
		if name.(string) == sysName.(string) {
			return &service, nil

		}
	}
	return nil, errors.New("Service not found")
}

func (smis *SMIS) GetControllerConfigurationService(systemInstanceName *gowbem.InstanceName) (*gowbem.InstanceName, error) {
	controllerServices, err := smis.EnumerateInstanceNames("EMC_ControllerConfigurationService")
	if err != nil {
		return nil, err
	}

	sysName, _ := smis.GetKeyFromInstanceName(systemInstanceName, "Name")
	for _, service := range controllerServices {
		name, _ := smis.GetKeyFromInstanceName(&service, "SystemName")
		if name.(string) == sysName.(string) {
			return &service, nil
		}
	}
	return nil, errors.New("Service not found")
}

func (smis *SMIS) GetStorageHardwareIDManagementService(systemInstanceName *gowbem.InstanceName) (*gowbem.InstanceName, error) {
	managementServices, err := smis.EnumerateInstanceNames("Symm_StorageHardwareIDManagementService")
	if err != nil {
		return nil, err
	}

	sysName, _ := smis.GetKeyFromInstanceName(systemInstanceName, "Name")
	for _, service := range managementServices {
		name, _ := smis.GetKeyFromInstanceName(&service, "SystemName")
		if name.(string) == sysName.(string) {
			return &service, nil
		}
	}
	return nil, errors.New("Service not found")
}

func (smis *SMIS) GetSoftwareIdentity(systemInstanceName *gowbem.InstanceName) (*gowbem.Instance, error) {
	softwareIdents, err := smis.EnumerateInstanceNames("Symm_StorageSystemSoftwareIdentity")
	if err != nil {
		return nil, err
	}

	sysName, _ := smis.GetKeyFromInstanceName(systemInstanceName, "Name")
	for _, swIdent := range softwareIdents {
		name, _ := smis.GetKeyFromInstanceName(&swIdent, "InstanceID")
		if name.(string) == sysName.(string) {
			return smis.GetInstance(&swIdent, false, nil)
		}
	}
	return nil, errors.New("Service not found")
}

func (smis *SMIS) IsArrayV3(systemInstanceName *gowbem.InstanceName) bool {
	swIdent, err := smis.GetSoftwareIdentity(systemInstanceName)
	if err != nil {
		return false
	}
	var major int
	ucode, e := smis.GetPropertyByName(swIdent, "EMCEnginuityFamily")
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
		name, err := smis.GetKeyFromInstanceName(volume.InstancePath.InstanceName, "DeviceID")
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
		panic(err)
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
		elementType, err := smis.GetPropertyByName(adapter, "EMCBSPElementType")
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
			nameProp, _ := smis.GetPropertyByName(volumeInstance, "ElementName")
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
	value, _ := smis.GetPropertyByName(resp, "JobState")
	jobState, _ = strconv.Atoi(value.(string))
	if jobStatus, ok = jobStatusMap[jobState]; !ok {
		jobStatus = "UNKNOWN"
	}
	return resp, jobStatus, err
}

func (smis *SMIS) FindJobIndex(returnParams []gowbem.ParamValue) (int, error) {
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

////////////////////////////////////////////////////////////
//            RESPONSE Struct used for                    //
//          volume creation on the VMAX3.                 //
////////////////////////////////////////////////////////////

type PostVolumesResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_Parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
				I_Size int `json:"i$Size"`
			} `json:"i$parameters"`
			I_ReturnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
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

	idx, _ := smis.FindJobIndex(retValues)
	if idx == -1 {
		return nil, errors.New("Job instance not found")
	}

	return smis.WaitForJob(retValues[idx].ValueReference.InstancePath, "CIM_StorageVolume")
}

//////////////////////////////////////
//   REQUEST Structs used for any   //
//   group creation on the VMAX3.   //
//                                  //
//    Storage Group (SG) - Type 4   //
//     Port Group (PG) - Type 3     //
//   Initiator Group (IG) - Type 2  //
//////////////////////////////////////

type PostGroupReq struct {
	PostGroupRequestContent *PostGroupReqContent `json:"content"`
}

type PostGroupReqContent struct {
	AtType    string `json:"@type"`
	GroupName string `json:"GroupName"`
	Type      string `json:"Type"`
}

////////////////////////////////////////////////////////////
//           RESPONSE Struct  used for any                //
//           group creation on the VMAX3.                 //
//                                                        //
//   Storage Group (SG) - Type SE_DeviceMaskingGroup      //
//      Port Group (PG) - Type SE_TargetMaskingGroup      //
//  Initiator Group (IG) - Type SE_InitiatorMaskingGroup  //
////////////////////////////////////////////////////////////

type PostGroupResp struct {
	Entries []struct {
		Content struct {
			_type        string `json:"@type"`
			I_parameters struct {
				I_MaskingGroup struct {
					_type         string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$MaskingGroup"`
			} `json:"i$parameters"`
			I_returnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

///////////////////////////////////////////////////////////////
//                  CREATE an Array Group                    //
// Type Depends on Type field specified in requesting struct //
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

////////////////////////////////////////////////////////////
//      RESPONSE Struct used for each SRP settings        //
////////////////////////////////////////////////////////////

type GetStoragePoolSettingsResp struct {
	Entries []struct {
		Content struct {
			AtType                           string  `json:"@type"`
			I_Changeable                     bool    `json:"i$Changeable"`
			I_ChangeableType                 int     `json:"i$ChangeableType"`
			I_CompressedElement              bool    `json:"i$CompressedElement"`
			I_CompressionRate                int     `json:"i$CompressionRate"`
			I_DataRedundancyGoal             int     `json:"i$DataRedundancyGoal"`
			I_DataRedundancyMax              int     `json:"i$DataRedundancyMax"`
			I_DataRedundancyMin              int     `json:"i$DataRedundancyMin"`
			I_DeltaReservationGoal           int     `json:"i$DeltaReservationGoal"`
			I_DeltaReservationMax            int     `json:"i$DeltaReservationMax"`
			I_DeltaReservationMin            int     `json:"i$DeltaReservationMin"`
			I_ElementName                    string  `json:"i$ElementName"`
			I_EMCApproxAverageResponseTime   float64 `json:"i$EMCApproxAverageResponseTime"`
			I_EMCDeduplicationRate           int     `json:"i$EMCDeduplicationRate"`
			I_EMCEnableDIF                   int     `json:"i$EMCEnableDIF"`
			I_EMCEnableEFDCache              int     `json:"i$EMCEnableEFDCache"`
			I_EMCFastSetting                 string  `json:"i$EMCFastSetting"`
			I_EMCParticipateInPowerSavings   int     `json:"i$EMCParticipateInPowerSavings"`
			I_EMCPoolCompressionState        int     `json:"i$EMCPoolCompressionState"`
			I_EMCPottedSetting               bool    `json:"i$EMCPottedSetting"`
			I_EMCRaidGroupLUN                bool    `json:"i$EMCRaidGroupLUN"`
			I_EMCRaidLevel                   string  `json:"i$EMCRaidLevel"`
			I_EMCSLO                         string  `json:"i$EMCSLO"`
			I_EMCSLOBaseName                 string  `json:"i$EMCSLOBaseName"`
			I_EMCSLOdescription              string  `json:"i$EMCSLOdescription"`
			I_EMCSRP                         string  `json:"i$EMCSRP"`
			I_EMCStorageSettingType          int     `json:"i$EMCStorageSettingType"`
			I_EMCUniqueID                    string  `json:"i$EMCUniqueID"`
			I_EMCWorkload                    string  `json:"i$EMCWorkload"`
			I_ExtentStripeLength             int     `json:"i$ExtentStripeLength"`
			I_ExtentStripeLengthMax          int     `json:"i$ExtentStripeLengthMax"`
			I_ExtentStripeLengthMin          int     `json:"i$ExtentStripeLengthMin"`
			I_InitialStorageTierMethodology  int     `json:"i$InitialStorageTierMethodology"`
			I_InitialStorageTieringSelection int     `json:"i$InitialStorageTieringSelection"`
			I_InitialSynchronization         int     `json:"i$InitialSynchronization"`
			I_InstanceID                     string  `json:"i$InstanceID"`
			I_NoSinglePointOfFailure         bool    `json:"i$NoSinglePointOfFailure"`
			I_PackageRedundancyGoal          int     `json:"i$PackageRedundancyGoal"`
			I_PackageRedundancyMax           int     `json:"i$PackageRedundancyMax"`
			I_PackageRedundancyMin           int     `json:"i$PackageRedundancyMin"`
			I_SpaceLimit                     int     `json:"i$SpaceLimit"`
			I_StorageExtentInitialUsage      int     `json:"i$StorageExtentInitialUsage"`
			I_StoragePoolInitialUsage        int     `json:"i$StoragePoolInitialUsage"`
			I_ThinProvisionedPoolType        int     `json:"i$ThinProvisionedPoolType"`
			I_UseReplicationBuffer           int     `json:"i$UseReplicationBuffer"`
			Links                            []struct {
				Href string `json:"href"`
				Rel  string `json:"rel"`
			} `json:"links"`
			Xmlns_i string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Gd_etag      string `json:"gd$etag"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

///////////////////////////////////////////////////////////////
//                GET Storage Pool Settings                  //
///////////////////////////////////////////////////////////////
func (smis *SMIS) GetStoragePoolCapabilities(srp_name *gowbem.InstanceName) (*gowbem.InstanceName, error) {
	capabilities, err := smis.EnumerateInstanceNames("Symm_StoragePoolCapabilities")

	name, err := smis.GetKeyFromInstanceName(srp_name, "InstanceID")
	if err != nil {
		return nil, err
	}
	for _, entry := range capabilities {
		key, err := smis.GetKeyFromInstanceName(&entry, "InstanceID")
		if err != nil {
			continue
		}
		if key.(string) == name.(string) {
			return &entry, nil
		}
	}
	return nil, errors.New("Capabilities not found")

}

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
			base_name, _ := smis.GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "EMCSLOBaseName")
			resp_time, _ := smis.GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "EMCApproxAverageResponseTime")
			srp, _ := smis.GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "EMCSRP")
			workload, _ := smis.GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "EMCWorkload")
			elem_name, _ := smis.GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "ElementName")
			inst_id, _ := smis.GetKeyFromInstanceName(storagePoolSetting.InstancePath.InstanceName, "InstanceID")
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

/////////////////////////////////////////////////////////
//               REQUEST Structs used for              //
//   creating a storage hardware ID for an initiator   //
/////////////////////////////////////////////////////////

type PostStorageHardwareIDReq struct {
	PostStorageHardwareIDRequestContent *PostStorageHardwareIDReqContent `json:"content"`
}

type PostStorageHardwareIDReqContent struct {
	AtType    string `json:"@type"`
	IDType    string `json:"IDType"`
	StorageID string `json:"StorageID"`
}

////////////////////////////////////////////////////////////
//            RESPONSE Struct used for                    //
//   creating a storage hardware ID for an initiator      //
////////////////////////////////////////////////////////////

type PostStorageHardwareIDResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_Parameters struct {
				I_HardwareID struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$HardwareID"`
			} `json:"i$parameters"`
			I_ReturnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

/////////////////////////////////////////////////////////
//               REQUEST Struct used for               //
//       adding AND removing an initiator to/from      //
//              a host group on the VMAX3.             //
/////////////////////////////////////////////////////////

type PostInitiatorToHGReq struct {
	PostInitiatorToHGRequestContent *PostInitiatorToHGReqContent `json:"content"`
}

type PostInitiatorToHGReqContent struct {
	AtType                                string                               `json:"@type"`
	PostInitiatorToHGRequestContentMG     *PostInitiatorToHGReqContentMG       `json:"MaskingGroup"`
	PostInitiatorToHGRequestContentMember []*PostInitiatorToHGReqContentMember `json:"Members"`
}

type PostInitiatorToHGReqContentMG struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}

type PostInitiatorToHGReqContentMember struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}

////////////////////////////////////////////////////////////
//                RESPONSE Struct used for                //
//        adding AND removing an initiator to/from        //
//             a host group on the VMAX3.                 //
////////////////////////////////////////////////////////////

type PostInitiatorToHGResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_Parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_ReturnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

///////////////////////////////////////////////////////////////
//             ADD Initiators to a Host Group                //
//                                                           //
//     1 -> Create Storage Hardware ID for the Initiator     //
//     2 -> Add Storage Hardware ID to Initiator Group       //
///////////////////////////////////////////////////////////////

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

/////////////////////////////////////////////////////////
//               REQUEST Structs used for              //
//        creating a masking view on the VMAX3.        //
/////////////////////////////////////////////////////////

type PostCreateMaskingViewReq struct {
	PostCreateMaskingViewRequestContent *PostCreateMaskingViewReqContent `json:"content"`
}

type PostCreateMaskingViewReqContent struct {
	AtType                           string                        `json:"@type"`
	ElementName                      string                        `json:"ElementName"`
	PostInitiatorMaskingGroupRequest *PostInitiatorMaskingGroupReq `json:"InitiatorMaskingGroup"`
	PostTargetMaskingGroupRequest    *PostTargetMaskingGroupReq    `json:"TargetMaskingGroup"`
	PostDeviceMaskingGroupRequest    *PostDeviceMaskingGroupReq    `json:"DeviceMaskingGroup"`
}

type PostInitiatorMaskingGroupReq struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}
type PostTargetMaskingGroupReq struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}
type PostDeviceMaskingGroupReq struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}

////////////////////////////////////////////////////////////
//            RESPONSE Struct used for                    //
//        creating a masking view on the VMAX3.           //
////////////////////////////////////////////////////////////

type PostCreateMaskingViewResp struct {
	Xmlns_gd string `json:"xmlns$gd"`
	Updated  string `json:"updated"`
	ID       string `json:"id"`

	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`

	Entries []struct {
		Updated string `json:"updated"`

		Links []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`

		Content_type string `json:"content-type"`

		Content struct {
			AtType       string `json:"@type"`
			Xmlns_i      string `json:"xmlns$i"`
			I_Parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					Xmlns_e0      string `json:"xmlns$e0"`
					E0_InstanceID string `json:"e0$InstanceID"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_ReturnValue int `json:"i$returnValue"`
		} `json:"content"`
	} `json:"entries"`
}

///////////////////////////////////////////////////////////////
//                  CREATE a Masking View                    //
///////////////////////////////////////////////////////////////

func (smis *SMIS) PostCreateMaskingView(req *PostCreateMaskingViewReq, sid string) (resp *PostCreateMaskingViewResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/CreateMaskingView", req, &resp)
	return resp, err
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

//////////////////////////////////////////////////////////////
//             REQUEST Structs used for any                 //
//          masking view deletion on the VMAX3.             //
//////////////////////////////////////////////////////////////

type DeleteMaskingViewReq struct {
	DeleteMaskingViewRequestContent *DeleteMaskingViewReqContent `json:"content"`
}

type DeleteMaskingViewReqContent struct {
	AtType                            string                         `json:"@type"`
	DeleteMaskingViewRequestContentPC *DeleteMaskingViewReqContentPC `json:"ProtocolController"`
}

type DeleteMaskingViewReqContentPC struct {
	AtType                  string `json:"@type"`
	DeviceID                string `json:"DeviceID"`
	CreationClassName       string `json:"CreationClassName"`
	SystemName              string `json:"SystemName"`
	SystemCreationClassName string `json:"SystemCreationClassName"`
}

////////////////////////////////////////////////////////////
//           RESPONSE Struct used for any                 //
//        masking view deletion on the VMAX3.             //
////////////////////////////////////////////////////////////

type DeleteMaskingViewResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_returnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

/////////////////////////////////////////////////////////////////
//               DELETE a Masking View                         //
/////////////////////////////////////////////////////////////////

func (smis *SMIS) PostDeleteMaskingView(req *DeleteMaskingViewReq, sid string) (resp *DeleteMaskingViewResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/DeleteMaskingView", req, &resp)
	return resp, err
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
		wwn, _ := smis.GetKeyFromInstanceName(initiator.InstanceName, "InstanceID")
		wwnSplit := strings.Split(wwn.(string), "-+-")

		eSystemName, _ := smis.GetKeyFromInstanceName(ref.InstancePath.InstanceName, "SystemName")
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
