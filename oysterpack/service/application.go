// Copyright (c) 2017 OysterPack, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"fmt"
	"github.com/oysterpack/oysterpack.go/oysterpack/commons"
	"github.com/oysterpack/oysterpack.go/oysterpack/internal/utils"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"
)

// App is a global variable
// It is used to register services application wide, i.e., process wide
var App Application = NewApplicationContext()

// Application is a service interface. It's functionality is defined by the interfaces it composes.
type Application interface {
	ServiceClientRegistry
}

// ApplicationContext.services map entry type
type registeredService struct {
	NewService ServiceClientConstructor
	ServiceClient
}

// ApplicationContext is used to manage a set of services
// ApplicationContext itself is a service, by composition.
//
// Features
// 1. Triggering the application to shutdown, triggers each of its registered services to shutdown.
// 2. Triggers shutdown when SIGTERM and SIGINT signals are received
//
// Use NewApplicationContext() to create a new ApplicationContext instance
type ApplicationContext struct {
	mutex sync.RWMutex
	// once a service is stopped, it will be removed from this map
	services map[commons.InterfaceType]*registeredService

	// ApplicationContext can be managed itself as a service
	service *Service

	// used to keep track of users who are waiting on services
	serviceTicketsMutex sync.RWMutex
	serviceTickets      []*ServiceTicket
}

// represents a ticket issued to a user waiting for a service
type ServiceTicket struct {
	// the type of service the user is waiting for
	commons.InterfaceType
	// used to deliver the ServiceClient to the user
	channel chan ServiceClient

	time.Time
}

// Channel used to wait for the ServiceClient
// if nil is returned, then it means the channel was closed
func (a *ServiceTicket) Channel() <-chan ServiceClient {
	return a.channel
}

// ServiceTicketCounts returns the number of tickets that have been issued per service
func (a *ApplicationContext) ServiceTicketCounts() map[commons.InterfaceType]int {
	a.serviceTicketsMutex.RLock()
	defer a.serviceTicketsMutex.RUnlock()
	counts := make(map[commons.InterfaceType]int)
	for _, ticket := range a.serviceTickets {
		counts[ticket.InterfaceType]++
	}
	return counts
}

// Service returns the Application Service
func (a *ApplicationContext) Service() *Service {
	return a.service
}

// NewApplicationContext returns a new ApplicationContext
func NewApplicationContext() *ApplicationContext {
	app := &ApplicationContext{
		services: make(map[commons.InterfaceType]*registeredService),
	}
	var service Application = app
	serviceInterface, _ := commons.ObjectInterface(&service)
	app.service = NewService(serviceInterface, nil, app.run, app.destroy)
	return app
}

// ServiceByType looks up a service via its service interface.
// If exists is true, then the service was found.
func (a *ApplicationContext) ServiceByType(serviceInterface commons.InterfaceType) ServiceClient {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	if s, exists := a.services[serviceInterface]; exists {
		return s.ServiceClient
	}
	return nil
}

// ServiceByTypeAsync returns channel that will be used to send the ServiceClient, when one is available
func (a *ApplicationContext) ServiceByTypeAsync(serviceInterface commons.InterfaceType) *ServiceTicket {
	ticket := &ServiceTicket{serviceInterface, make(chan ServiceClient, 1), time.Now()}
	serviceClient := a.ServiceByType(serviceInterface)
	if serviceClient != nil {
		ticket.channel <- serviceClient
		close(ticket.channel)
		return ticket
	}

	a.serviceTicketsMutex.Lock()
	a.serviceTickets = append(a.serviceTickets, ticket)
	a.serviceTicketsMutex.Unlock()

	go a.checkServiceTickets()
	return ticket
}

