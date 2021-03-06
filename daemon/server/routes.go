// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package server

import (
	"net/http"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type routes []route

func (r *Router) initBackendRoutes() {
	r.routes = routes{
		route{
			"Ping", "GET", "/ping", r.ping,
		},
		route{
			"GlobalStatus", "GET", "/healthz", r.globalStatus,
		},
		route{
			"Update", "POST", "/update", r.update,
		},
		// adaptor handlers:
		route{
			"AdaptorAttach", "POST", "/adaptor", r.adaptorAttach,
		},
		route{
			"AdaptorDetach", "DELETE", "/adaptor/{machineID}/{adaptorType}/{adaptorID}", r.adaptorDetach,
		},
		// driver handlers:
		route{
			"DriverStart", "POST", "/driver", r.driverStart,
		},
		route{
			"DriverStop", "DELETE", "/driver/{machineID}/{driverType}/{driverID}", r.driverStop,
		},
		// controller handlers:
		route{
			"ControllerStart", "POST", "/controller", r.controllerStart,
		},
		route{
			"ControllerStop", "DELETE", "/controller/{machineID}/{controllerID}", r.controllerStop,
		},
		// machine handlers:
		route{
			"MachineCreate", "POST", "/machine/{machineID}", r.machineCreate,
		},
		route{
			"MachineDelete", "DELETE", "/machine/{machineID}", r.machineDelete,
		},
		route{
			"MachineGet", "GET", "/machine/{machineID}", r.machineGet,
		},
		route{
			"MachineUpdate", "POST", "/machine/update/{machineID}", r.machineUpdate,
		},
		route{
			"MachinesGet", "GET", "/machines", r.machinesGet,
		},
	}
}
