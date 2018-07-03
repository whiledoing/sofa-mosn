/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"gitlab.alipay-inc.com/afe/mosn/pkg/api/v2"
	"gitlab.alipay-inc.com/afe/mosn/pkg/log"
	"gitlab.alipay-inc.com/afe/mosn/pkg/network"
	"gitlab.alipay-inc.com/afe/mosn/pkg/network/buffer"
	"gitlab.alipay-inc.com/afe/mosn/pkg/protocol"
	_ "gitlab.alipay-inc.com/afe/mosn/pkg/protocol/sofarpc/codec"
	_ "gitlab.alipay-inc.com/afe/mosn/pkg/router/basic"
	"gitlab.alipay-inc.com/afe/mosn/pkg/server"
	"gitlab.alipay-inc.com/afe/mosn/pkg/server/config/proxy"
	_ "gitlab.alipay-inc.com/afe/mosn/pkg/stream/http2"
	"gitlab.alipay-inc.com/afe/mosn/pkg/types"
	"gitlab.alipay-inc.com/afe/mosn/pkg/upstream/cluster"
	"golang.org/x/net/http2"
)

const (
	RealServerAddr = "127.0.0.1:8088"
	MeshServerAddr = "127.0.0.1:2045"
	TestCluster    = "tstCluster"
	TestListener   = "tstListener"
)

var trReqBytes = []byte{0x0d, 0x00, 0x04, 0x02, 0x00, 0x00, 0x00, 0x00, 0x7f, 0x2c, 0x00, 0x00, 0x01, 0x85, 0x4f, 0xba, 0x63, 0x6f, 0x6d, 0x2e, 0x74, 0x61, 0x6f, 0x62, 0x61, 0x6f, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x69, 0x6d, 0x70, 0x6c, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x91, 0x03, 0x63, 0x74, 0x78, 0x6f, 0x90, 0x4f, 0xc8, 0x39, 0x63, 0x6f, 0x6d, 0x2e, 0x74, 0x61, 0x6f, 0x62, 0x61, 0x6f, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x69, 0x6d, 0x70, 0x6c, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x24, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x92, 0x02, 0x69, 0x64, 0x06, 0x74, 0x68, 0x69, 0x73, 0x24, 0x30, 0x6f, 0x91, 0xe1, 0x4a, 0x00, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x6c, 0x69, 0x70, 0x61, 0x79, 0x2e, 0x73, 0x6f, 0x66, 0x61, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x53, 0x6f, 0x66, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4f, 0xbc, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x6c, 0x69, 0x70, 0x61, 0x79, 0x2e, 0x73, 0x6f, 0x66, 0x61, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x53, 0x6f, 0x66, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x95, 0x0d, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x41, 0x70, 0x70, 0x4e, 0x61, 0x6d, 0x65, 0x0a, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x17, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x55, 0x6e, 0x69, 0x71, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x0c, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x73, 0x0d, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x41, 0x72, 0x67, 0x53, 0x69, 0x67, 0x73, 0x6f, 0x90, 0x01, 0x2d, 0x03, 0x73, 0x61, 0x79, 0x53, 0x00, 0x21, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x6c, 0x69, 0x70, 0x61, 0x79, 0x2e, 0x64, 0x65, 0x6d, 0x6f, 0x2e, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x3a, 0x31, 0x2e, 0x30, 0x4d, 0x11, 0x72, 0x70, 0x63, 0x5f, 0x74, 0x72, 0x61, 0x63, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x4d, 0x09, 0x73, 0x6f, 0x66, 0x61, 0x52, 0x70, 0x63, 0x49, 0x64, 0x01, 0x30, 0x07, 0x45, 0x6c, 0x61, 0x73, 0x74, 0x69, 0x63, 0x01, 0x46, 0x0b, 0x73, 0x79, 0x73, 0x50, 0x65, 0x6e, 0x41, 0x74, 0x74, 0x72, 0x73, 0x00, 0x09, 0x7a, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x55, 0x49, 0x44, 0x00, 0x10, 0x7a, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5a, 0x6f, 0x6e, 0x65, 0x00, 0x0c, 0x73, 0x6f, 0x66, 0x61, 0x43, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x49, 0x70, 0x0d, 0x31, 0x30, 0x2e, 0x31, 0x35, 0x2e, 0x32, 0x34, 0x36, 0x2e, 0x31, 0x38, 0x32, 0x0b, 0x73, 0x6f, 0x66, 0x61, 0x54, 0x72, 0x61, 0x63, 0x65, 0x49, 0x64, 0x1e, 0x30, 0x61, 0x30, 0x66, 0x66, 0x36, 0x62, 0x36, 0x31, 0x35, 0x32, 0x32, 0x36, 0x35, 0x32, 0x32, 0x31, 0x35, 0x37, 0x39, 0x35, 0x31, 0x30, 0x30, 0x31, 0x33, 0x35, 0x37, 0x38, 0x34, 0x0c, 0x73, 0x6f, 0x66, 0x61, 0x50, 0x65, 0x6e, 0x41, 0x74, 0x74, 0x72, 0x73, 0x00, 0x0d, 0x73, 0x6f, 0x66, 0x61, 0x43, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x41, 0x70, 0x70, 0x04, 0x74, 0x65, 0x73, 0x74, 0x0d, 0x7a, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x04, 0x33, 0x30, 0x30, 0x30, 0x7a, 0x7a, 0x56, 0x74, 0x00, 0x07, 0x5b, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x6e, 0x00, 0x7a}