func (a *ApplicationContext) checkServiceTickets() {
	a.serviceTicketsMutex.RLock()
	defer a.serviceTicketsMutex.RUnlock()
	for _, ticket := range a.serviceTickets {
		serviceClient := a.ServiceByType(ticket.InterfaceType)
		if serviceClient != nil {
			go func(ticket *ServiceTicket) {
				defer utils.IgnorePanic()
				ticket.channel <- serviceClient
				close(ticket.channel)
			}(ticket)
			go a.deleteServiceTicket(ticket)
		}
	}
}

func (a *ApplicationContext) deleteServiceTicket(ticket *ServiceTicket) {
	a.serviceTicketsMutex.Lock()
	defer a.serviceTicketsMutex.Unlock()
	for i, elem := range a.serviceTickets {
		if elem == ticket {
			a.serviceTickets[i] = a.serviceTickets[len(a.serviceTickets)-1] // Replace it with the last one.
			a.serviceTickets = a.serviceTickets[:len(a.serviceTickets)-1]
			return
		}
	}
}

// ServiceByKey looks up a service by ServiceKey and returns the registered ServiceClient.
// If the service is not found, then nil is returned.
func (a *ApplicationContext) ServiceByKey(serviceKey ServiceKey) ServiceClient {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for _, s := range a.services {
		if InterfaceTypeToServiceKey(s.Service().serviceInterface) == serviceKey {
			return s.ServiceClient
		}
	}
	return nil
}

// Services returns all registered services
func (a *ApplicationContext) Services() []ServiceClient {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	var services []ServiceClient
	for _, s := range a.services {
		services = append(services, s.ServiceClient)
	}
	return services
}

// ServiceInterfaces returns all service interfaces for all registered services
func (a *ApplicationContext) ServiceInterfaces() []commons.InterfaceType {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	interfaces := []commons.InterfaceType{}
	for k := range a.services {
		interfaces = append(interfaces, k)
	}
	return interfaces
}

// ServiceCount returns the number of registered services
func (a *ApplicationContext) ServiceCount() int {
	return len(a.services)
}

// ServiceKeys returns ServiceKey(s) for all registered services
func (a *ApplicationContext) ServiceKeys() []ServiceKey {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	interfaces := []ServiceKey{}
	for _, v := range a.ServiceInterfaces() {
		interfaces = append(interfaces, InterfaceTypeToServiceKey(v))
	}
	return interfaces
}

// RegisterService will register the service and start it, if it is not already registered.
// Returns the new registered service or nil if a service with the same interface was already registered.
// A panic occurs if the ServiceClient type is not assignable to the Service.
func (a *ApplicationContext) RegisterService(newService ServiceClientConstructor) ServiceClient {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	service := newService()
	if !reflect.TypeOf(service).AssignableTo(service.Service().serviceInterface) {
		panic(fmt.Sprintf("%T is not assignable to %v", reflect.TypeOf(service), service.Service().serviceInterface))
	}
	if _, exists := a.services[service.Service().serviceInterface]; !exists {
		a.services[service.Service().serviceInterface] = &registeredService{NewService: newService, ServiceClient: service}
		service.Service().StartAsync()
		go a.checkServiceTickets()
		return service
	}
	return nil
}

// UnRegisterService unregisters the specified service.
// The service is simply unregistered, i.e., it is not stopped.
func (a *ApplicationContext) UnRegisterService(service ServiceClient) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if _, exists := a.services[service.Service().serviceInterface]; exists {
		delete(a.services, service.Service().serviceInterface)
		return true
	}
	return false
}

func (a *ApplicationContext) run(ctx *RunContext) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-sigs:
			return nil
		case <-ctx.StopTrigger():
			return nil
		}
	}
}

func (a *ApplicationContext) destroy(ctx *Context) error {
	for _, v := range a.services {
		v.Service().StopAsyc()
	}
	for _, v := range a.services {
		for {
			v.Service().AwaitStopped(5 * time.Second)
			if v.Service().State().Stopped() {
				break
			}
			v.Service().Logger.Warn().Msg("Waiting for service to stop")
		}
	}
	return nil
}
