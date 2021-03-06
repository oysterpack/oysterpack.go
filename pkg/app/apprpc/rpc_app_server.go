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

package apprpc

import (
	"runtime"

	"bytes"
	"compress/zlib"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"github.com/oysterpack/oysterpack.go/pkg/app/config"
	"github.com/rs/zerolog"
	"zombiezen.com/go/capnproto2"
)

const (
	APP_RPC_SERVICE_ID = app.ServiceID(0xe49214fa20b35ba8)
)

// if the app RPC server fails to start, then this is considered a fatal error, which will terminate the process.
func runRPCAppServer() {
	msg, err := app.Configs.Config(APP_RPC_SERVICE_ID)
	if err != nil {
		app.CONFIG_LOADING_ERR.Log(app.Logger().Panic()).Err(err).Msg("")
	}
	if msg == nil {
		return
	}
	spec, err := config.ReadRootRPCServerSpec(msg)
	if err != nil {
		app.CONFIG_LOADING_ERR.Log(app.Logger().Panic()).Err(err).Msg("")
		return
	}
	serverSpec, err := NewRPCServerSpec(spec)
	if err != nil {
		APP_RPC_START_ERR.Log(Logger().Panic()).Err(err).Msg("")
		return
	}
	startRPCAppServer(serverSpec.ListenerFactory(), serverSpec.TLSConfigProvider(), uint(serverSpec.MaxConns))
}

func startRPCAppServer(listenerFactory ListenerFactory, tlsConfigProvider TLSConfigProvider, maxConns uint) (*RPCService, error) {
	rpcServer := NewService(APP_RPC_SERVICE_ID)
	if err := Services.Register(rpcServer); err != nil {
		APP_RPC_START_ERR.Log(Logger().Panic()).Err(err).Msg("Failed to register app RPC server")
	}

	server := capnprpc.App_ServerToClient(NewAppServer())
	rpcMainInterface := func() (capnp.Client, error) {
		return server.Client, nil
	}

	return StartRPCService(rpcServer, listenerFactory, tlsConfigProvider, rpcMainInterface, maxConns)
}

func NewAppServer() capnprpc.App_Server {
	return rpcAppServer{}
}

type rpcAppServer struct {
	runtimeServer rpcRuntimeServer
	configsServer rpcConfigsServer
}

func (a rpcAppServer) Id(call capnprpc.App_id) error {
	call.Results.SetAppId(uint64(ID()))
	return nil
}

func (a rpcAppServer) ReleaseId(call capnprpc.App_releaseId) error {
	call.Results.SetReleaseId(uint64(Release()))
	return nil
}

func (a rpcAppServer) Instance(call capnprpc.App_instance) error {
	return call.Results.SetInstanceId(string(Instance()))
}

func (a rpcAppServer) StartedOn(call capnprpc.App_startedOn) error {
	call.Results.SetStartedOn(StartedOn().UnixNano())
	return nil
}

func (a rpcAppServer) LogLevel(call capnprpc.App_logLevel) error {
	level, err := ZerologLevel2capnprpcLogLevel(LogLevel())
	if err != nil {
		return err
	}
	call.Results.SetLevel(level)
	return nil
}

func (a rpcAppServer) ServiceIds(call capnprpc.App_serviceIds) error {
	ids, err := Services.ServiceIDs()
	if err != nil {
		return err
	}

	list, err := capnp.NewUInt64List(call.Results.Segment(), int32(len(ids)))
	if err != nil {
		return err
	}
	for i := 0; i < list.Len(); i++ {
		list.Set(i, uint64(ids[i]))
	}
	call.Results.SetServiceIds(list)
	return nil
}

func (a rpcAppServer) Service(call capnprpc.App_service) error {
	service, err := Services.Service(ServiceID(call.Params.Id()))
	if err != nil {
		return err
	}

	return call.Results.SetService(capnprpc.Service_ServerToClient(rpcServiceServer{service.ID()}))
}

func (a rpcAppServer) RpcServiceIds(call capnprpc.App_rpcServiceIds) error {
	ids, err := RPC.ServiceIDs()
	if err != nil {
		return err
	}

	list, err := capnp.NewUInt64List(call.Results.Segment(), int32(len(ids)))
	if err != nil {
		return err
	}
	for i := 0; i < list.Len(); i++ {
		list.Set(i, uint64(ids[i]))
	}
	call.Results.SetServiceIds(list)
	return nil
}

func (a rpcAppServer) RpcService(call capnprpc.App_rpcService) error {
	service, err := RPC.Service(ServiceID(call.Params.Id()))
	if err != nil {
		return err
	}
	return call.Results.SetService(capnprpc.RPCService_ServerToClient(rpcRPCServiceServer{service.ID()}))
}

func (a rpcAppServer) Kill(call capnprpc.App_kill) error {
	app.Kill(nil)
	return nil
}

