package com.samouraiwallet.soroban.client;

import java.io.IOException;

import com.samouraiwallet.soroban.rpc.RpcClient;


public final class SorobanClient {

    public static void main(String[] args) throws IOException {
        RpcClient rpc = new RpcClient("127.0.0.1", 9050, "http://sorzvujomsfbibm7yo3k52f3t2bl6roliijnm7qql43bcoe2kxwhbcyd.onion");
        try{
            // add hello key to directory
            if (!rpc.directoryAdd("hello_java", "Hello, from Java!", "short")) {
                System.err.println("Failed to add entry to directory");
                return;
            }
            System.out.println("Entry added to directoy.");

            // retieve hello key from directory
            String[] entries = rpc.directoryList("hello_java");
            if (entries.length == 0 ){
                System.err.println("Entry not found from directory");
                return;
            }
            System.out.println("Entry found in directory: " + entries[0]);

            // remove hello key from directory
            if (!rpc.directoryRemove("hello_java", "Hello, from Java!")){
                System.err.println("Failed to remove entry from directory");
                return;
            }
            System.out.println("Entry removed from directoy.");

            // retieve hello key from directory
            entries = rpc.directoryList("hello_java");
            if (entries.length != 0 ){
                System.err.println("Found deleted entry from directory");
                return;
            }
            System.out.println("Finished with no entries in directory.");
        } finally {
            rpc.close();
        }
    }
}
    