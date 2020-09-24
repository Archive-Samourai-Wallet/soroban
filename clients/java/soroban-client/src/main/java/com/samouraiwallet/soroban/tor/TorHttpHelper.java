package com.samouraiwallet.soroban.tor;

import org.apache.http.config.Registry;
import org.apache.http.config.RegistryBuilder;
import org.apache.http.conn.socket.ConnectionSocketFactory;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.impl.conn.PoolingHttpClientConnectionManager;
import org.apache.http.ssl.SSLContexts;

public final class TorHttpHelper {

    public static CloseableHttpClient createClient()
    {
        Registry<ConnectionSocketFactory> reg = RegistryBuilder.<ConnectionSocketFactory> create()
            .register("http", new TorConnectionSocketFactory())
            .register("https", new TorSSLConnectionSocketFactory(SSLContexts.createSystemDefault()))
            .build();
        PoolingHttpClientConnectionManager cm = new PoolingHttpClientConnectionManager(reg, new TorDnsResolver());
        return HttpClients.custom().setConnectionManager(cm)
            .build();

    }
}