func (a rpcAppServer) Runtime(call capnprpc.App_runtime) error {
	return call.Results.SetRuntime(capnprpc.Runtime_ServerToClient(a.runtimeServer))
}

func (a rpcAppServer) Configs(call capnprpc.App_configs) error {
	return call.Results.SetConfigs(capnprpc.Configs_ServerToClient(a.configsServer))
}

// CapnprpcLogLevel2zerologLevel capnproc.LogLevel -> zerolog.Level
// error : ErrUnknownLogLevel
func CapnprpcLogLevel2zerologLevel(logLevel capnprpc.LogLevel) (zerolog.Level, error) {
	switch logLevel {
	case capnprpc.LogLevel_debug:
		return zerolog.DebugLevel, nil
	case capnprpc.LogLevel_info:
		return zerolog.InfoLevel, nil
	case capnprpc.LogLevel_warn:
		return zerolog.WarnLevel, nil
	case capnprpc.LogLevel_error:
		return zerolog.ErrorLevel, nil
	default:
		return zerolog.ErrorLevel, ErrUnknownLogLevel
	}
}

// ZerologLevel2capnprpcLogLevel zerolog.Level -> capnprpc.LogLevel
// error : ErrUnknownLogLevel
func ZerologLevel2capnprpcLogLevel(logLevel zerolog.Level) (capnprpc.LogLevel, error) {
	switch logLevel {
	case zerolog.DebugLevel:
		return capnprpc.LogLevel_debug, nil
	case zerolog.InfoLevel:
		return capnprpc.LogLevel_info, nil
	case zerolog.WarnLevel:
		return capnprpc.LogLevel_warn, nil
	case zerolog.ErrorLevel:
		return capnprpc.LogLevel_error, nil
	default:
		return capnprpc.LogLevel_error, ErrUnknownLogLevel
	}
}

type rpcServiceServer struct {
	ServiceID
}

func (a rpcServiceServer) Id(call capnprpc.Service_id) error {
	call.Results.SetServiceId(uint64(a.ServiceID))
	return nil
}

func (a rpcServiceServer) LogLevel(call capnprpc.Service_logLevel) error {
	service, err := Services.Service(a.ServiceID)
	if err != nil {
		return err
	}
	logLevel, err := ZerologLevel2capnprpcLogLevel(service.LogLevel())
	if err != nil {
		return err
	}
	call.Results.SetLevel(logLevel)
	return nil
}

func (a rpcServiceServer) Alive(call capnprpc.Service_alive) error {
	service, err := Services.Service(a.ServiceID)
	if err != nil {
		if err == ErrServiceNotRegistered {
			call.Results.SetAlive(false)
			return nil
		}
		return err
	}
	call.Results.SetAlive(service.Alive())
	return nil
}

type rpcRuntimeServer struct{}

func (a rpcRuntimeServer) GoVersion(call capnprpc.Runtime_goVersion) error {
	return call.Results.SetVersion(runtime.Version())
}

func (a rpcRuntimeServer) NumCPU(call capnprpc.Runtime_numCPU) error {
	call.Results.SetCount(uint32(runtime.NumCPU()))
	return nil
}

func (a rpcRuntimeServer) NumGoroutine(call capnprpc.Runtime_numGoroutine) error {
	call.Results.SetCount(uint32(runtime.NumGoroutine()))
	return nil
}