func main() {
	go func() {
		// pprof server
		http.ListenAndServe("0.0.0.0:9099", nil)
	}()
	
	log.InitDefaultLogger("", log.DEBUG)

	stopChan := make(chan bool)
	meshReadyChan := make(chan bool)

	go func() {
		// upstream
		server := &http.Server{
			Addr:         ":8080",
			Handler:      &serverHandler{},
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}
		s2 := &http2.Server{
			IdleTimeout: 1 * time.Minute,
		}

		http2.ConfigureServer(server, s2)
		l, _ := net.Listen("tcp", RealServerAddr)
		defer l.Close()

		for {
			rwc, err := l.Accept()
			if err != nil {
				fmt.Println("accept err:", err)
				continue
			}
			go s2.ServeConn(rwc, &http2.ServeConnOpts{BaseConfig: server})
		}
	}()

	select {
	case <-time.After(2 * time.Second):
	}

	go func() {
		//  mesh
		cmf := &clusterManagerFilterRPC{}

		cm := cluster.NewClusterManager(nil, nil, nil, false, false)

		//RPC
		srv := server.NewServer(&server.Config{}, cmf, cm)

		srv.AddListener(rpcProxyListener(), &proxy.GenericProxyFilterConfigFactory{
			Proxy: genericProxyConfig(),
		}, nil)
		cmf.cccb.UpdateClusterConfig(clustersrpc())
		cmf.chcb.UpdateClusterHost(TestCluster, 0, rpchosts())

		meshReadyChan <- true

		srv.Start() //开启连接

		select {
		case <-stopChan:
			srv.Close()
		}
	}()

	go func() {
		select {
		case <-meshReadyChan:
			// client
			remoteAddr, _ := net.ResolveTCPAddr("tcp", MeshServerAddr)
			cc := network.NewClientConnection(nil, nil, remoteAddr, stopChan, log.DefaultLogger)
			cc.AddConnectionEventListener(&rpclientConnCallbacks{ //ADD  connection callback
				cc: cc,
			})
			cc.Connect(true)
			cc.FilterManager().AddReadFilter(&rpcclientConnReadFilter{})

			select {
			case <-stopChan:
				cc.Close(types.NoFlush, types.LocalClose)
			}
		}
	}()

	select {
	case <-time.After(time.Second * 120):
		stopChan <- true
		fmt.Println("[MAIN]closing..")
	}
}

type serverHandler struct{}

func (sh *serverHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ShowRequestInfoHandler(w, req)
}

func ShowRequestInfoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[UPSTREAM]receive request %s", r.URL)
	fmt.Println()

	w.Header().Set("Content-Type", "text/plain")

	for k, _ := range r.Header {
		w.Header().Set(k, r.Header.Get(k))
	}

	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "Protocol: %s\n", r.Proto)
	fmt.Fprintf(w, "Host: %s\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "RequestURI: %q\n", r.RequestURI)
	fmt.Fprintf(w, "URL: %#v\n", r.URL)
	fmt.Fprintf(w, "Body.ContentLength: %d (-1 means unknown)\n", r.ContentLength)
	fmt.Fprintf(w, "Close: %v (relevant for HTTP/1 only)\n", r.Close)
	fmt.Fprintf(w, "TLS: %#v\n", r.TLS)
	fmt.Fprintf(w, "\nHeaders:\n")

	r.Header.Write(w)
}

func genericProxyConfig() *v2.Proxy {
	proxyConfig := &v2.Proxy{
		DownstreamProtocol: string(protocol.SofaRpc),
		UpstreamProtocol:   string(protocol.Http2),
	}

	header := v2.HeaderMatcher{
		Name:  "service",
		Value: "com.alipay.rpc.common.service.facade.SampleService:1.0",
	}

	var envoyvalue = map[string]interface{}{"stage": "pre-release", "version": "1.1", "label": "gray"}

	var value = map[string]interface{}{"mosn.lb": envoyvalue}

	routerV2 := v2.Router{
		Match: v2.RouterMatch{
			Headers: []v2.HeaderMatcher{header},
		},

		Route: v2.RouteAction{
			ClusterName: TestCluster,
			MetadataMatch: v2.Metadata{
				"filter_metadata": value,
			},
		},
	}

	proxyConfig.VirtualHosts = append(proxyConfig.VirtualHosts, &v2.VirtualHost{
		Name:    "testSofaRoute",
		Domains: []string{"*"},
		Routers: []v2.Router{routerV2},
	})

	return proxyConfig
}

