using Go = import "/go.capnp";
@0x99f3cbccce65aee8;
$Go.package("config");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/config");

struct RPCServerSpec @0xfc13c8456771ca68 {
    rpcServiceSpec  @0 :RPCServiceSpec;
    serverCert      @1 :X509KeyPair;
    caCert          @2 :Data;
    maxConns        @3 :UInt32;
}

struct RPCClientSpec @0xbec6688394d29776 {
    rpcServiceSpec  @0 :RPCServiceSpec;
    clientCert      @1 :X509KeyPair;
    caCert          @2 :Data $Go.doc("PEM file format");
}

struct RPCServiceSpec @0xb6e32df5c504ebf2 {
    domainID        @0 :UInt64;
    appId           @1 :UInt64;
    serviceId       @2 :UInt64;

    port            @3 :UInt16;
}

struct X509KeyPair @0xf4dd73213f6e70a6 {
    key     @0 :Data $Go.doc("PEM file format");
    cert    @1 :Data $Go.doc("PEM file format");
}