func (a rpcRuntimeServer) MemStats(call capnprpc.Runtime_memStats) error {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	s := call.Results.Segment()
	stats, err := capnprpc.NewMemStats(s)
	if err != nil {
		return err
	}
	stats.SetAlloc(memStats.Alloc)
	stats.SetTotalAlloc(memStats.TotalAlloc)
	stats.SetSys(memStats.Sys)
	stats.SetLookups(memStats.Lookups)
	stats.SetMallocs(memStats.Mallocs)
	stats.SetFrees(memStats.Frees)

	stats.SetHeapAlloc(memStats.HeapAlloc)
	stats.SetHeapSys(memStats.HeapSys)
	stats.SetHeapIdle(memStats.HeapIdle)
	stats.SetHeapInUse(memStats.HeapInuse)
	stats.SetHeapReleased(memStats.HeapReleased)
	stats.SetHeapObjects(memStats.HeapObjects)

	stats.SetStackInUse(memStats.StackInuse)
	stats.SetStackSys(memStats.StackSys)

	stats.SetMSpanInUse(memStats.MSpanInuse)
	stats.SetMSpanSys(memStats.MSpanSys)
	stats.SetMCacheInUse(memStats.MCacheInuse)
	stats.SetMCacheSys(memStats.MCacheSys)

	stats.SetBuckHashSys(memStats.BuckHashSys)
	stats.SetGCSys(memStats.GCSys)
	stats.SetOtherSys(memStats.OtherSys)

	stats.SetNextGC(memStats.NextGC)
	stats.SetLastGC(memStats.LastGC)

	stats.SetPauseTotalNs(memStats.PauseTotalNs)

	pauses, err := capnp.NewUInt64List(s, int32(len(memStats.PauseNs)))
	if err != nil {
		return err
	}
	for i, pause := range memStats.PauseNs {
		pauses.Set(i, pause)
	}
	if err := stats.SetPauseNs(pauses); err != nil {
		return err
	}

	pauses, err = capnp.NewUInt64List(s, int32(len(memStats.PauseEnd)))
	if err != nil {
		return err
	}
	for i, pause := range memStats.PauseEnd {
		pauses.Set(i, pause)
	}
	if err := stats.SetPauseEnd(pauses); err != nil {
		return err
	}

	stats.SetNumGC(memStats.NumGC)
	stats.SetNumForcedGC(memStats.NumForcedGC)
	stats.SetGCCPUFraction(memStats.GCCPUFraction)

	bySizeList, err := capnprpc.NewMemStats_BySize_List(s, int32(len(memStats.BySize)))
	for i, sizeClassStats := range memStats.BySize {
		memstatsBySize, err := capnprpc.NewMemStats_BySize(s)
		if err != nil {
			return err
		}
		memstatsBySize.SetSize(sizeClassStats.Size)
		memstatsBySize.SetMallocs(sizeClassStats.Mallocs)
		memstatsBySize.SetFrees(sizeClassStats.Frees)
		bySizeList.Set(i, memstatsBySize)
	}
	if err := stats.SetBySize(bySizeList); err != nil {
		return err
	}

	return call.Results.SetStats(stats)
}

func (a rpcRuntimeServer) StackDump(call capnprpc.Runtime_stackDump) error {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	DumpAllStacks(w)
	w.Close()
	return call.Results.SetStackDump(b.Bytes())
}

type rpcRPCServiceServer struct {
	ServiceID
}

func (a rpcRPCServiceServer) Id(call capnprpc.RPCService_id) error {
	call.Results.SetServiceId(uint64(a.ServiceID))
	return nil
}

func (a rpcRPCServiceServer) ListenerAlive(call capnprpc.RPCService_listenerAlive) error {
	service, err := RPC.Service(a.ServiceID)
	if err != nil {
		if err == ErrServiceNotRegistered {
			call.Results.SetListenerAlive(false)
			return nil
		}
		return err
	}
	call.Results.SetListenerAlive(service.Alive())
	return nil
}

func (a rpcRPCServiceServer) ListenerAddress(call capnprpc.RPCService_listenerAddress) error {
	addr, err := a.listenerAddress(call.Results.Segment())
	if err != nil {
		return err
	}
	return call.Results.SetAddress(*addr)
}

func (a rpcRPCServiceServer) listenerAddress(s *capnp.Segment) (*capnprpc.NetworkAddress, error) {
	service, err := RPC.Service(a.ServiceID)
	if err != nil {
		return nil, err
	}

	addr, err := service.ListenerAddress()
	if err != nil {
		return nil, err
	}
	capnpAddr, err := capnprpc.NewNetworkAddress(s)
	if err != nil {
		return nil, err
	}
	capnpAddr.SetNetwork(addr.Network())
	capnpAddr.SetAddress(addr.String())
	return &capnpAddr, nil
}

func (a rpcRPCServiceServer) ActiveConns(call capnprpc.RPCService_activeConns) error {
	service, err := RPC.Service(a.ServiceID)
	if err != nil {
		return err
	}
	call.Results.SetCount(uint32(service.ActiveConns()))
	return nil
}

func (a rpcRPCServiceServer) MaxConns(call capnprpc.RPCService_maxConns) error {
	service, err := RPC.Service(a.ServiceID)
	if err != nil {
		return err
	}
	call.Results.SetCount(uint32(service.MaxConns()))
	return nil
}

type rpcConfigsServer struct{}

func (a rpcConfigsServer) ConfigDir(call capnprpc.Configs_configDir) error {
	call.Results.SetConfigDir(Configs.ConfigDir())
	return nil
}

func (a rpcConfigsServer) ConfigDirExists(call capnprpc.Configs_configDirExists) error {
	call.Results.SetExists(Configs.ConfigDirExists())
	return nil
}

func (a rpcConfigsServer) ServiceIds(call capnprpc.Configs_serviceIds) error {
	serviceIds, err := Configs.ConfigServiceIDs()
	if err != nil {
		return err
	}

	serviceIdsResults, err := capnp.NewUInt64List(call.Results.Segment(), int32(len(serviceIds)))
	for i, id := range serviceIds {
		serviceIdsResults.Set(i, uint64(id))
	}

	call.Results.SetServiceIds(serviceIdsResults)
	return nil
}