//func genericProxyConfig() *v2.Proxy {
//	proxyConfig := &v2.Proxy{
//		DownstreamProtocol: string(protocol.SofaRpc),
//		UpstreamProtocol:   string(protocol.Http2),
//	}
//
//	proxyConfig.BasicRoutes = append(proxyConfig.BasicRoutes, &v2.BasicServiceRoute{
//		Name:    "tstSofRpcRouter",
//		Service: ".*",
//		Cluster: TestCluster,
//	})
//
//	return proxyConfig
//}

func rpcProxyListener() *v2.ListenerConfig {
	addr, _ := net.ResolveTCPAddr("tcp", MeshServerAddr)

	return &v2.ListenerConfig{
		Name:                    TestListener,
		Addr:                    addr,
		BindToPort:              true,
		PerConnBufferLimitBytes: 1024 * 32,
		LogPath:                 "",
		LogLevel:                uint8(log.DEBUG),
	}
}

func rpchosts() []v2.Host {
	var hosts []v2.Host

	hosts = append(hosts, v2.Host{
		Address: RealServerAddr,
		Weight:  100,
		MetaData: map[string]interface{}{
			"stage":   "pre-release",
			"version": "1.1",
			"label":   "gray",
		},
	})

	return hosts
}

//func rpchosts() []v2.Host {
//	var hosts []v2.Host
//
//	hosts = append(hosts, v2.Host{
//		Address: RealServerAddr,
//		Weight:  100,
//	})
//
//	return hosts
//}

func clustersrpc() []v2.Cluster {
	var configs []v2.Cluster
	var lbsubsetconfig = v2.LBSubsetConfig{

		FallBackPolicy: 2,
		DefaultSubset: map[string]string{
			"stage":   "pre-release",
			"version": "1.1",
			"label":   "gray",
		},
		SubsetSelectors: [][]string{{"stage", "type"},
			{"stage", "label", "version"},
			{"version"}},
	}

	/*
			"stage":   "pre-release",
		"version": "1.1",
		"label":   "gray",*/

	configs = append(configs, v2.Cluster{
		Name:                 TestCluster,
		ClusterType:          v2.SIMPLE_CLUSTER,
		LbType:               v2.LB_RANDOM,
		MaxRequestPerConn:    1024,
		ConnBufferLimitBytes: 32 * 1024,
		CirBreThresholds:     v2.CircuitBreakers{},
		LBSubSetConfig:       lbsubsetconfig,
	})

	return configs
}

type clusterManagerFilterRPC struct {
	cccb types.ClusterConfigFactoryCb
	chcb types.ClusterHostFactoryCb
}

func (cmf *clusterManagerFilterRPC) OnCreated(cccb types.ClusterConfigFactoryCb, chcb types.ClusterHostFactoryCb) {
	cmf.cccb = cccb
	cmf.chcb = chcb
}

//func clustersrpc() []v2.Cluster {
//	var configs []v2.Cluster
//	configs = append(configs, v2.Cluster{
//		Name:                 TestCluster,
//		ClusterType:          v2.SIMPLE_CLUSTER,
//		LbType:               v2.LB_RANDOM,
//		MaxRequestPerConn:    1024,
//		ConnBufferLimitBytes: 32 * 1024,
//	})
//
//	return configs
//}

type rpclientConnCallbacks struct {
	cc types.Connection
}

func (ccc *rpclientConnCallbacks) OnEvent(event types.ConnectionEvent) {
	fmt.Printf("[CLIENT]connection event %s", string(event))
	fmt.Println()

	switch event {
	case types.Connected:
		time.Sleep(3 * time.Second)

		fmt.Println("[CLIENT]write 'TR Protoocl' to remote server")

		boltV2PostData := buffer.NewIoBufferBytes(trReqBytes)

		ccc.cc.Write(boltV2PostData)
	}
}

func (ccc *rpclientConnCallbacks) OnAboveWriteBufferHighWatermark() {}

func (ccc *rpclientConnCallbacks) OnBelowWriteBufferLowWatermark() {}

type rpcclientConnReadFilter struct {
}

func (ccrf *rpcclientConnReadFilter) OnData(buffer types.IoBuffer) types.FilterStatus {
	fmt.Println()
	fmt.Println("[CLIENT]Receive data:")
	fmt.Printf("%s", buffer.String())
	buffer.Reset()

	return types.Continue
}

func (ccrf *rpcclientConnReadFilter) OnNewConnection() types.FilterStatus {
	return types.Continue
}

func (ccrf *rpcclientConnReadFilter) InitializeReadFilterCallbacks(cb types.ReadFilterCallbacks) {}
