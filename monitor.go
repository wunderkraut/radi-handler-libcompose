package libcompose

import (
	"context"
	"errors"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"

	api_operation "github.com/wunderkraut/radi-api/operation"
	api_property "github.com/wunderkraut/radi-api/property"
	api_result "github.com/wunderkraut/radi-api/result"
	api_usage "github.com/wunderkraut/radi-api/usage"

	api_monitor "github.com/wunderkraut/radi-api/operation/monitor"
)

/**
 * Monitoring operations provided by the libcompose
 * handler.
 */

const (
	OPERATION_ID_COMPOSE_MONITOR_LOGS = "monitor.libcompose.logs"
	OPERATION_ID_COMPOSE_MONITOR_PS   = "monitor.libcompose.ps"
)

// An operations which streams the container logs from libcompose
type LibcomposeMonitorLogsOperation struct {
	api_monitor.BaseMonitorLogsOperation
	BaseLibcomposeNameFilesOperation
}

// Use a different Id() than the parent
func (logs *LibcomposeMonitorLogsOperation) Id() string {
	return OPERATION_ID_COMPOSE_MONITOR_LOGS
}

// Validate
func (logs *LibcomposeMonitorLogsOperation) Usage() api_usage.Usage {
	return api_operation.Usage_External()
}

func (logs *LibcomposeMonitorLogsOperation) Validate() api_result.Result {
	return api_result.MakeSuccessfulResult()
}

// Provide static properties for the operation
func (logs *LibcomposeMonitorLogsOperation) Properties() api_property.Properties {
	props := api_property.New_SimplePropertiesEmpty()

	props.Merge(logs.BaseLibcomposeNameFilesOperation.Properties())
	props.Add(api_property.Property(&LibcomposeDetachProperty{}))

	return props.Properties()
}

// Execute the libCompose monitor logs operation
func (logs *LibcomposeMonitorLogsOperation) Exec(props api_property.Properties) api_result.Result {
	res := api_result.New_StandardResult()

	// pass all confs to make a project
	project, _ := MakeComposeProject(props)

	// some confs we will use locally

	var netContext context.Context
	// net context
	if netContextProp, found := props.Get(OPERATION_PROPERTY_LIBCOMPOSE_CONTEXT); found {
		netContext = netContextProp.Get().(context.Context)
	} else {
		res.MarkFailed()
		res.AddError(errors.New("Libcompose up operation is missing the context property"))
	}

	var follow bool
	// follow conf
	if followProp, found := props.Get(OPERATION_PROPERTY_LIBCOMPOSE_DETACH); found {
		follow = !followProp.Get().(bool)
	} else {
		res.AddError(errors.New("Libcompose logs operation is missing the detach property"))
		res.MarkFailed()
	}

	// output handling test
	if outputProp, found := props.Get(OPERATION_PROPERTY_LIBCOMPOSE_OUTPUT); found {
		outputProp.Set(io.Writer(os.Stdout))
	}

	if res.Success() {
		if err := project.APIProject.Log(netContext, follow); err != nil {
			res.MarkFailed()
			res.AddError(err)
			res.AddError(errors.New("Could not attach to the project for logs"))
		}
	}

	res.MarkFinished()

	return res.Result()
}

// LibCompose based ps orchestrate operation
type LibcomposeOrchestratePsOperation struct {
	BaseLibcomposeNameFilesOperation
}

// Label the operation
func (ps *LibcomposeOrchestratePsOperation) Id() string {
	return OPERATION_ID_COMPOSE_MONITOR_PS
}

// Label the operation
func (ps *LibcomposeOrchestratePsOperation) Label() string {
	return "List containers"
}

// Description for the operation
func (ps *LibcomposeOrchestratePsOperation) Description() string {
	return "List all containers used by libCompose."
}

// Man page for the operation
func (ps *LibcomposeOrchestratePsOperation) Help() string {
	return ""
}

// Is this an internal API operation
func (ps *LibcomposeOrchestratePsOperation) Usage() api_usage.Usage {
	return api_operation.Usage_External()
}

// Validate the libCompose Orchestrate Ps operation
func (ps *LibcomposeOrchestratePsOperation) Validate() api_result.Result {
	return api_result.MakeSuccessfulResult()
}

// Provide static properties for the operation
func (ps *LibcomposeOrchestratePsOperation) Properties() api_property.Properties {
	return ps.BaseLibcomposeNameFilesOperation.Properties()
}

// Execute the libCompose Orchestrate Ps operation
func (ps *LibcomposeOrchestratePsOperation) Exec(props api_property.Properties) api_result.Result {
	res := api_result.New_StandardResult()

	// pass all props to make a project
	project, _ := MakeComposeProject(props)

	// some props we will use locally
	var netContext context.Context

	// net context
	if netContextProp, found := props.Get(OPERATION_PROPERTY_LIBCOMPOSE_CONTEXT); found {
		netContext = netContextProp.Get().(context.Context)
	} else {
		res.MarkFinished()
		res.AddError(errors.New("Libcompose ps operation is missing the context property"))
	}

	if res.Success() {
		if infoset, err := project.APIProject.Ps(netContext); err == nil {
			if len(infoset) == 0 {
				log.Info("No running containers found.")
			} else {
				for index, info := range infoset {
					id, _ := info["Id"]
					name, _ := info["Name"]
					state, _ := info["State"]
					log.WithFields(log.Fields{"index": index, "id": id, "name": name, "state": state, "info": info}).Info("Compose info")
				}
			}
		} else {
			res.MarkFailed()
			res.AddError(err)
		}
	}

	return res.Result()
}
