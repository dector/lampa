package space.dector.lampa.server;

import exported.ServerConfig;

public final class ServerRuntimeConfig {
    public static final String PROXY_HOST = "localhost";
    public static final int PROXY_PORT = 8080;
    public static final String CONTROL_HOST = "localhost";
    public static final int CONTROL_PORT = 8081;

    public static final String PROXY_ROUTE_PATH = "/";
    public static final String CONTROL_ROUTE_PATH = "/";

    private ServerRuntimeConfig() {
    }

    public static String proxyEndpoint() {
        return buildListenAddress(PROXY_HOST, PROXY_PORT);
    }

    public static String controlEndpoint() {
        return buildListenAddress(CONTROL_HOST, CONTROL_PORT);
    }

    public static String buildListenAddress(String host, int port) {
        return host + ":" + port;
    }

    public static ServerConfig toServerConfig() {
        ServerConfig config = new ServerConfig();
        config.setProxyListenAddress(proxyEndpoint());
        config.setProxyRoutePath(PROXY_ROUTE_PATH);
        config.setControlListenAddress(controlEndpoint());
        config.setControlRoutePath(CONTROL_ROUTE_PATH);
        return config;
    }
}
