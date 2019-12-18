package com.samouraiwallet.soroban.rpc;

import java.io.IOException;
import java.io.InputStream;
import java.net.InetSocketAddress;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

import com.samouraiwallet.soroban.tor.TorHttpHelper;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.protocol.HttpClientContext;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.util.EntityUtils;

public class RpcClient {
    String url;
    HttpClient httpclient;
    HttpClientContext context;

    public RpcClient(String proxy, int port, String url) {
        this.url = url;
        httpclient = TorHttpHelper.createClient();
        context = HttpClientContext.create();

        context.setAttribute("socks.address", new InetSocketAddress(proxy, port));
    }

    public void close() throws IOException {
        ((CloseableHttpClient) httpclient).close();
    }

    private Map<String, Object> call(String method, HashMap<String, Object> params) throws IOException {
        Map<String, Object> result = new HashMap<String, Object>();

        HttpPost request = new HttpPost(url + "/rpc");
        request.setHeader("content-type", "application/json");
        request.setHeader("User-Agent", "HotJava/1.1.2 FCS");

        HashMap<String, Object> hashmap = new HashMap<String, Object>();
        hashmap.put("method", method);
        hashmap.put("jsonrpc", "2.0");
        hashmap.put("id", 1);
        hashmap.put("params", Arrays.asList(params));

        ObjectMapper mapper = new ObjectMapper();
        String json = mapper.writeValueAsString(hashmap);

        StringEntity stringEntity = new StringEntity(json);
        request.setEntity(stringEntity);

        CloseableHttpResponse response = (CloseableHttpResponse) httpclient.execute(request, context);
        try {
            // System.out.println(response.getStatusLine());
            InputStream stream = response.getEntity().getContent();
            JsonNode jsonNode = mapper.readTree(stream);

            mapper = new ObjectMapper();
            result = mapper.convertValue(jsonNode, new TypeReference<Map<String, Object>>() {
            });

            EntityUtils.consume(response.getEntity());
        } finally {
            response.close();
        }
        return result;
    }

    public String[] directoryList(String name) throws IOException {
        HashMap<String, Object> params = new HashMap<String, Object>();
        params.put("Name", name);
        params.put("Entries", new String[0]);

        Map<?, ?> rpc = (Map<?, ?>) call("directory.List", params);
        if (rpc.get("error") != null) {
            return new String[0];
        }
        Map<?, ?> result = (Map<?, ?>) rpc.get("result");

        ArrayList<?> src = (ArrayList<?>) result.get("Entries");
        String[] dest = new String[src.size()];
        System.arraycopy(src.toArray(), 0, dest, 0, src.size());
        return dest;
    }

    public Boolean directoryAdd(String name, String entry, String mode) throws IOException {
        HashMap<String, Object> params = new HashMap<String, Object>();

        params.put("Name", name);
        params.put("Entry", entry);
        params.put("Mode", mode);

        Map<?, ?> rpc = (Map<?, ?>) call("directory.Add", params);
        if (rpc.get("error") != null) {
            return false;
        }
        Map<?, ?> result = (Map<?, ?>) rpc.get("result");
        String status = result.get("Status").toString();
        return status.equals("success");
    }

    public Boolean directoryRemove(String name, String entry) throws IOException {
        HashMap<String, Object> params = new HashMap<String, Object>();
        params.put("Name", name);
        params.put("Entry", entry);

        Map<?, ?> rpc = (Map<?, ?>) call("directory.Remove", params);
        if (rpc.get("error") != null) {
            return false;
        }
        Map<?, ?> result = (Map<?, ?>) rpc.get("result");
        String status = result.get("Status").toString();
        return status.equals("success");
    }
}
